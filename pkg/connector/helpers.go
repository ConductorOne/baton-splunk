package connector

import (
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const ResourcesPageSize = 50

var titleCaser = cases.Title(language.English)

func annotationsForUserResourceType() annotations.Annotations {
	annos := annotations.Annotations{}
	annos.Update(&v2.SkipEntitlementsAndGrants{})
	return annos
}

func parsePageToken(i string, resourceID *v2.ResourceId) (*pagination.Bag, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(i)
	if err != nil {
		return nil, err
	}

	if b.Current() == nil {
		b.Push(pagination.PageState{
			ResourceTypeID: resourceID.ResourceType,
			ResourceID:     resourceID.Resource,
		})
	}

	return b, nil
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == "*" {
			return true
		}

		if r == role {
			return true
		}
	}

	return false
}

func removeLeadingUrl(url string) (string, error) {
	slashIndex := strings.LastIndex(url, "/")

	if slashIndex != -1 {
		return url[slashIndex+1:], nil
	}

	return "", fmt.Errorf("failed to parse resource id: %s", url)
}

// Id of entitlement has following format <resource_type>:<resource_id>:<entitlement_id>
// extract resource_id from it.
func extractResourceId(fullId string) (string, error) {
	idParts := strings.Split(fullId, ":")

	if len(idParts) != 3 {
		return "", fmt.Errorf("invalid resource id: %s", fullId)
	}

	return idParts[1], nil
}

func extractEntitlementId(fullId string) (string, error) {
	idParts := strings.Split(fullId, ":")

	if len(idParts) != 3 {
		return "", fmt.Errorf("invalid resource id: %s", fullId)
	}

	return idParts[2], nil
}

func removeResource(resources []string, targetResource string) []string {
	var nr []string

	for _, r := range resources {
		if r != targetResource {
			nr = append(nr, r)
		}
	}

	return nr
}

func isResourcePresent(resources []string, targetResource string) bool {
	for _, r := range resources {
		if r == targetResource {
			return true
		}
	}

	return false
}
