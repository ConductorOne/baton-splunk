package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	AccessToken string   `mapstructure:"token"`
	Unsafe      bool     `mapstructure:"unsafe"`
	Verbose     bool     `mapstructure:"verbose"`
	Cloud       bool     `mapstructure:"cloud"`
	Deployments []string `mapstructure:"deployments"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.AccessToken == "" {
		return fmt.Errorf("access token is missing")
	}

	if cfg.Cloud && len(cfg.Deployments) == 0 {
		return fmt.Errorf("cloud mode requires at least one deployment")
	}

	return nil
}

// cmdFlags sets the cmdFlags required for the connector.
func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("token", "", "The Splunk access token used to connect to the Splunk API. ($BATON_TOKEN)")
	cmd.PersistentFlags().Bool("unsafe", false, "Allow insecure TLS connections to Splunk. ($BATON_UNSAFE)")
	cmd.PersistentFlags().Bool("verbose", false, "Enable listing verbose entitlements for Role capabilities. ($BATON_VERBOSE)")
	cmd.PersistentFlags().Bool("cloud", false, "Switches to cloud API endpoints. ($BATON_CLOUD)")
	cmd.PersistentFlags().StringSlice("deployments", []string{}, "Limit syncing to specific deployments by specifying cloud deployment names or IP addresses of on-premise deployments. ($BATON_DEPLOYMENTS)")
}
