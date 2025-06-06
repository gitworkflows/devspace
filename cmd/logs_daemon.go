package cmd

import (
	"context"
	"fmt"
	"os"

	"dev.khulnasoft.com/cmd/flags"
	"dev.khulnasoft.com/pkg/client"
	"dev.khulnasoft.com/pkg/config"
	provider2 "dev.khulnasoft.com/pkg/provider"
	"dev.khulnasoft.com/pkg/workspace"
	"dev.khulnasoft.com/log"
	"github.com/spf13/cobra"
)

// LogsDaemonCmd holds the configuration
type LogsDaemonCmd struct {
	*flags.GlobalFlags
}

// NewLogsDaemonCmd creates a new destroy command
func NewLogsDaemonCmd(flags *flags.GlobalFlags) *cobra.Command {
	cmd := &LogsDaemonCmd{
		GlobalFlags: flags,
	}
	startCmd := &cobra.Command{
		Use:   "logs-daemon",
		Short: "Prints the daemon logs on the machine",
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.Run(context.Background(), args)
		},
	}

	return startCmd
}

// Run runs the command logic
func (cmd *LogsDaemonCmd) Run(ctx context.Context, args []string) error {
	devSpaceConfig, err := config.LoadConfig(cmd.Context, cmd.Provider)
	if err != nil {
		return err
	}

	baseClient, err := workspace.Get(ctx, devSpaceConfig, args, false, cmd.Owner, false, log.Default)
	if err != nil {
		return err
	} else if baseClient.WorkspaceConfig().Machine.ID == "" {
		return fmt.Errorf("selected workspace is not a machine provider, there is not daemon running")
	}

	workspaceClient, ok := baseClient.(client.WorkspaceClient)
	if !ok {
		return fmt.Errorf("this command is not supported for proxy providers")
	}

	_, agentInfo, err := workspaceClient.AgentInfo(provider2.CLIOptions{})
	if err != nil {
		return err
	}

	command := fmt.Sprintf("'%s' agent workspace logs-daemon --context '%s' --id '%s'", workspaceClient.AgentPath(), workspaceClient.Context(), workspaceClient.Workspace())
	if agentInfo.Agent.DataPath != "" {
		command += fmt.Sprintf(" --agent-dir '%s'", agentInfo.Agent.DataPath)
	}

	// read daemon logs
	return workspaceClient.Command(ctx, client.CommandOptions{
		Command: command,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
}
