---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dtz_identity_apikey Resource - terraform-provider-dtz"
subcategory: ""
description: |-
  Manages an API key in the DownToZero.cloud service.
---

# dtz_identity_apikey (Resource)

The `dtz_identity_apikey` resource allows you to create, update, and delete API keys in the DownToZero.cloud service.

## Example Usage

```terraform
resource "dtz_identity_apikey" "example" {
  alias = "my-api-key"
  context_id = "my-api-key-context"
}
```

## Schema

### Required

- `context_id` (String) The context ID of the API key.

### Optional

- `alias` (String) The alias of the API key.

### Read-Only

- `apikey` (String) The API key.

## Import

Import is supported using the following syntax:

```shell
terraform import dtz_identity_apikey.example <apikey_id>
```