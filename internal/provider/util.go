package provider

import (
	"context"
	"io"

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