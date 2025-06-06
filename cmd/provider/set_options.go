package provider

import (
	"context"
	"fmt"
	"os"

	"dev.khulnasoft.com/cmd/completion"
	"dev.khulnasoft.com/cmd/flags"
	"dev.khulnasoft.com/pkg/config"
	"dev.khulnasoft.com/pkg/workspace"
	"dev.khulnasoft.com/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// SetOptionsCmd holds the use cmd flags
type SetOptionsCmd struct {
	flags.GlobalFlags

	Dry bool

	Reconfigure   bool
	SingleMachine bool
	Options       []string
}

// NewSetOptionsCmd creates a new command
func NewSetOptionsCmd(flags *flags.GlobalFlags) *cobra.Command {
	cmd := &SetOptionsCmd{
		GlobalFlags: *flags,
	}
	setOptionsCmd := &cobra.Command{
		Use:   "set-options [provider]",
		Short: "Sets options for the given provider. Similar to 'devspace provider use', but does not switch the default provider.",
		RunE: func(_ *cobra.Command, args []string) error {
			logger := log.Logger(log.Default)
			if cmd.Dry {
				logger = log.Default.ErrorStreamOnly()
			}

			return cmd.Run(context.Background(), args, logger)
		},
		ValidArgsFunction: func(rootCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.GetProviderSuggestions(rootCmd, cmd.Context, cmd.Provider, args, toComplete, cmd.Owner, log.Default)
		},
	}

	setOptionsCmd.Flags().BoolVar(&cmd.SingleMachine, "single-machine", false, "If enabled will use a single machine for all workspaces")
	setOptionsCmd.Flags().BoolVar(&cmd.Reconfigure, "reconfigure", false, "If enabled will not merge existing provider config")
	setOptionsCmd.Flags().StringArrayVarP(&cmd.Options, "option", "o", []string{}, "Provider option in the form KEY=VALUE")
	setOptionsCmd.Flags().BoolVar(&cmd.Dry, "dry", false, "Dry will not persist the options to file and instead return the new filled options")
	return setOptionsCmd
}

// Run runs the command logic
func (cmd *SetOptionsCmd) Run(ctx context.Context, args []string, log log.Logger) error {
	devSpaceConfig, err := config.LoadConfig(cmd.Context, cmd.Provider)
	if err != nil {
		return err
	}

	providerName := devSpaceConfig.Current().DefaultProvider
	if len(args) > 0 {
		providerName = args[0]
	} else if providerName == "" {
		return fmt.Errorf("please specify a provider")
	}
	log.Debugf("providerName=%+v", providerName)

	if os.Getenv("DEVSPACE_UI") == "" && len(cmd.Options) == 0 {
		return fmt.Errorf("please specify option")
	}
	log.Debugf("Options=%+v", cmd.Options)

	providerWithOptions, err := workspace.FindProvider(devSpaceConfig, providerName, log)
	if err != nil {
		return err
	}

	devSpaceConfig, err = setOptions(
		ctx,
		providerWithOptions.Config,
		devSpaceConfig.DefaultContext,
		cmd.Options,
		cmd.Reconfigure,
		cmd.Dry,
		cmd.Dry,
		false,
		&cmd.SingleMachine,
		log,
	)
	if err != nil {
		return err
	}

	// save provider config
	if !cmd.Dry {
		err = config.SaveConfig(devSpaceConfig)
		if err != nil {
			return errors.Wrap(err, "save config")
		}
	} else {
		// print options to stdout
		err = printOptions(devSpaceConfig, providerWithOptions, "json", true)
		if err != nil {
			return fmt.Errorf("print options: %w", err)
		}
	}

	// print success message
	log.Donef("Successfully set options for provider '%s'", providerWithOptions.Config.Name)
	return nil
}
