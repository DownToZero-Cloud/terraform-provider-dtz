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

	// Handle the case where there's no '/' (just an image name)
	if !strings.Contains(image, "/") {
		// Check if it already has a tag or digest
		if strings.Contains(image, ":") || strings.Contains(image, "@") {
			return image
		}
		return image + ":latest"
	}

	// Split by '/' to separate registry from image path
	parts := strings.Split(image, "/")

	// If we have more than 2 parts, the first part might be a registry with port
	// We need to check if the first part contains a port (colon followed by digits)
	registryPart := parts[0]
	imagePath := strings.Join(parts[1:], "/")

	// Check if the registry part contains a port (colon followed by digits)
	if strings.Contains(registryPart, ":") {
		// Extract the port part after the colon
		colonIndex := strings.Index(registryPart, ":")
		portPart := registryPart[colonIndex+1:]

		// Check if the port part is numeric (simple check - just digits)
		// Also ensure the port part is not empty
		isNumericPort := portPart != ""
		for _, char := range portPart {
			if char < '0' || char > '9' {
				isNumericPort = false
				break
			}
		}

		// If it's a numeric port, we need to check the image path for tags
		if isNumericPort {
			if strings.Contains(imagePath, ":") || strings.Contains(imagePath, "@") {
				return image
			}
			return image + ":latest"
		}
	}

	// For cases without ports or with non-numeric "ports", check the last part
	lastPart := parts[len(parts)-1]
	if strings.Contains(lastPart, ":") || strings.Contains(lastPart, "@") {
		return image
	}

	// No tag or digest found, append :latest
	return image + ":latest"
}
