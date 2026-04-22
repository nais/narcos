package debug

import (
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKyvernoDenial_SingleViolation(t *testing.T) {
	err := fmt.Errorf(`admission webhook "validate.kyverno.svc" denied the request: resource Pod/ns/pod was blocked due to the following policies:

disallow-capabilities:
  adding-capabilities: Any capabilities added beyond the allowed list are disallowed.
`)
	violations := ParseKyvernoDenial(err)
	require.Len(t, violations, 1)
	assert.Equal(t, "disallow-capabilities", violations[0].PolicyName)
	assert.Equal(t, "adding-capabilities", violations[0].RuleName)
}

func TestParseKyvernoDenial_MultipleViolations(t *testing.T) {
	err := fmt.Errorf(`admission webhook "validate.kyverno.svc" denied the request: resource Pod/ns/pod was blocked due to the following policies:

disallow-capabilities:
  adding-capabilities: Any capabilities added beyond the allowed list are disallowed.

disallow-privilege-escalation:
  privilege-escalation: Privilege escalation is disallowed.
`)
	violations := ParseKyvernoDenial(err)
	require.Len(t, violations, 2)
	assert.Equal(t, "disallow-capabilities", violations[0].PolicyName)
	assert.Equal(t, "adding-capabilities", violations[0].RuleName)
	assert.Equal(t, "disallow-privilege-escalation", violations[1].PolicyName)
	assert.Equal(t, "privilege-escalation", violations[1].RuleName)
}

func TestParseKyvernoDenial_NotKyverno(t *testing.T) {
	err := fmt.Errorf("pods is forbidden: User cannot patch resource")
	violations := ParseKyvernoDenial(err)
	assert.Empty(t, violations)
}

func TestParseKyvernoDenial_EmptyError(t *testing.T) {
	violations := ParseKyvernoDenial(nil)
	assert.Empty(t, violations)
}

func TestParseKyvernoDenial_MalformedMessage(t *testing.T) {
	err := fmt.Errorf("blocked due to the following policies: garbage data without proper format")
	violations := ParseKyvernoDenial(err)
	assert.Empty(t, violations)
}

func TestBuildPolicyException_OwnerRef(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "my-ns",
			UID:       types.UID("abc-123"),
		},
	}
	violations := []PolicyViolation{{PolicyName: "pol", RuleName: "rule"}}

	pe := BuildPolicyException(pod, violations)

	refs := pe.GetOwnerReferences()
	require.Len(t, refs, 1)
	assert.Equal(t, "v1", refs[0].APIVersion)
	assert.Equal(t, "Pod", refs[0].Kind)
	assert.Equal(t, "my-pod", refs[0].Name)
	assert.Equal(t, types.UID("abc-123"), refs[0].UID)
	require.NotNil(t, refs[0].Controller)
	assert.True(t, *refs[0].Controller)
	require.NotNil(t, refs[0].BlockOwnerDeletion)
	assert.True(t, *refs[0].BlockOwnerDeletion)
}

func TestBuildPolicyException_Spec(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "my-ns",
			UID:       types.UID("uid-1"),
		},
	}
	violations := []PolicyViolation{
		{PolicyName: "disallow-capabilities", RuleName: "adding-capabilities"},
		{PolicyName: "disallow-privilege-escalation", RuleName: "privilege-escalation"},
	}

	pe := BuildPolicyException(pod, violations)

	assert.Equal(t, "debug-my-pod", pe.GetName())
	assert.Equal(t, "my-ns", pe.GetNamespace())
	assert.Equal(t, "PolicyException", pe.GetKind())
	assert.Equal(t, "kyverno.io/v2", pe.GetAPIVersion())

	spec, ok := pe.Object["spec"].(map[string]any)
	require.True(t, ok)

	match := spec["match"].(map[string]any)
	anyList := match["any"].([]any)
	require.Len(t, anyList, 1)
	resources := anyList[0].(map[string]any)["resources"].(map[string]any)
	assert.Equal(t, []any{"Pod"}, resources["kinds"])
	assert.Equal(t, []any{"my-pod"}, resources["names"])
	assert.Equal(t, []any{"my-ns"}, resources["namespaces"])

	exceptions := spec["exceptions"].([]any)
	require.Len(t, exceptions, 2)
	assert.Equal(t, "disallow-capabilities", exceptions[0].(map[string]any)["policyName"])
	assert.Equal(t, []any{"adding-capabilities"}, exceptions[0].(map[string]any)["ruleNames"])
}

func TestBuildPolicyException_NameTruncation(t *testing.T) {
	longName := strings.Repeat("a", 260)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      longName,
			Namespace: "ns",
		},
	}
	pe := BuildPolicyException(pod, []PolicyViolation{{PolicyName: "p", RuleName: "r"}})
	assert.LessOrEqual(t, len(pe.GetName()), 253)
}
