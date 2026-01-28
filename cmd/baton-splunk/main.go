package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	cfg "github.com/conductorone/baton-splunk/pkg/config"
	"github.com/conductorone/baton-splunk/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-splunk",
		getConnector,
		cfg.Config,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilder(&connector.Splunk{}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func constructAuth(c *cfg.Splunk) string {
	if c.Token != "" {
		return "Bearer " + c.Token
	}

	if c.Username != "" {
		credentials := fmt.Sprintf("%s:%s", c.Username, c.Password)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

		return "Basic " + encodedCredentials
	}

	return ""
}

func getConnector(ctx context.Context, c *cfg.Splunk) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	splunkConnector, err := connector.New(
		ctx,
		constructAuth(c),
		connector.CLIConfig{
			Unsafe:  c.Unsafe,
			Verbose: c.Verbose,
			Cloud:   c.Cloud,
		},
		c.Deployments,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	conn, err := connectorbuilder.NewConnector(ctx, splunkConnector)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return conn, nil
}
