---
page_title: "dtz_containers_service Resource - terraform-provider-dtz"
subcategory: ""
description: |-
  Manages a container service in the DownToZero.cloud service.
---

# dtz_containers_service (Resource)

The `dtz_containers_service` resource allows you to create, update, and delete container services in the DownToZero.cloud service.

## Example Usage

```terraform
# Example 1: Using container_image with tag
resource "dtz_containers_service" "service-with-tag" {
    prefix = "/my-service"
    container_image = "docker.io/library/nginx:latest"
    env_variables = {
        "KEY1" = "VALUE1"
    }
    login {
        provider_name = "github"
    }
}

# Example 2: Using container_image with digest
resource "dtz_containers_service" "service-with-digest" {
    prefix = "/my-service-digest"
    container_image = "docker.io/library/nginx@sha256:abc123def456789abcdef123456789abcdef123456789abcdef123456789abcd"
    env_variables = {
        "KEY1" = "VALUE1"
    }
    login {
        provider_name = "github"
    }
}

# Example 3: Using container_image without tag (automatically appends :latest)
resource "dtz_containers_service" "service-auto-latest" {
    prefix = "/my-auto-latest"
    container_image = "docker.io/library/nginx"  # Will become docker.io/library/nginx:latest
    env_variables = {
        "KEY1" = "VALUE1"
        "KEY2" = "VALUE2"
    }
    login {
        provider_name = "github"
    }
}

# Example 4: Minimal example without login block (login is optional)
resource "dtz_containers_service" "service-minimal" {
    prefix = "/minimal"
    container_image = "nginx"
}
```

## Schema

### Required

- `prefix` (String) A unique identifier for the service.
- `container_image` (String) The container image to use for the service. Can include:
  - **Tags**: `nginx:1.21`, `nginx:latest`
  - **Digests**: `nginx@sha256:abc123...`
  - **No tag/digest**: `nginx` (automatically becomes `nginx:latest`)

### Optional

- `container_image_version` (String) **DEPRECATED**: Include the tag or digest directly in the `container_image` field instead. This field is maintained for backward compatibility only.
- `container_pull_user` (String) Username for pulling the container image if it's in a private repository.
- `container_pull_pwd` (String, Sensitive) Password for pulling the container image if it's in a private repository.
- `env_variables` (Map of String) Environment variables to set in the container.
- `login` (Block) Login configuration for the service. Can only contain `provider_name`.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```

```