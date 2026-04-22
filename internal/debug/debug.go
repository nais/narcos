package debug

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultImage = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"

type Options struct {
	PodName         string
	Namespace       string
	Image           string
	ExtraCaps       []string
	TargetContainer string
	KubeContext     string
}

func Run(ctx context.Context, opts Options) error {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl not found in PATH — required for interactive session")
	}

	if opts.Image == "" {
		opts.Image = defaultImage
	}

	typedClient, dynClient, _, err := NewClients(opts.KubeContext)
	if err != nil {
		return fmt.Errorf("creating K8s clients: %w", err)
	}

	pod, err := typedClient.CoreV1().Pods(opts.Namespace).Get(ctx, opts.PodName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("pod %q not found in namespace %q", opts.PodName, opts.Namespace)
		}
		if k8serrors.IsForbidden(err) {
			return fmt.Errorf("access denied for pod %q in namespace %q — check your RBAC permissions: %w", opts.PodName, opts.Namespace, err)
		}
		return fmt.Errorf("getting pod %q in namespace %q: %w", opts.PodName, opts.Namespace, err)
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return fmt.Errorf("pod is not running (phase: %s)", pod.Status.Phase)
	}

	targetContainer := opts.TargetContainer
	if targetContainer == "" {
		targetContainer = FirstContainerName(pod)
	}

	caps := []corev1.Capability{"NET_RAW", "SYS_PTRACE"}
	for _, c := range opts.ExtraCaps {
		caps = append(caps, corev1.Capability(c))
	}

	containerName := GenerateContainerName(pod)
	ec := BuildEphemeralContainer(containerName, opts.Image, targetContainer, caps)

	fmt.Fprintf(os.Stderr, "Creating debug container %q in pod %s/%s...\n", containerName, opts.Namespace, opts.PodName)

	const maxRetries = 5
	var kyvernoCRDChecked bool
	for attempt := range maxRetries + 1 {
		_, err = PatchEphemeralContainer(ctx, typedClient, pod, ec)
		if err == nil {
			break
		}

		violations := ParseKyvernoDenial(err)
		if len(violations) == 0 {
			return fmt.Errorf("creating ephemeral container: %w", err)
		}

		if attempt == maxRetries {
			break
		}

		if !kyvernoCRDChecked {
			if !KyvernoCRDExists(ctx, dynClient) {
				return fmt.Errorf("kyverno denied the request but PolicyException CRD not found — cannot create exception: %w", err)
			}
			kyvernoCRDChecked = true
		}

		pe := BuildPolicyException(pod, violations)
		if createErr := CreatePolicyException(ctx, dynClient, pe); createErr != nil {
			return fmt.Errorf("creating PolicyException: %w", createErr)
		}
		fmt.Fprintf(os.Stderr, "Created PolicyException for %d policy violation(s), retrying (%d/%d)...\n", len(violations), attempt+1, maxRetries)
	}
	if err != nil {
		return fmt.Errorf("creating ephemeral container after %d retries: %w", maxRetries, err)
	}

	fmt.Fprintf(os.Stderr, "Waiting for container to start...\n")
	if err := WaitForContainerRunning(ctx, typedClient, opts.Namespace, opts.PodName, containerName, 30*time.Second); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Attaching to container %q...\n", containerName)

	args := []string{"kubectl", "exec", "-it", "-n", opts.Namespace, opts.PodName, "-c", containerName}
	if opts.KubeContext != "" {
		args = append(args, "--context", opts.KubeContext)
	}
	args = append(args, "--", "sh", "-c", "exec $( command -v bash || command -v sh )")

	return syscall.Exec(kubectlPath, args, os.Environ())
}
