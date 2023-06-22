package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-splunk/pkg/splunk"
)

type deploymentResourceType struct {
	resourceType *v2.ResourceType
	client       *splunk.Client
	deployments  []string
	verbose      bool
}

func (d *deploymentResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return d.resourceType
}

// deploymentResource creates a new connector resource for a Splunk Deployment under which all other resources are scoped.
func deploymentResource(ctx context.Context, deployment string) (*v2.Resource, error) {
	displayName := titleCaser.String(deployment)

	resource, err := rs.NewResource(
		displayName,
		resourceTypeRole,
		deployment,
		rs.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: resourceTypeRole.Id},
			&v2.ChildResourceType{ResourceTypeId: resourceTypeUser.Id},
			&v2.ChildResourceType{ResourceTypeId: resourceTypeApplication.Id},
			// TODO: idp Applications
			// &v2.ChildResourceType{ResourceTypeId: resourceTypeApplicationIdp.Id},
		),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (d *deploymentResourceType) List(ctx context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	rv := make([]*v2.Resource, 0)

	// If no deployments are specified, return the localhost deployment
	if len(d.deployments) == 0 {
		dr, err := deploymentResource(ctx, "localhost")
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, dr)

		return rv, "", nil, nil
	}

	for _, deployment := range d.deployments {
		dr, err := deploymentResource(ctx, deployment)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, dr)
	}

	return rv, "", nil, nil
}

func (d *deploymentResourceType) Entitlements(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	if d.verbose {
		bag, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: resourceTypeUser.Id})
		if err != nil {
			return nil, "", nil, err
		}

		d.client.PointToDeployment(resource.Id.Resource)
		capabilitiesEntry, nextPage, err := d.client.GetCapabilities(
			ctx,
			splunk.PaginationVars{
				Limit: ResourcesPageSize,
				Page:  bag.PageToken(),
			},
		)
		if err != nil {
			return nil, "", nil, fmt.Errorf("splunk-connector: failed to get capabilities: %w", err)
		}

		pageToken, err := bag.NextToken(nextPage)
		if err != nil {
			return nil, "", nil, err
		}

		var rv []*v2.Entitlement
		for _, capabilityEntry := range capabilitiesEntry {
			for _, capability := range capabilityEntry.Content.Capabilities {
				entitlementOptions := []ent.EntitlementOption{
					ent.WithGrantableTo(resourceTypeRole),
					ent.WithDisplayName(fmt.Sprintf("%s capability", capability)),
					ent.WithDescription(fmt.Sprintf("%s Splunk capability", capability)),
				}

				rv = append(rv, ent.NewPermissionEntitlement(resource, capability, entitlementOptions...))
			}

		}

		return rv, pageToken, nil, nil
	}

	return nil, "", nil, nil

}

func (d *deploymentResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	// Grant only the deployment capabilities if verbose mode is enabled.
	if d.verbose {
		bag, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: resourceTypeUser.Id})
		if err != nil {
			return nil, "", nil, err
		}

		d.client.PointToDeployment(resource.Id.Resource)
		roles, nextPage, err := d.client.GetRoles(
			ctx,
			splunk.PaginationVars{
				Limit: ResourcesPageSize,
				Page:  bag.PageToken(),
			},
		)
		if err != nil {
			return nil, "", nil, fmt.Errorf("splunk-connector: failed to get roles: %w", err)
		}

		pageToken, err := bag.NextToken(nextPage)
		if err != nil {
			return nil, "", nil, err
		}

		var rv []*v2.Grant
		for _, role := range roles {
			roleCopy := role

			rr, err := roleResource(ctx, &roleCopy)
			if err != nil {
				return nil, "", nil, fmt.Errorf("splunk-connector: failed to build role resource: %w", err)
			}

			for _, capability := range role.Content.Capabilities {
				rv = append(rv, grant.NewGrant(
					resource,
					capability,
					rr.Id,
				))
			}
		}

		return rv, pageToken, nil, nil
	}

	return nil, "", nil, nil
}

func deploymentBuilder(client *splunk.Client, verbose bool, deployments []string) *deploymentResourceType {
	return &deploymentResourceType{
		resourceType: resourceTypeDeployment,
		client:       client,
		verbose:      verbose,
		deployments:  deployments,
	}
}
