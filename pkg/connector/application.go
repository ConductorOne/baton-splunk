package connector

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-splunk/pkg/splunk"
)

const readPerm = "read"
const writePerm = "write"

type applicationResourceType struct {
	resourceType *v2.ResourceType
	client       *splunk.Client
}

func (a *applicationResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return a.resourceType
}

// applicationResource creates a new connector resource for a Splunk Application.
func applicationResource(ctx context.Context, application *splunk.Application, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	// get rid of leading url address in Id
	var applicationID string
	slashIndex := strings.LastIndex(application.Id, "/")
	if slashIndex != -1 {
		applicationID = application.Id[slashIndex+1:]
	} else {
		return nil, fmt.Errorf("splunk-connector: failed to parse application id: %s", application.Id)
	}

	displayName := titleCaser.String(application.Name)
	resource, err := rs.NewResource(
		displayName,
		resourceTypeApplication,
		applicationID,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (a *applicationResourceType) List(ctx context.Context, parentID *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentID != nil {
		return nil, "", nil, nil
	}

	bag, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: resourceTypeApplication.Id})
	if err != nil {
		return nil, "", nil, err
	}

	applications, nextPage, err := a.client.GetApplications(
		ctx,
		splunk.PaginationVars{
			Limit: ResourcesPageSize,
			Page:  bag.PageToken(),
		},
	)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to list applications: %w", err)
	}

	pageToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, 0, len(applications))
	for _, application := range applications {
		applicationCopy := application

		rr, err := applicationResource(ctx, &applicationCopy, parentID)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	return rv, pageToken, nil, nil
}

func (a *applicationResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	entitlementOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDescription(fmt.Sprintf("%s Splunk application", resource.DisplayName)),
	}

	rv = append(rv, ent.NewPermissionEntitlement(
		resource,
		readPerm,
		append(
			[]ent.EntitlementOption{ent.WithDisplayName(fmt.Sprintf("%s application READ", resource.DisplayName))},
			entitlementOptions...,
		)...,
	))
	rv = append(rv, ent.NewPermissionEntitlement(
		resource,
		writePerm,
		append(
			[]ent.EntitlementOption{ent.WithDisplayName(fmt.Sprintf("%s application WRITE", resource.DisplayName))},
			entitlementOptions...,
		)...,
	))

	return rv, "", nil, nil
}

func (a *applicationResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: resourceTypeUser.Id})
	if err != nil {
		return nil, "", nil, err
	}

	application, err := a.client.GetApplication(ctx, resource.Id.Resource)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to get application: %w", err)
	}

	applicationReadRoles, applicationWriteRoles := application.ACL.Perms.Read, application.ACL.Perms.Write

	var rv []*v2.Grant
	users, nextPage, err := a.client.GetUsers(
		ctx,
		splunk.PaginationVars{
			Limit: ResourcesPageSize,
			Page:  bag.PageToken(),
		},
	)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to get users: %w", err)
	}

	pageToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		userCopy := user

		ur, err := userResource(ctx, &userCopy, resource.ParentResourceId)
		if err != nil {
			return nil, "", nil, fmt.Errorf("splunk-connector: failed to build user resource: %w", err)
		}

		for _, role := range user.Content.Roles {
			if containsRole(applicationReadRoles, role) {
				rv = append(rv, grant.NewGrant(
					resource,
					readPerm,
					ur.Id,
				))
			}

			if containsRole(applicationWriteRoles, role) {
				rv = append(rv, grant.NewGrant(
					resource,
					writePerm,
					ur.Id,
				))
			}
		}
	}

	return rv, pageToken, nil, nil
}

func applicationBuilder(client *splunk.Client) *applicationResourceType {
	return &applicationResourceType{
		resourceType: resourceTypeApplication,
		client:       client,
	}
}
