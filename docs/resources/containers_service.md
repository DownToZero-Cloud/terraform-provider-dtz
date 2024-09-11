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
resource "dtz_containers_service" "my-service" {
    prefix = "/whatever"
    container_image = "docker.io/library/nginx"
    container_image_version = "latest"
    env_variables = {
        "KEY1" = "VALUE1"
        "KEY2" = "VALUE2"
    }
    login {
        provider_name = "github"
    }
}
```

## Schema

### Required

- `prefix` (String) A unique identifier for the service.
- `container_image` (String) The container image to use for the service.
- `container_image_version` (String) The version of the container image to use.

### Optional

- `container_pull_user` (String) Username for pulling the container image if it's in a private repository.
- `container_pull_pwd` (String, Sensitive) Password for pulling the container image if it's in a private repository.
- `env_variables` (Map of String) Environment variables to set in the container.
- `login` (Block) Login configuration for the service. Can only contain `provider_name`.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:
