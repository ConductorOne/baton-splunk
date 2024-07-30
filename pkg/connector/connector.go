package connector

import (
	"context"
	"crypto/tls"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-splunk/pkg/splunk"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	resourceTypeDeployment = &v2.ResourceType{
		Id:          "deployment",
		DisplayName: "Deployment",
	}
	resourceTypeUser = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_USER,
		},
		Annotations: annotationsForUserResourceType(),
	}
	resourceTypeRole = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Role",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_GROUP,
		},
	}
	resourceTypeApplication = &v2.ResourceType{
		Id:          "application",
		DisplayName: "Application",
	}
)

type Splunk struct {
	client  *splunk.Client
	verbose bool

	cloud       bool
	deployments []string
}

func (sp *Splunk) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	builders := []connectorbuilder.ResourceSyncer{
		deploymentBuilder(sp.client, sp.verbose, sp.deployments),
		userBuilder(sp.client),
		roleBuilder(sp.client),
	}

	// Applications are only supported for on-premise Splunk deployments.
	if !sp.client.Cloud {
		builders = append(builders, applicationBuilder(sp.client, sp.verbose))
	}

	return builders
}

// Metadata returns metadata about the connector.
func (sp *Splunk) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Splunk",
		Description: "Connector syncing Splunk users their roles and other Splunk resources to Baton.",
	}, nil
}

// Validate hits the Splunk API to validate that the configured credentials are valid and compatible.
func (sp *Splunk) Validate(ctx context.Context) (annotations.Annotations, error) {
	if len(sp.deployments) == 0 {
		// no need to point to a deployment since client is initialized with default localhost deployment
		_, _, err := sp.client.GetUsers(ctx, splunk.PaginationVars{Limit: 1})
		if err != nil {
			return nil, status.Error(
				codes.Unauthenticated,
				"Provided Password or Access Token is invalid for the given localhost deployment",
			)
		}
	}

	for _, deployment := range sp.deployments {
		sp.client.PointToDeployment(deployment)

		// should be able to list users
		_, _, err := sp.client.GetUsers(ctx, splunk.PaginationVars{Limit: 1})
		if err != nil {
			return nil, status.Errorf(
				codes.Unauthenticated,
				"Provided Password or Access Token is invalid for the given deployment %s",
				deployment,
			)
		}
	}

	sp.client.ResetPointer()

	return nil, nil
}

type CLIConfig struct {
	Unsafe  bool
	Verbose bool
	Cloud   bool
}

// New returns the Splunk connector.
func New(ctx context.Context, auth string, config CLIConfig, deployments []string) (*Splunk, error) {
	options := []uhttp.Option{
		uhttp.WithLogger(true, ctxzap.Extract(ctx)),
	}

	// Skip TLS verification if flag `unsafe` is specified.
	if config.Unsafe { // #nosec G402
		options = append(
			options,
			uhttp.WithTLSClientConfig(
				&tls.Config{InsecureSkipVerify: true}, // #nosec G402
			),
		)
	}

	httpClient, err := uhttp.NewClient(
		ctx,
		options...,
	)
	if err != nil {
		return nil, err
	}

	return &Splunk{
		client:      splunk.NewClient(httpClient, auth, config.Cloud),
		verbose:     config.Verbose,
		cloud:       config.Cloud,
		deployments: deployments,
	}, nil
}
