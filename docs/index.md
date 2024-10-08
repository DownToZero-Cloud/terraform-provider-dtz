---
page_title: "Provider: DownToZero Provider"
subcategory: ""
description: |-
  The DownToZero (dtz) provider is used to manage resources on the DownToZero.cloud platform.
---

# DownToZero Provider

The dtz provider allows you to manage various resources and services on the [DownToZero.cloud](https://downtozero.cloud) platform. It provides support for containers, object storage, container registry, RSS2Email, and observability services.

## Example Usage

```terraform
terraform {
  required_providers {
    dtz = {
      source = "DownToZero-Cloud/dtz"
      version = "~> 0.1.25"
    }
  }
}

provider "dtz" {
    api_key = var.dtz_api_key
    enable_service_containers = true
    enable_service_rss2email = true
}
```

## Schema

### Required

- `api_key` (String, Sensitive) The API key for authentication

### Optional

- `enable_service_containers` (Boolean) Enable the containers service. Defaults to `false`.
- `enable_service_objectstore` (Boolean) Enable the object store service. Defaults to `false`.
- `enable_service_containerregistry` (Boolean) Enable the container registry service. Defaults to `false`.
- `enable_service_rss2email` (Boolean) Enable the RSS2Email service. Defaults to `false`.
- `enable_service_observability` (Boolean) Enable the observability service. Defaults to `false`.
