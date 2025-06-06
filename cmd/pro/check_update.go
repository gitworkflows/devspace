package pro

import (
	"context"
	"encoding/json"
	"fmt"

	"dev.khulnasoft.com/cmd/agent"
	"dev.khulnasoft.com/cmd/pro/flags"
	"dev.khulnasoft.com/pkg/config"
	"dev.khulnasoft.com/pkg/platform"
	"dev.khulnasoft.com/pkg/provider"
	versionpkg "dev.khulnasoft.com/pkg/version"
	"dev.khulnasoft.com/log"
	"github.com/spf13/cobra"
)

// CheckUpdateCmd holds the cmd flags
type CheckUpdateCmd struct {
	*flags.GlobalFlags
	Log log.Logger

	Host string
}

// NewCheckUpdateCmd creates a new command
func NewCheckUpdateCmd(globalFlags *flags.GlobalFlags) *cobra.Command {
	cmd := &CheckUpdateCmd{
		GlobalFlags: globalFlags,
		Log:         log.GetInstance(),
	}
	c := &cobra.Command{
		Use:    "check-update",
		Short:  "Check platform provider update",
		Hidden: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			devSpaceConfig, provider, err := findProProvider(cobraCmd.Context(), cmd.Context, cmd.Provider, cmd.Host, cmd.Log)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), devSpaceConfig, provider)
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()
			if root == nil {
				return
			}
			if root.Annotations == nil {
				root.Annotations = map[string]string{}
			}
			// Don't print debug message
			root.Annotations[agent.AgentExecutedAnnotation] = "true"
		},
	}

	c.Flags().StringVar(&cmd.Host, "host", "", "The pro instance to use")
	_ = c.MarkFlagRequired("host")

	return c
}

type ProviderUpdateInfo struct {
	Available  bool   `json:"available,omitempty"`
	NewVersion string `json:"newVersion,omitempty"`
}

func (cmd *CheckUpdateCmd) Run(ctx context.Context, devSpaceConfig *config.Config, provider *provider.ProviderConfig) error {
	remoteVersion, err := platform.GetDevSpaceVersion(fmt.Sprintf("https://%s", cmd.Host))
	if err != nil {
		return err
	}

	providerUpdate := ProviderUpdateInfo{}
	if provider.Version == versionpkg.DevVersion {
		providerUpdate.Available = false
	} else if provider.Version != remoteVersion {
		providerUpdate.Available = true
		providerUpdate.NewVersion = remoteVersion
	}

	out, err := json.Marshal(providerUpdate)
	if err != nil {
		return err
	}

	fmt.Print(string(out))

	return nil
}
