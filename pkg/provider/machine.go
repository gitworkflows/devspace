package provider

import (
	"dev.khulnasoft.com/pkg/config"
	"dev.khulnasoft.com/pkg/types"
)

type Machine struct {
	// ID is the machine id to use
	ID string `json:"id,omitempty"`

	// Provider is the provider used to create this workspace
	Provider MachineProviderConfig `json:"provider,omitempty"`

	// CreationTimestamp is the timestamp when this workspace was created
	CreationTimestamp types.Time `json:"creationTimestamp,omitempty"`

	// Context is the context where this config file was loaded from
	Context string `json:"context,omitempty"`

	// Origin is the place where this config file was loaded from
	Origin string `json:"-"`
}

type MachineProviderConfig struct {
	// Name is the provider name used to deploy this machine
	Name string `json:"name,omitempty"`

	// Options are the local options that override the global ones
	Options map[string]config.OptionValue `json:"options,omitempty"`
}
