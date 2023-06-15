package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const BaseURL = "https://ms-work-g.local:8089/"
const UsersBaseURL = BaseURL + "/services/authentication/users"
const RolesBaseURL = BaseURL + "/services/authorization/roles"
const CapabilitiesBaseURL = BaseURL + "/services/authorization/capabilities"
const ApplicationsBaseURL = BaseURL + "/services/apps/local"

type Client struct {
	httpClient *http.Client
	Token      string
}

type PaginationData struct {
	Total   int `json:"total"`
	PerPage int `json:"perPage"`
	Offset  int `json:"offset"`
}

type PaginationVars struct {
	Limit int
	Page  string
}

type Response[T any] struct {
	Values         []T `json:"entry"`
	PaginationData `json:"paging"`
}

func NewClient(httpClient *http.Client, token string) *Client {
	return &Client{
		httpClient: httpClient,
		Token:      token,
	}
}

// GetUsers returns all users under specific Splunk instance.
func (c *Client) GetUsers(ctx context.Context, getUsersVars PaginationVars) ([]User, string, error) {
	var usersResponse Response[User]

	err := c.doRequest(
		ctx,
		UsersBaseURL,
		&usersResponse,
		&getUsersVars,
		"",
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&usersResponse)
}

// GetUsersByRole returns all users in some specific role under one Splunk instance.
func (c *Client) GetUsersByRole(ctx context.Context, getUsersVars PaginationVars, role string) ([]User, string, error) {
	var usersResponse Response[User]
	var roleFilter string

	if role != "" {
		roleFilter = fmt.Sprintf("roles=\"%s\"", role)
	}

	err := c.doRequest(
		ctx,
		UsersBaseURL,
		&usersResponse,
		&getUsersVars,
		roleFilter,
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&usersResponse)
}

// GetRoles returns all roles under specific Splunk instance.
func (c *Client) GetRoles(ctx context.Context, getRolesVars PaginationVars) ([]Role, string, error) {
	var rolesResponse Response[Role]

	err := c.doRequest(
		ctx,
		RolesBaseURL,
		&rolesResponse,
		&getRolesVars,
		"",
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&rolesResponse)
}

// GetApplications returns all applications under specific Splunk instance.
func (c *Client) GetApplications(ctx context.Context, getApplicationsVars PaginationVars) ([]Application, string, error) {
	var applicationsResponse Response[Application]

	err := c.doRequest(
		ctx,
		ApplicationsBaseURL,
		&applicationsResponse,
		&getApplicationsVars,
		"",
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&applicationsResponse)
}

// Handles pagination for Splunk API
// `offset` is 0-indexed representation of the current page,
// `perPage` is the number of items per page and
// `total` is the total number of items in the response.
func handlePagination[T any](response *Response[T]) ([]T, string, error) {
	total, perPage, offset := response.Total, response.PerPage, response.Offset+1

	if (offset * perPage) < total {
		return response.Values, strconv.Itoa(offset), nil
	}

	return response.Values, "", nil
}

func setupPagination(query *url.Values, limit int, page string) {
	// add limit
	if limit != 0 {
		query.Set("count", strconv.Itoa(limit))
	}

	// add page
	if page != "" {
		query.Set("offset", page)
	}
}

func setupFiltering(query *url.Values, filter string) {
	// add filter
	if filter != "" {
		query.Set("search", filter)
	}
}

func setupQueryParams(query *url.Values) {
	// setup response format to JSON
	query.Set("output_mode", "json")
}

func (c *Client) doRequest(
	ctx context.Context,
	urlAddress string,
	resourceResponse interface{},
	paginationVars *PaginationVars,
	filter string,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlAddress, nil)
	if err != nil {
		return err
	}

	// setup query params
	queryParams := url.Values{}
	setupQueryParams(&queryParams)
	setupPagination(&queryParams, paginationVars.Limit, paginationVars.Page)
	setupFiltering(&queryParams, filter)

	if queryParams != nil {
		req.URL.RawQuery = queryParams.Encode()
	}

	// setup headers
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	rawResponse, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer rawResponse.Body.Close()

	if rawResponse.StatusCode >= 300 {
		return status.Error(codes.Code(rawResponse.StatusCode), "Request failed")
	}

	if err := json.NewDecoder(rawResponse.Body).Decode(&resourceResponse); err != nil {
		return err
	}

	return nil
}
