package debug

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateContainerName(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers:     []corev1.Container{{Name: "app"}},
			InitContainers: []corev1.Container{{Name: "init"}},
			EphemeralContainers: []corev1.EphemeralContainer{
				{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debugger-old"}},
			},
		},
	}

	name := GenerateContainerName(pod)
	assert.Regexp(t, `^debugger-[a-f0-9]{5}$`, name)
	assert.NotEqual(t, "app", name)
	assert.NotEqual(t, "init", name)
	assert.NotEqual(t, "debugger-old", name)
}

func TestBuildEphemeralContainer(t *testing.T) {
	ec := BuildEphemeralContainer("debugger-abc12", "debug:latest", "app", []corev1.Capability{"NET_RAW", "SYS_PTRACE"})

	assert.Equal(t, "debugger-abc12", ec.Name)
	assert.Equal(t, "debug:latest", ec.Image)
	assert.Equal(t, "app", ec.TargetContainerName)
	assert.True(t, ec.Stdin)
	assert.True(t, ec.TTY)

	require.NotNil(t, ec.SecurityContext)
	require.NotNil(t, ec.SecurityContext.RunAsUser)
	assert.Equal(t, int64(0), *ec.SecurityContext.RunAsUser)

	require.NotNil(t, ec.SecurityContext.Capabilities)
	assert.Contains(t, ec.SecurityContext.Capabilities.Add, corev1.Capability("NET_RAW"))
	assert.Contains(t, ec.SecurityContext.Capabilities.Add, corev1.Capability("SYS_PTRACE"))

	assert.Equal(t, corev1.ResourceRequirements{}, ec.Resources)
}

func TestBuildEphemeralContainer_CustomCaps(t *testing.T) {
	caps := []corev1.Capability{"NET_RAW", "SYS_PTRACE", "SYS_ADMIN"}
	ec := BuildEphemeralContainer("dbg", "img", "target", caps)

	assert.Len(t, ec.SecurityContext.Capabilities.Add, 3)
	assert.Contains(t, ec.SecurityContext.Capabilities.Add, corev1.Capability("SYS_ADMIN"))
}

func TestFirstContainerName(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main-app"},
				{Name: "sidecar"},
			},
		},
	}
	assert.Equal(t, "main-app", FirstContainerName(pod))
}

func TestFirstContainerName_Empty(t *testing.T) {
	pod := &corev1.Pod{}
	assert.Equal(t, "", FirstContainerName(pod))
}

func TestGenerateContainerName_Uniqueness(t *testing.T) {
	pod := &corev1.Pod{}
	names := make(map[string]struct{})
	for range 100 {
		name := GenerateContainerName(pod)
		require.NotContains(t, names, name, "generated duplicate name")
		names[name] = struct{}{}
	}
}
