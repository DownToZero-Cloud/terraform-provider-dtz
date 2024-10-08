---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dtz_rss2email_feed Resource - terraform-provider-dtz"
subcategory: ""
description: |-
  Manages an RSS2Email feed.
---

# dtz_rss2email_feed (Resource)

The `dtz_rss2email_feed` resource allows you to create, update, and delete RSS2Email feeds in the DownToZero.cloud service.

## Example Usage

```terraform
resource "dtz_rss2email_feed" "example" {
  url = "https://example.com/rss-feed"
  enabled = true
}
```


## Schema

### Required

- `url` (String) The URL of the RSS feed.

### Optional

- `enabled` (Boolean) Whether the feed is enabled or not. Defaults to `false`.

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String) The name of the RSS feed.
- `last_check` (String) The timestamp of the last check performed on the feed.
- `last_data_found` (String) The timestamp when data was last found in the feed.

## Import

Import is supported using the following syntax:

```shell
terraform import dtz_rss2email_feed.example <feed_id>
```

