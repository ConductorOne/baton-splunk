package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/conductorone/baton-splunk/pkg/connector"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	cfg := &config{}
	cmd, err := cli.NewCmd(ctx, "baton-splunk", cfg, validateConfig, getConnector)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version
	cmdFlags(cmd)

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func constructAuth(cfg *config) string {
	if cfg.AccessToken != "" {
		return "Bearer " + cfg.AccessToken
	}

	if cfg.Username != "" {
		credentials := fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

		return "Basic " + encodedCredentials
	}

	return ""
}

func getConnector(ctx context.Context, cfg *config) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	splunkConnector, err := connector.New(
		ctx,
		constructAuth(cfg),
		connector.CLIConfig{
			Unsafe:  cfg.Unsafe,
			Verbose: cfg.Verbose,
			Cloud:   cfg.Cloud,
		},
		cfg.Deployments,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	connector, err := connectorbuilder.NewConnector(ctx, splunkConnector)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return connector, nil
}
