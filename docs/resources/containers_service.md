---
page_title: "dtz_containers_service Resource - terraform-provider-dtz"
subcategory: ""
description: |-
  Manages a container service in the DownToZero.cloud platform.
---

# dtz_containers_service (Resource)

Creates and manages a container service that runs your Docker container on the DownToZero.cloud platform.

## Example Usage

```terraform
# Basic web service (no authentication required)
resource "dtz_containers_service" "web_app" {
  prefix          = "/api"
  container_image = "nginx:alpine"
}

# Application with environment variables (no authentication)
resource "dtz_containers_service" "app" {
  prefix          = "/app"
  container_image = "myregistry.com/myapp:v1.2.3"
  
  env_variables = {
    PORT        = "8080"
    DATABASE_URL = "postgres://..."
    API_KEY     = var.api_key
  }
}

# Private registry with authentication
resource "dtz_containers_service" "private_app" {
  prefix              = "/private"
  container_image     = "private-registry.com/app:latest"
  container_pull_user = "registry-user"
  container_pull_pwd  = var.registry_password
  
  login = {
    provider_name = "dtz"
  }
}

# Using specific digest for immutable deployments
resource "dtz_containers_service" "production" {
  prefix          = "/prod"
  container_image = "myapp@sha256:a1b2c3d4e5f6789..."
  
  env_variables = {
    ENV = "production"
  }
  
  login = {
    provider_name = "dtz"
  }
}
```

## Schema

### Required

- `prefix` (String) The URL path prefix for your service (e.g., `/api`, `/app`). Must be unique within your context.
- `container_image` (String) The Docker image to run. Must include a tag or a digest:
  - **With tag**: `nginx:1.21` or `myregistry.com/app:v2.0`
  - **With digest**: `nginx@sha256:abc123...` (recommended for production)

### Optional

- `container_pull_user` (String) Username for authenticating with private container registries.
- `container_pull_pwd` (String, Sensitive) Password for authenticating with private container registries.
- `env_variables` (Map of String) Environment variables passed to the container at runtime.
- `login` (Object, Optional) Enables DTZ authentication for the service. If provided, must contain:
  - `provider_name` (String, Required) Must be `"dtz"` (only supported provider).

### Read-Only

- `id` (String) The unique identifier of the service.
- `container_image_version` (String) Computed output. Use the tag or digest directly in `container_image` instead.

## Argument Reference

### Container Image Validation

The `container_image` must include either a tag (e.g., `:1.2` or `:latest`) or a digest (e.g., `@sha256:...`).

### Private Registry Authentication

For private registries, provide both `container_pull_user` and `container_pull_pwd`:

```terraform
resource "dtz_containers_service" "private" {
  prefix              = "/app"
  container_image     = "private.registry.com/app:latest"
  container_pull_user = "username"
  container_pull_pwd  = var.registry_password
}
```

### DTZ Authentication (Login Attribute)

The `login` attribute is **optional** and can be used in two ways:

- **No login attribute**: Service is publicly accessible
- **Login attribute with provider_name = "dtz"**: Service requires DTZ authentication to access

```terraform
# Public service (no login attribute)
resource "dtz_containers_service" "public_api" {
  prefix          = "/public"
  container_image = "my-public-api:latest"
}

# Authenticated service (login attribute with provider_name)
resource "dtz_containers_service" "private_api" {
  prefix          = "/private"
  container_image = "my-private-api:latest"
  
  login = {
    provider_name = "dtz"
  }
}
```

## Import

Services can be imported using their service ID:

```shell
terraform import dtz_containers_service.example <service_id>
```

Find your service ID in the DTZ dashboard or via the API.