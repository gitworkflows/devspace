package provider

import (
	"context"
	"fmt"
	"os"

	"dev.khulnasoft.com/cmd/completion"
	"dev.khulnasoft.com/cmd/flags"
	"dev.khulnasoft.com/pkg/config"
	"dev.khulnasoft.com/pkg/platform"
	provider2 "dev.khulnasoft.com/pkg/provider"
	"dev.khulnasoft.com/pkg/workspace"
	logpkg "dev.khulnasoft.com/log"
	"github.com/spf13/cobra"
)

// DeleteCmd holds the delete cmd flags
type DeleteCmd struct {
	*flags.GlobalFlags

	IgnoreNotFound bool
	Force          bool
}

// NewDeleteCmd creates a new command
func NewDeleteCmd(flags *flags.GlobalFlags) *cobra.Command {
	cmd := &DeleteCmd{
		GlobalFlags: flags,
	}
	deleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a provider",
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.Run(context.Background(), args)
		},
		ValidArgsFunction: func(rootCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.GetProviderSuggestions(rootCmd, cmd.Context, cmd.Provider, args, toComplete, cmd.Owner, logpkg.Default)
		},
	}

	deleteCmd.Flags().BoolVar(&cmd.IgnoreNotFound, "ignore-not-found", false, "Treat \"provider not found\" as a successful delete")
	deleteCmd.Flags().BoolVar(&cmd.Force, "force", false, "Force delete the provider and ignore provider is already used")
	_ = deleteCmd.Flags().MarkHidden("force")
	return deleteCmd
}

func (cmd *DeleteCmd) Run(ctx context.Context, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("please specify a provider to delete")
	}

	devSpaceConfig, err := config.LoadConfig(cmd.Context, cmd.Provider)
	if err != nil {
		return err
	}

	provider := devSpaceConfig.Current().DefaultProvider
	if len(args) > 0 {
		provider = args[0]
	} else if provider == "" {
		return fmt.Errorf("please specify a provider to delete")
	}

	// delete the provider
	err = DeleteProvider(ctx, devSpaceConfig, provider, cmd.IgnoreNotFound, cmd.Force, logpkg.Default)
	if err != nil {
		return err
	}

	logpkg.Default.Donef("Successfully deleted provider '%s'", provider)
	return nil
}

func DeleteProvider(ctx context.Context, devSpaceConfig *config.Config, provider string, ignoreNotFound, force bool, log logpkg.Logger) error {
	// if force is not set, check if the provider is associated with a pro instance or workspace
	if !force {
		// check if this provider is associated with a pro instance
		proInstances, err := workspace.ListProInstances(devSpaceConfig, logpkg.Default)
		if err != nil {
			return fmt.Errorf("list pro instances: %w", err)
		}
		for _, instance := range proInstances {
			if instance.Provider == provider {
				return fmt.Errorf("cannot delete provider '%s', because it is connected to Pro instance '%s'. Removing the Pro instance will automatically delete this provider", instance.Provider, instance.Host)
			}
		}

		// check if there are workspaces that still use this provider
		workspaces, err := workspace.List(ctx, devSpaceConfig, true, platform.AllOwnerFilter, log)
		if err != nil {
			return err
		}

		// search for workspace that uses this machine
		for _, workspace := range workspaces {
			if workspace.Provider.Name == provider {
				return fmt.Errorf("cannot delete provider '%s', because workspace '%s' is still using it. Please delete the workspace '%s' before deleting the provider", workspace.Provider.Name, workspace.ID, workspace.ID)
			}
		}
	}

	return DeleteProviderConfig(devSpaceConfig, provider, ignoreNotFound)
}

func DeleteProviderConfig(devSpaceConfig *config.Config, provider string, ignoreNotFound bool) error {
	if devSpaceConfig.Current().DefaultProvider == provider {
		devSpaceConfig.Current().DefaultProvider = ""
	}
	delete(devSpaceConfig.Current().Providers, provider)
	err := config.SaveConfig(devSpaceConfig)
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	providerDir, err := provider2.GetProviderDir(devSpaceConfig.DefaultContext, provider)
	if err != nil {
		return err
	}
	_, err = os.Stat(providerDir)
	if err != nil {
		if os.IsNotExist(err) {
			if ignoreNotFound {
				return nil
			}

			return fmt.Errorf("provider '%s' does not exist", provider)
		}

		return err
	}
	err = os.RemoveAll(providerDir)
	if err != nil {
		return fmt.Errorf("delete provider dir: %w", err)
	}

	return nil
}
