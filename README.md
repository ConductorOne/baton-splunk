![Baton Logo](./docs/images/baton-logo.png)

# `baton-splunk` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-splunk.svg)](https://pkg.go.dev/github.com/conductorone/baton-splunk) ![main ci](https://github.com/conductorone/baton-splunk/actions/workflows/main.yaml/badge.svg)

`baton-splunk` is a connector for Splunk built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the Splunk API to sync data about applications, users, roles and their capabilities.

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Prerequisites

Gaining access to the Splunk API varies based on platform, Enterprise or Cloud. 

## Splunk API Access
### Splunk Cloud

In case of Splunk Cloud, you need to create a new instance and white list ip addresses of machines where this connector will be running or submit a support case. For more information, see [here](https://docs.splunk.com/Documentation/SplunkCloud/9.0.2303/RESTTUT/RESTandCloud). 


### Splunk Enterprise

Enterprise version of Splunk comes with admin account and REST API access right away but requires more actions to prepare the instance for the connector locally. There are multiple supported platforms, look through left sidebar [here](https://docs.splunk.com/Documentation/Splunk/9.0.5/Installation/Whatsinthismanual). You can also use Splunk docker image, which simplifies the process.

## Credentials

To access the API, you can either use Basic authentication, using username and password you use to login to web view, or you can obtain API access token. After you log in to Splunk, go to the top menu bar, select `Settings` -> `Tokens` under Users and Authentication and create a new token with button `New Token`. Be aware that to sync all the users, roles and capabilities associated with them, you have to have necessary permissions.

# Getting Started

As mentioned above, you can use cloud or on-premise platform to run the connector on. In case of on-premise platform, you have to prepare the Splunk instance for the connector. Splunk docker image is the easiest way to do so.

The instance comes by default with SSL disabled, so to bypass validation of SSL certificates you have to set `BATON_UNSAFE` environment variable to `true` or use `--unsafe` flag.

To gain more verbose output, you can set `BATON_VERBOSE` environment variable to `true` or use `--verbose` flag. This mode includes listing of Application and Capability entitlements and grants.

In case you want to sync multiple deployments, you can set `BATON_DEPLOYMENTS` environment variable or use `--deployments` flag. You can specify multiple deployments by separating them with comma. You can specify deployments by their name or IP address. If you don't specify any deployment, the connector will sync only the localhost deployment. This flag is required for syncing cloud deployments (when `BATON_CLOUD` is set to `true`).

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-splunk

BATON_TOKEN=token baton-splunk
baton resources
```

## docker

### Limitations
Splunk docker image supports only `x86_64` CPU architecture. 

### With Splunk image
```
SPLUNK_ADMIN_PASSWORD=admin_pass BATON_TOKEN=token BATON_UNSAFE=true docker-compose up
```

### Without Splunk image
```
docker run --rm -v $(pwd):/out -e BATON_TOKEN=token BATON_UNSAFE=true ghcr.io/conductorone/baton-splunk:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-splunk/cmd/baton-splunk@main

BATON_TOKEN=token baton-splunk
baton resources
```

# Data Model

`baton-splunk` will fetch information about the following Splunk resources:

- Deployments
- Users
- Roles
- Capabilities
- Applications

By default, `baton-splunk` will sync information only from account based on provided credential and from deployments based on provided flag.

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-splunk` Command Line Usage

```
baton-splunk

Usage:
  baton-splunk [flags]
  baton-splunk [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --client-id string              The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string          The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --cloud                         Switches to cloud API endpoints. ($BATON_CLOUD)
      --deployments strings           Limit syncing to specific deployments by specifying cloud deployment names or IP addresses of on-premise deployments. ($BATON_DEPLOYMENTS)
  -f, --file string                   The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
      --grant-entitlement string      The entitlement to grant to the supplied principal ($BATON_GRANT_ENTITLEMENT)
      --grant-principal string        The resource to grant the entitlement to ($BATON_GRANT_PRINCIPAL)
      --grant-principal-type string   The resource type of the principal to grant the entitlement to ($BATON_GRANT_PRINCIPAL_TYPE)
  -h, --help                          help for baton-splunk
      --log-format string             The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string              The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --password string               Password of user used to connect to the Splunk API. ($BATON_PASSWORD)
      --revoke-grant string           The grant to revoke ($BATON_REVOKE_GRANT)
      --token string                  The Splunk access token used to connect to the Splunk API. ($BATON_TOKEN)
      --unsafe                        Allow insecure TLS connections to Splunk. ($BATON_UNSAFE)
      --username string               Username of user used to connect to the Splunk API. ($BATON_USERNAME)
      --verbose                       Enable listing verbose entitlements for Role capabilities. ($BATON_VERBOSE)
  -v, --version                       version for baton-splunk

Use "baton-splunk [command] --help" for more information about a command.

```
