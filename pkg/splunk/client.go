package splunk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const BaseURL = "https://ms-work-g.local:8089/"
const UsersBaseURL = BaseURL + "/services/authentication/users"
const RolesBaseURL = BaseURL + "/services/authorization/roles"
const CapabilitiesBaseURL = BaseURL + "/services/authorization/capabilities"

type Client struct {
	httpClient *http.Client
	Token      string
}

type UsersResponse struct {
	Values []User `json:"entry"`
}

type RolesResponse struct {
	Values []Role `json:"entry"`
}

func NewClient(httpClient *http.Client, token string) *Client {
	return &Client{
		httpClient: httpClient,
		Token:      token,
	}
}

// GetUsers returns all users under specific Splunk instance.
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	var usersResponse UsersResponse

	err := c.doRequest(
		ctx,
		UsersBaseURL,
		&usersResponse,
	)

	if err != nil {
		return nil, err
	}

	return usersResponse.Values, nil
}

// GetRoles returns all roles under specific Splunk instance.
func (c *Client) GetRoles(ctx context.Context) ([]Role, error) {
	var rolesResponse RolesResponse

	err := c.doRequest(
		ctx,
		RolesBaseURL,
		&rolesResponse,
	)

	if err != nil {
		return nil, err
	}

	return rolesResponse.Values, nil
}

func setupQueryParams(query *url.Values) {
	// setup response format to JSON
	query.Set("output_mode", "json")
}

func (c *Client) doRequest(
	ctx context.Context,
	urlAddress string,
	resourceResponse interface{},
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlAddress, nil)
	if err != nil {
		return err
	}

	// setup query params
	queryParams := url.Values{}
	setupQueryParams(&queryParams)
	req.URL.RawQuery = queryParams.Encode()

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
