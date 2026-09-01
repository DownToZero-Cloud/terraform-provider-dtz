package provider

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// closeResponseBody closes an HTTP response body and logs any error that occurs
// while closing. It is meant to be used directly in a defer statement.
func closeResponseBody(ctx context.Context, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		tflog.Error(ctx, "error closing response body", map[string]interface{}{
			"error": err,
		})
	}
}
