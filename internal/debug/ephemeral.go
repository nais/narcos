package debug

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

// GenerateContainerName creates a unique debugger container name that doesn't
// collide with any existing container in the pod.
func GenerateContainerName(pod *corev1.Pod) string {
	existing := make(map[string]struct{})
	for _, c := range pod.Spec.Containers {
		existing[c.Name] = struct{}{}
	}
	for _, c := range pod.Spec.InitContainers {
		existing[c.Name] = struct{}{}
	}
	for _, c := range pod.Spec.EphemeralContainers {
		existing[c.Name] = struct{}{}
	}

	for {
		name := "debugger-" + randomSuffix(5)
		if _, ok := existing[name]; !ok {
			return name
		}
	}
}

func randomSuffix(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// BuildEphemeralContainer constructs an ephemeral container spec with root access
// and the specified capabilities.
func BuildEphemeralContainer(name, image, targetContainer string, caps []corev1.Capability) corev1.EphemeralContainer {
	return corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  name,
			Image: image,
			Stdin: true,
			TTY:   true,
			SecurityContext: &corev1.SecurityContext{
				RunAsUser:    ptr.To(int64(0)),
				RunAsNonRoot: ptr.To(false),
				Capabilities: &corev1.Capabilities{
					Add: caps,
				},
			},
		},
		TargetContainerName: targetContainer,
	}
}

// PatchEphemeralContainer adds an ephemeral container to a pod using a strategic
// merge patch on the ephemeralcontainers subresource.
func PatchEphemeralContainer(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod, ec corev1.EphemeralContainer) (*corev1.Pod, error) {
	oldData, err := json.Marshal(pod)
	if err != nil {
		return nil, fmt.Errorf("marshaling pod: %w", err)
	}

	modified := pod.DeepCopy()
	modified.Spec.EphemeralContainers = append(modified.Spec.EphemeralContainers, ec)

	newData, err := json.Marshal(modified)
	if err != nil {
		return nil, fmt.Errorf("marshaling modified pod: %w", err)
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Pod{})
	if err != nil {
		return nil, fmt.Errorf("creating patch: %w", err)
	}

	result, err := client.CoreV1().Pods(pod.Namespace).Patch(
		ctx, pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers",
	)
	if err != nil {
		return nil, fmt.Errorf("patching ephemeral container: %w", err)
	}

	return result, nil
}

// WaitForContainerRunning polls until the named ephemeral container reaches
// Running state or the timeout expires.
func WaitForContainerRunning(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastStatus string
	for {
		select {
		case <-deadline:
			return fmt.Errorf("timed out waiting for container %q to start (last status: %s)", containerName, lastStatus)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("getting pod: %w", err)
			}

			for _, cs := range pod.Status.EphemeralContainerStatuses {
				if cs.Name != containerName {
					continue
				}
				if cs.State.Running != nil {
					return nil
				}
				if cs.State.Waiting != nil {
					lastStatus = cs.State.Waiting.Reason
					if lastStatus == "ImagePullBackOff" || lastStatus == "ErrImagePull" {
						return fmt.Errorf("container %q failed to pull image: %s", containerName, cs.State.Waiting.Message)
					}
				}
				if cs.State.Terminated != nil {
					return fmt.Errorf("container %q terminated: %s (exit code %d)", containerName, cs.State.Terminated.Reason, cs.State.Terminated.ExitCode)
				}
			}
			lastStatus = "pending"
		}
	}
}

// FirstContainerName returns the name of the first container in the pod spec.
func FirstContainerName(pod *corev1.Pod) string {
	if len(pod.Spec.Containers) == 0 {
		return ""
	}
	return pod.Spec.Containers[0].Name
}
