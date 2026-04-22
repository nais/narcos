package debug

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var policyExceptionGVR = schema.GroupVersionResource{
	Group:    "kyverno.io",
	Version:  "v2",
	Resource: "policyexceptions",
}

type PolicyViolation struct {
	PolicyName string
	RuleName   string
}

var kyvernoDenialMarker = "blocked due to the following policies"

// policyBlockRe matches "policyName:\n  ruleName: message" blocks in Kyverno denial messages.
var policyBlockRe = regexp.MustCompile(`(?m)^([a-zA-Z][\w-]*):\n((?:[ \t]+\S[^\n]*\n?)+)`)

// ruleLineRe matches "  ruleName: message" lines within a policy block.
var ruleLineRe = regexp.MustCompile(`(?m)^\s+(\S+):`)

func ParseKyvernoDenial(err error) []PolicyViolation {
	if err == nil {
		return nil
	}

	msg := err.Error()
	if !strings.Contains(msg, kyvernoDenialMarker) {
		return nil
	}

	idx := strings.Index(msg, kyvernoDenialMarker)
	body := msg[idx+len(kyvernoDenialMarker):]

	var violations []PolicyViolation
	for _, match := range policyBlockRe.FindAllStringSubmatch(body, -1) {
		policyName := strings.TrimSpace(match[1])
		rulesBlock := match[2]
		for _, ruleMatch := range ruleLineRe.FindAllStringSubmatch(rulesBlock, -1) {
			violations = append(violations, PolicyViolation{
				PolicyName: policyName,
				RuleName:   ruleMatch[1],
			})
		}
	}

	return violations
}

func KyvernoCRDExists(ctx context.Context, client dynamic.Interface) bool {
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	_, err := client.Resource(crdGVR).Get(ctx, "policyexceptions.kyverno.io", metav1.GetOptions{})
	return err == nil
}

func BuildPolicyException(pod *corev1.Pod, violations []PolicyViolation) *unstructured.Unstructured {
	name := "debug-" + pod.Name
	if len(name) > 253 {
		name = name[:253]
	}

	exceptions := make([]any, 0, len(violations))
	for _, v := range violations {
		exceptions = append(exceptions, map[string]any{
			"policyName": v.PolicyName,
			"ruleNames":  []any{v.RuleName},
		})
	}

	pe := &unstructured.Unstructured{}
	pe.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kyverno.io",
		Version: "v2",
		Kind:    "PolicyException",
	})
	pe.SetName(name)
	pe.SetNamespace(pod.Namespace)
	pe.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         "v1",
			Kind:               "Pod",
			Name:               pod.Name,
			UID:                pod.UID,
			Controller:         ptrBool(true),
			BlockOwnerDeletion: ptrBool(true),
		},
	})
	pe.Object["spec"] = map[string]any{
		"match": map[string]any{
			"any": []any{
				map[string]any{
					"resources": map[string]any{
						"kinds":      []any{"Pod"},
						"names":      []any{pod.Name},
						"namespaces": []any{pod.Namespace},
					},
				},
			},
		},
		"exceptions": exceptions,
	}

	return pe
}

func ptrBool(b bool) *bool { return &b }

func CreatePolicyException(ctx context.Context, client dynamic.Interface, pe *unstructured.Unstructured) error {
	ns := pe.GetNamespace()
	_, err := client.Resource(policyExceptionGVR).Namespace(ns).Create(ctx, pe, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		existing, getErr := client.Resource(policyExceptionGVR).Namespace(ns).Get(ctx, pe.GetName(), metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("getting existing PolicyException: %w", getErr)
		}
		pe.SetResourceVersion(existing.GetResourceVersion())
		_, err = client.Resource(policyExceptionGVR).Namespace(ns).Update(ctx, pe, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("creating PolicyException: %w", err)
	}
	return nil
}
