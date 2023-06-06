package connector

import (
	"context"
	"crypto/tls"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-splunk/pkg/splunk"
)

var (
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
			v2.ResourceType_TRAIT_ROLE,
		},
	}
)

type Splunk struct {
	client *splunk.Client
}

func (sp *Splunk) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		userBuilder(sp.client),
		roleBuilder(sp.client),
	}
}

// Metadata returns metadata about the connector.
func (sp *Splunk) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Splunk",
		Description: "Connector syncing Splunk users their roles and capabilities to Baton.",
	}, nil
}

// TODO: implement this
// Validate hits the Splunk API to validate that the configured credentials are valid and compatible.
func (sp *Splunk) Validate(ctx context.Context) (annotations.Annotations, error) {
	// should be able to list users
	// _, err := sp.client.GetUsers(ctx)
	// if err != nil {
	// 	return nil, status.Error(codes.Unauthenticated, "Provided Access Token is invalid")
	// }

	return nil, nil
}

// New returns the Splunk connector.
func New(ctx context.Context, password string) (*Splunk, error) {
	// TODO: remove TLS skip
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// httpClient, err := uhttp.NewClient(
	// 	ctx,
	// 	uhttp.WithLogger(true, ctxzap.Extract(ctx)),
	// )
	// if err != nil {
	// 	return nil, err
	// }

	return &Splunk{
		client: splunk.NewClient(&httpClient, password),
	}, nil
}
