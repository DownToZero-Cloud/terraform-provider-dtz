package provider

import (
	"context"
	"fmt"
	"io"
	"regexp"
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

// validateContainerImage validates that the container image does not contain tags or digests
// Valid: docker.io/library/nginx, nginx, myregistry.com/myimage
// Invalid: nginx:latest, nginx:1.0, nginx@sha256:abc123
func validateContainerImage(image string) error {
	if image == "" {
		return nil // empty is allowed (will be caught by required validation)
	}

	// Check for tag (contains :)
	if strings.Contains(image, ":") {
		return fmt.Errorf("container_image must not contain tags (found ':'). Use container_image_version field for specifying versions")
	}

	// Check for digest (contains @)
	if strings.Contains(image, "@") {
		return fmt.Errorf("container_image must not contain digests (found '@'). Use container_image_version field for specifying versions")
	}

	return nil
}

// validateContainerImageVersion validates that the container image version contains only digests and not tags
// Valid: @sha256:abc123def456..., @sha1:abc123
// Invalid: latest, 1.0, v1.2.3
func validateContainerImageVersion(version string) error {
	if version == "" {
		return nil // empty is allowed (optional field)
	}

	// Must start with @ to be a digest
	if !strings.HasPrefix(version, "@") {
		return fmt.Errorf("container_image_version must be a digest starting with '@' (e.g., @sha256:abc123). Tags like 'latest' or '1.0' are not allowed")
	}

	// Validate digest format: @<algorithm>:<hex>
	digestPattern := `^@[a-z0-9]+:[a-f0-9]+$`
	matched, err := regexp.MatchString(digestPattern, version)
	if err != nil {
		return fmt.Errorf("error validating container_image_version format: %v", err)
	}
	if !matched {
		return fmt.Errorf("container_image_version must be in digest format @<algorithm>:<hex> (e.g., @sha256:abc123)")
	}

	return nil
}
