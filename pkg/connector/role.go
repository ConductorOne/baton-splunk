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

const roleMember = "member"

type roleResourceType struct {
	resourceType *v2.ResourceType
	client       *splunk.Client
}

func (r *roleResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return r.resourceType
}

// roleResource creates a new connector resource for a Splunk Role.
func roleResource(ctx context.Context, role *splunk.Role) (*v2.Resource, error) {
	displayName := titleCaser.String(role.Name)
	profile := map[string]interface{}{
		"role_id":   role.Id,
		"role_name": role.Name,
	}

	resource, err := rs.NewRoleResource(
		displayName,
		resourceTypeRole,
		role.Id,
		[]rs.RoleTraitOption{rs.WithRoleProfile(profile)},
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r *roleResourceType) List(ctx context.Context, parentID *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	roles, err := r.client.GetRoles(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to list roles: %w", err)
	}

	rv := make([]*v2.Resource, 0, len(roles))
	for _, role := range roles {
		roleCopy := role

		rr, err := roleResource(ctx, &roleCopy)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	return rv, "", nil, nil
}

func (r *roleResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	entitlementOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDisplayName(fmt.Sprintf("%s role", resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("%s Splunk role", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, roleMember, entitlementOptions...))

	return rv, "", nil, nil
}

func (r *roleResourceType) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	roleName, ok := rs.GetProfileStringValue(roleTrait.Profile, "role_name")
	if !ok {
		return nil, "", nil, fmt.Errorf("splunk-connector: error parsing role name from role profile")
	}

	users, err := r.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to get users: %w", err)
	}

	var rv []*v2.Grant
	for _, user := range users {
		// log the user
		fmt.Printf("api user: %s\n", user.Name)

		userCopy := user

		ur, err := userResource(ctx, &userCopy)
		if err != nil {
			return nil, "", nil, fmt.Errorf("splunk-connector: failed to build user resource: %w", err)
		}

		for _, role := range user.Content.Roles {

			// log the roles
			fmt.Printf("api role: %s\nresource role: %s\n", role, roleName)

			if role == roleName {
				rv = append(rv, grant.NewGrant(
					resource,
					roleMember,
					ur.Id,
				))
			}
		}
	}

	return rv, "", nil, nil
}

func roleBuilder(client *splunk.Client) *roleResourceType {
	return &roleResourceType{
		resourceType: resourceTypeRole,
		client:       client,
	}
}
