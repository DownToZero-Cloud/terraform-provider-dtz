package provider

import (
	"context"
	"io"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// deferredCloseResponseBody creates a deferred function to close an HTTP response body
// and log any errors that occur during closing
func deferredCloseResponseBody(ctx context.Context, body io.ReadCloser) func() {
	return func() {
		if err := body.Close(); err != nil {
			tflog.Error(ctx, "error closing response body", map[string]interface{}{
				"error": err,
			})
		}
	}
}

// normalizeContainerImage ensures that a container image has a tag or digest,
// appending :latest if none is specified
func normalizeContainerImage(image string) string {
	if image == "" {
		return image
	}

	// Check if image already has a tag (contains :) or digest (contains @)
	if strings.Contains(image, ":") || strings.Contains(image, "@") {
		return image
	}

	// No tag or digest found, append :latest
	return image + ":latest"
}
