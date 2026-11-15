package ui

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFlagsRequiredHidesDefault(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Int("count", 0, "Node count")
	cmd.Flags().String("name", "default-name", "Node group name")
	cmd.Flags().Int("replicas", 3, "Replica count")
	require.NoError(t, cmd.MarkFlagRequired("count"))
	require.NoError(t, cmd.MarkFlagRequired("name"))

	out := formatFlags(cmd.Flags())

	assert.NotContains(t, out, "Default: 0")
	assert.NotContains(t, out, "Default: default-name")
	assert.Contains(t, out, "Default: 3")
}
