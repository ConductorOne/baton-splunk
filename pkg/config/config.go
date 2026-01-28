package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	Token = field.StringField(
		"token",
		field.WithDescription("The Splunk access token used to connect to the Splunk API."),
		field.WithIsSecret(true),
	)

	Username = field.StringField(
		"username",
		field.WithDescription("Username of user used to connect to the Splunk API."),
	)

	Password = field.StringField(
		"password",
		field.WithDescription("Password of user used to connect to the Splunk API."),
		field.WithIsSecret(true),
	)

	Unsafe = field.BoolField(
		"unsafe",
		field.WithDescription("Allow insecure TLS connections to Splunk."),
	)

	Verbose = field.BoolField(
		"verbose",
		field.WithDescription("Enable listing verbose entitlements for Role capabilities."),
	)

	Cloud = field.BoolField(
		"cloud",
		field.WithDescription("Switches to cloud API endpoints."),
	)

	Deployments = field.StringSliceField(
		"deployments",
		field.WithDescription("Limit syncing to specific deployments by specifying cloud deployment names or IP addresses of on-premise deployments."),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration([]field.SchemaField{
	Token,
	Username,
	Password,
	Unsafe,
	Verbose,
	Cloud,
	Deployments,
})
