package pro

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dev.khulnasoft.com/cmd/pro/flags"
	"dev.khulnasoft.com/pkg/client/clientimplementation"
	"dev.khulnasoft.com/pkg/config"
	"dev.khulnasoft.com/pkg/provider"
	"dev.khulnasoft.com/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// WatchWorkspacesCmd holds the cmd flags
type WatchWorkspacesCmd struct {
	*flags.GlobalFlags
	Log log.Logger

	Host          string
	Project       string
	FilterByOwner bool
}

// NewWatchWorkspacesCmd creates a new command
func NewWatchWorkspacesCmd(globalFlags *flags.GlobalFlags) *cobra.Command {
	cmd := &WatchWorkspacesCmd{
		GlobalFlags: globalFlags,
		Log:         log.GetInstance(),
	}
	c := &cobra.Command{
		Use:    "watch-workspaces",
		Short:  "Watch workspaces",
		Hidden: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			devSpaceConfig, provider, err := findProProvider(cobraCmd.Context(), cmd.Context, cmd.Provider, cmd.Host, cmd.Log)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), devSpaceConfig, provider)
		},
	}

	c.Flags().StringVar(&cmd.Host, "host", "", "The pro instance to use")
	_ = c.MarkFlagRequired("host")
	c.Flags().StringVar(&cmd.Project, "project", "", "The project to use")
	_ = c.MarkFlagRequired("project")
	c.Flags().BoolVar(&cmd.FilterByOwner, "filter-by-owner", true, "If true only shows workspaces of current owner")

	return c
}

func (cmd *WatchWorkspacesCmd) Run(ctx context.Context, devSpaceConfig *config.Config, providerConfig *provider.ProviderConfig) error {
	opts := devSpaceConfig.ProviderOptions(providerConfig.Name)
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if cmd.FilterByOwner {
		opts[provider.LOFT_FILTER_BY_OWNER] = config.OptionValue{Value: "true"}
	}
	opts[provider.LOFT_PROJECT] = config.OptionValue{Value: cmd.Project}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	// ignore --debug because we tunnel json through stdio
	cmd.Log.SetLevel(logrus.InfoLevel)

	err := clientimplementation.RunCommandWithBinaries(
		cancelCtx,
		"watchWorkspaces",
		providerConfig.Exec.Proxy.Watch.Workspaces,
		devSpaceConfig.DefaultContext,
		nil,
		nil,
		opts,
		providerConfig,
		nil,
		nil,
		os.Stdout,
		log.Default.ErrorStreamOnly().Writer(logrus.ErrorLevel, false),
		cmd.Log)
	if err != nil {
		return fmt.Errorf("watch workspaces with provider \"%s\": %w", providerConfig.Name, err)
	}

	return nil
}
