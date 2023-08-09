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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
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
func roleResource(ctx context.Context, role *splunk.Role, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	roleID, err := removeLeadingUrl(role.Id)
	if err != nil {
		return nil, fmt.Errorf("splunk-connector: %w", err)
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

	resource, err := rs.NewGroupResource(
		displayName,
		resourceTypeRole,
		roleID,
		[]rs.GroupTraitOption{rs.WithGroupProfile(profile)},
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r *roleResourceType) List(ctx context.Context, parentID *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentID != nil {
		return nil, "", nil, nil
	}

	bag, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: resourceTypeRole.Id})
	if err != nil {
		return nil, "", nil, err
	}

	roles, nextPage, err := r.client.GetRoles(
		ctx,
		splunk.PaginationVars{
			Limit: ResourcesPageSize,
			Page:  bag.PageToken(),
		},
	)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to list roles: %w", err)
	}

	pageToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, 0, len(roles))
	for _, role := range roles {
		roleCopy := role

		rr, err := roleResource(ctx, &roleCopy, parentID)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	return rv, pageToken, nil, nil
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

func (r *roleResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: resourceTypeUser.Id})
	if err != nil {
		return nil, "", nil, err
	}

	roleTrait, err := rs.GetGroupTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	roleName, ok := rs.GetProfileStringValue(roleTrait.Profile, "role_name")
	if !ok {
		return nil, "", nil, fmt.Errorf("splunk-connector: error parsing role name from role profile")
	}

	users, nextPage, err := r.client.GetUsersByRole(
		ctx,
		splunk.PaginationVars{
			Limit: ResourcesPageSize,
			Page:  bag.PageToken(),
		},
		roleName,
	)
	if err != nil {
		return nil, "", nil, fmt.Errorf("splunk-connector: failed to get users: %w", err)
	}

	pageToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	for _, user := range users {
		userCopy := user

		ur, err := userResource(ctx, &userCopy, resource.ParentResourceId)
		if err != nil {
			return nil, "", nil, fmt.Errorf("splunk-connector: failed to build user resource: %w", err)
		}

		rv = append(rv, grant.NewGrant(
			resource,
			roleMember,
			ur.Id,
		))
	}

	return rv, pageToken, nil, nil
}

func (r *roleResourceType) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if principal.Id.ResourceType != resourceTypeUser.Id {
		l.Warn(
			"splunk-connector: only users can be granted role membership",
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)

		return nil, fmt.Errorf("splunk-connector: only users can be granted role membership")
	}

	roleId := entitlement.Resource.Id.Resource

	// get existing roles under user
	user, err := r.client.GetUser(ctx, principal.Id.Resource)
	if err != nil {
		return nil, fmt.Errorf("splunk-connector: failed to find user: %w", err)
	}

	// check if role is already granted
	if isResourcePresent(user.Content.Roles, roleId) {
		return nil, fmt.Errorf("splunk-connector: role %s already granted to user", roleId)
	}

	// merge new role into existing roles
	user.Content.Roles = append(user.Content.Roles, roleId)

	// grant role membership
	err = r.client.UpdateUserRoles(ctx, principal.Id.Resource, user.Content.Roles)
	if err != nil {
		return nil, fmt.Errorf("splunk-connector: failed to grant role membership: %w", err)
	}

	return nil, nil
}

func (r *roleResourceType) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	entitlement := grant.Entitlement
	principal := grant.Principal

	if principal.Id.ResourceType != resourceTypeUser.Id {
		l.Warn(
			"splunk-connector: only users can have role membership revoked",
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)

		return nil, fmt.Errorf("splunk-connector: only users can have role membership revoked")
	}

	roleId := entitlement.Resource.Id.Resource

	// get existing roles under user
	user, err := r.client.GetUser(ctx, principal.Id.Resource)
	if err != nil {
		return nil, fmt.Errorf("splunk-connector: failed to find user: %w", err)
	}

	// check if role is present in user's roles
	if !isResourcePresent(user.Content.Roles, roleId) {
		return nil, fmt.Errorf("splunk-connector: role %s not present in user's roles", roleId)
	}

	// remove new role from existing roles
	user.Content.Roles = removeResource(user.Content.Roles, roleId)

	// revoke role membership
	err = r.client.UpdateUserRoles(ctx, principal.Id.Resource, user.Content.Roles)
	if err != nil {
		return nil, fmt.Errorf("splunk-connector: failed to revoke role membership: %w", err)
	}

	return nil, nil
}

func roleBuilder(client *splunk.Client) *roleResourceType {
	return &roleResourceType{
		resourceType: resourceTypeRole,
		client:       client,
	}
}
