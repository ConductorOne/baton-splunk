package splunk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const Localhost = "localhost"
const BaseURL = "https://%s:8089/"
const CloudBaseURL = "https://%s.splunkcloud.com:8089/"

const UsersBaseURL = "/services/authentication/users"
const UserBaseURL = "/services/authentication/users/%s"
const RolesBaseURL = "/services/authorization/roles"
const CapabilitiesBaseURL = "/services/authorization/grantable_capabilities/capabilities"
const ApplicationsBaseURL = "/services/apps/local"
const ApplicationBaseURL = "/services/apps/local/%s"

type Client struct {
	httpClient *http.Client
	Auth       string
	Cloud      bool
	Deployment string
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

func NewClient(httpClient *http.Client, auth string, cloud bool) *Client {
	return &Client{
		httpClient: httpClient,
		Auth:       auth,
		Cloud:      cloud,
		Deployment: Localhost,
	}
}

func (c *Client) PointToDeployment(deployment string) {
	c.Deployment = deployment
}

func (c *Client) PointToLocalhost() {
	c.Deployment = Localhost
}

func (c *Client) ResetPointer() {
	c.Deployment = ""
}

// GetUrl returns the full URL for the given endpoint based on platform.
func (c *Client) CreateUrl(endpoint string) string {
	if c.Cloud {
		return fmt.Sprintf(CloudBaseURL, c.Deployment) + endpoint
	} else {
		return fmt.Sprintf(BaseURL, c.Deployment) + endpoint
	}
}

func (c *Client) IsCloudPlatform() bool {
	return c.Cloud
}

// GetUsers returns all users under specific Splunk instance.
func (c *Client) GetUsers(ctx context.Context, getUsersVars PaginationVars) ([]User, string, error) {
	var usersResponse Response[User]

	err := c.get(
		ctx,
		c.CreateUrl(UsersBaseURL),
		&usersResponse,
		&getUsersVars,
		"",
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&usersResponse)
}

// GetUser returns information regarding one specific user under Splunk instance.
func (c *Client) GetUser(ctx context.Context, userId string) (*User, error) {
	var usersResponse Response[User]

	err := c.get(
		ctx,
		c.CreateUrl(fmt.Sprintf(UserBaseURL, userId)),
		&usersResponse,
		nil,
		"",
	)

	if err != nil {
		return nil, err
	}

	return &usersResponse.Values[0], nil
}

// GetUsersByRole returns all users in some specific role under one Splunk instance.
func (c *Client) GetUsersByRole(ctx context.Context, getUsersVars PaginationVars, role string) ([]User, string, error) {
	var usersResponse Response[User]
	var roleFilter string

	if role != "" {
		roleFilter = fmt.Sprintf("roles=\"%s\"", role)
	}

	err := c.get(
		ctx,
		c.CreateUrl(UsersBaseURL),
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

	err := c.get(
		ctx,
		c.CreateUrl(RolesBaseURL),
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

	err := c.get(
		ctx,
		c.CreateUrl(ApplicationsBaseURL),
		&applicationsResponse,
		&getApplicationsVars,
		"",
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&applicationsResponse)
}

// GetApplication returns specific application under Splunk instance.
func (c *Client) GetApplication(ctx context.Context, applicationName string) (*Application, error) {
	var applicationResponse Response[Application]

	err := c.get(
		ctx,
		c.CreateUrl(fmt.Sprintf(ApplicationBaseURL, applicationName)),
		&applicationResponse,
		nil,
		"",
	)

	if err != nil {
		return nil, err
	}

	return &applicationResponse.Values[0], nil
}

// GetCapabilities returns all grantable capabilities under specific Splunk instance.
func (c *Client) GetCapabilities(ctx context.Context, getCapabilitiesVars PaginationVars) ([]Capability, string, error) {
	var capabilitiesResponse Response[Capability]

	err := c.get(
		ctx,
		c.CreateUrl(CapabilitiesBaseURL),
		&capabilitiesResponse,
		&getCapabilitiesVars,
		"",
	)

	if err != nil {
		return nil, "", err
	}

	return handlePagination(&capabilitiesResponse)
}

// UpdateUser updates specific user under Splunk instance.
func (c *Client) UpdateUser(ctx context.Context, userId string, roles []string) error {
	data := url.Values{}

	for _, role := range roles {
		data.Add("roles", role)
	}

	err := c.post(
		ctx,
		c.CreateUrl(fmt.Sprintf(UserBaseURL, userId)),
		data,
		"",
	)

	if err != nil {
		return err
	}

	return nil
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

func setupPagination(query *url.Values, paginationVars *PaginationVars) {
	if paginationVars == nil {
		return
	}

	// add limit
	if paginationVars.Limit != 0 {
		query.Set("count", strconv.Itoa(paginationVars.Limit))
	}

	// add page
	if paginationVars.Page != "" {
		query.Set("offset", paginationVars.Page)
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

func (c *Client) get(
	ctx context.Context,
	urlAddress string,
	resourceResponse interface{},
	paginationVars *PaginationVars,
	filter string,
) error {
	return c.doRequest(ctx, http.MethodGet, urlAddress, nil, resourceResponse, paginationVars, filter)
}

func (c *Client) post(
	ctx context.Context,
	urlAddress string,
	data url.Values,
	filter string,
) error {
	return c.doRequest(ctx, http.MethodPost, urlAddress, data, nil, nil, filter)
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	urlAddress string,
	data url.Values,
	resourceResponse interface{},
	paginationVars *PaginationVars,
	filter string,
) error {
	var body strings.Reader

	if data != nil {
		encodedData := data.Encode()
		bodyReader := strings.NewReader(encodedData)
		body = *bodyReader
	}

	req, err := http.NewRequestWithContext(ctx, method, urlAddress, &body)
	if err != nil {
		return err
	}

	// setup query params
	queryParams := url.Values{}
	setupQueryParams(&queryParams)
	setupPagination(&queryParams, paginationVars)
	setupFiltering(&queryParams, filter)

	if queryParams != nil {
		req.URL.RawQuery = queryParams.Encode()
	}

	// setup headers
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", c.Auth)

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
