package nodegroup

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-cli/v3/internal/apierror"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/commands"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/completion"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/output"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/resolver"
	upcloudv9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type scaleCommand struct {
	*commands.BaseCommand
	name  string
	count int
	completion.Kubernetes
	resolver.CachingKubernetes
}

// ScaleCommand creates the "kubernetes nodegroup scale" command
func ScaleCommand() commands.Command {
	return &scaleCommand{
		BaseCommand: commands.New(
			"scale",
			"Scale the number of nodes in the node group.",
			"upctl kubernetes nodegroup scale 55199a44-4751-4e27-9394-7c7661910be3 --name secondary-node-group --count 3",
		),
	}
}

// InitCommand implements Command.InitCommand
func (s *scaleCommand) InitCommand() {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVar(&s.name, "name", "", "Node group name")
	flagSet.IntVar(&s.count, "count", 0, "Node count")
	s.AddFlags(flagSet)

	commands.Must(s.Cobra().MarkFlagRequired("name"))
	commands.Must(s.Cobra().MarkFlagRequired("count"))
	commands.Must(s.Cobra().RegisterFlagCompletionFunc("name", cobra.NoFileCompletions))
	commands.Must(s.Cobra().RegisterFlagCompletionFunc("count", cobra.NoFileCompletions))
}

// ExecuteSingleArgument implements commands.SingleArgumentCommand
func (s *scaleCommand) ExecuteSingleArgument(exec commands.Executor, arg string) (output.Output, error) {
	msg := fmt.Sprintf("Scaling node group %s of cluster %v", s.name, arg)
	exec.PushProgressStarted(msg)

	client, err := v9Client(exec)
	if err != nil {
		return commands.HandleError(exec, msg, err)
	}
	id, err := uuid.Parse(arg)
	if err != nil {
		return commands.HandleError(exec, msg, fmt.Errorf("invalid kubernetes cluster UUID %q: %w", arg, err))
	}

	res, err := client.ModifyKubernetesClusterNodeGroupWithResponse(exec.Context(), id, s.name, upcloudv9.KubernetesNodeGroupPatch{
		Count: int32(s.count),
	})
	if err != nil {
		return commands.HandleError(exec, msg, err)
	}
	if res.JSON200 == nil {
		return commands.HandleError(exec, msg, apierror.FromResponse(res.StatusCode(), res.Body))
	}

	legacy, err := legacyNodeGroupOutput(res.JSON200)
	if err != nil {
		return commands.HandleError(exec, msg, err)
	}

	exec.PushProgressSuccess(msg)

	return output.OnlyMarshaled{Value: legacy}, nil
}
