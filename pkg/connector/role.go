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
	// get rid of leading url address in Id
	var roleID string
	slashIndex := strings.LastIndex(role.Id, "/")
	if slashIndex != -1 {
		roleID = role.Id[slashIndex+1:]
	} else {
		return nil, fmt.Errorf("splunk-connector: failed to parse role id: %s", role.Id)
	}

	displayName := titleCaser.String(role.Name)

	// merge role.capabilities and role.imported_capabilities and join into a string
	roleCapabilities := append([]string(nil), role.Content.Capabilities...)
	roleCapabilities = append(roleCapabilities, role.Content.ImportedCapabilities...)
	roleCapabilitiesString := strings.Join(roleCapabilities, ",")

	profile := map[string]interface{}{
		"role_id":           roleID,
		"role_name":         role.Name,
		"role_capabilities": roleCapabilitiesString,
	}

	resource, err := rs.NewRoleResource(
		displayName,
		resourceTypeRole,
		roleID,
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

	// add also role capabilities as entitlements
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	roleCapabilitiesPayload, ok := rs.GetProfileStringValue(roleTrait.Profile, "role_capabilities")
	if !ok {
		return nil, "", nil, fmt.Errorf("splunk-connector: error parsing role capabilities from role profile")
	}

	roleCapabilities := strings.Split(roleCapabilitiesPayload, ",")
	for _, roleCapability := range roleCapabilities {
		entitlementOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(resourceTypeRole),
			ent.WithDisplayName(fmt.Sprintf("%s capability", roleCapability)),
			ent.WithDescription(fmt.Sprintf("%s Splunk capability", roleCapability)),
		}

		rv = append(rv, ent.NewPermissionEntitlement(resource, roleCapability, entitlementOptions...))
	}

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

	roleCapabilitiesPayload, ok := rs.GetProfileStringValue(roleTrait.Profile, "role_capabilities")
	if !ok {
		return nil, "", nil, fmt.Errorf("splunk-connector: error parsing role capabilities from role profile")
	}

	roleCapabilities := strings.Split(roleCapabilitiesPayload, ",")
	var rv []*v2.Grant

	for _, roleCapability := range roleCapabilities {
		rv = append(rv, grant.NewGrant(
			resource,
			roleCapability,
			resource.Id,
		))
	}

	users, err := r.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to get users: %w", err)
	}

	for _, user := range users {
		userCopy := user

		ur, err := userResource(ctx, &userCopy)
		if err != nil {
			return nil, "", nil, fmt.Errorf("splunk-connector: failed to build user resource: %w", err)
		}

		for _, role := range user.Content.Roles {
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
