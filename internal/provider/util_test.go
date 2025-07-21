package provider

import (
	"testing"
)

func TestNormalizeContainerImage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple image without tag",
			input:    "nginx",
			expected: "nginx:latest",
		},
		{
			name:     "simple image with tag",
			input:    "nginx:1.21",
			expected: "nginx:1.21",
		},
		{
			name:     "simple image with digest",
			input:    "nginx@sha256:abc123",
			expected: "nginx@sha256:abc123",
		},
		{
			name:     "registry image without tag",
			input:    "docker.io/library/nginx",
			expected: "docker.io/library/nginx:latest",
		},
		{
			name:     "registry image with tag",
			input:    "docker.io/library/nginx:1.21",
			expected: "docker.io/library/nginx:1.21",
		},
		{
			name:     "registry with port without tag",
			input:    "localhost:5000/myimage",
			expected: "localhost:5000/myimage:latest",
		},
		{
			name:     "registry with port with tag",
			input:    "localhost:5000/myimage:v1.0",
			expected: "localhost:5000/myimage:v1.0",
		},
		{
			name:     "registry with port with digest",
			input:    "localhost:5000/myimage@sha256:abc123",
			expected: "localhost:5000/myimage@sha256:abc123",
		},
		{
			name:     "registry with port and path without tag",
			input:    "localhost:5000/project/myimage",
			expected: "localhost:5000/project/myimage:latest",
		},
		{
			name:     "registry with port and path with tag",
			input:    "localhost:5000/project/myimage:v1.0",
			expected: "localhost:5000/project/myimage:v1.0",
		},
		{
			name:     "registry with non-numeric port-like suffix",
			input:    "myregistry:latest/myimage",
			expected: "myregistry:latest/myimage:latest",
		},
		{
			name:     "registry with non-numeric port-like suffix and image tag",
			input:    "myregistry:latest/myimage:v1.0",
			expected: "myregistry:latest/myimage:v1.0",
		},
		// Bug fix test cases - empty ports
		{
			name:     "registry with empty port without tag",
			input:    "registry:/myimage",
			expected: "registry:/myimage:latest",
		},
		{
			name:     "registry with empty port with tag",
			input:    "registry:/myimage:v1.0",
			expected: "registry:/myimage:v1.0",
		},
		{
			name:     "registry with empty port with digest",
			input:    "registry:/myimage@sha256:abc123",
			expected: "registry:/myimage@sha256:abc123",
		},
		{
			name:     "localhost with empty port without tag",
			input:    "localhost:/myimage",
			expected: "localhost:/myimage:latest",
		},
		{
			name:     "localhost with empty port with tag",
			input:    "localhost:/myimage:v1.0",
			expected: "localhost:/myimage:v1.0",
		},
		{
			name:     "localhost with empty port with digest",
			input:    "localhost:/myimage@sha256:def456",
			expected: "localhost:/myimage@sha256:def456",
		},
		// Additional edge cases for robustness
		{
			name:     "registry with non-numeric port",
			input:    "registry:abc/myimage",
			expected: "registry:abc/myimage:latest",
		},
		{
			name:     "registry with non-numeric port and tag",
			input:    "registry:abc/myimage:v2.0",
			expected: "registry:abc/myimage:v2.0",
		},
		{
			name:     "registry with mixed port (starts numeric)",
			input:    "registry:123abc/myimage",
			expected: "registry:123abc/myimage:latest",
		},
		{
			name:     "registry with mixed port (starts non-numeric)",
			input:    "registry:abc123/myimage",
			expected: "registry:abc123/myimage:latest",
		},
		{
			name:     "complex registry path without tag",
			input:    "gcr.io/myproject/subproject/myimage",
			expected: "gcr.io/myproject/subproject/myimage:latest",
		},
		{
			name:     "complex registry path with tag",
			input:    "gcr.io/myproject/subproject/myimage:v1.0",
			expected: "gcr.io/myproject/subproject/myimage:v1.0",
		},
		{
			name:     "complex registry path with port without tag",
			input:    "myregistry.com:8080/namespace/project/app",
			expected: "myregistry.com:8080/namespace/project/app:latest",
		},
		{
			name:     "complex registry path with port with tag",
			input:    "myregistry.com:8080/namespace/project/app:v1.2.3",
			expected: "myregistry.com:8080/namespace/project/app:v1.2.3",
		},
		{
			name:     "complex registry path with empty port",
			input:    "myregistry.com:/namespace/project/app",
			expected: "myregistry.com:/namespace/project/app:latest",
		},
		// DTZ registry specific test cases
		{
			name:     "dtz registry with port and tag",
			input:    "cr.dtz.rocks:3214/image-name:v0.1.2.3",
			expected: "cr.dtz.rocks:3214/image-name:v0.1.2.3",
		},
		{
			name:     "dtz registry with port and digest",
			input:    "cr.dtz.rocks:3214/image-name@sha256:abc1234567890",
			expected: "cr.dtz.rocks:3214/image-name@sha256:abc1234567890",
		},
		{
			name:     "dtz registry with port, tag and digest",
			input:    "cr.dtz.rocks:3214/image-name:v0.1.2.3@sha256:abc1234567890",
			expected: "cr.dtz.rocks:3214/image-name:v0.1.2.3@sha256:abc1234567890",
		},
		{
			name:     "dtz registry with port without tag",
			input:    "cr.dtz.rocks:3214/image-name",
			expected: "cr.dtz.rocks:3214/image-name:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeContainerImage(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeContainerImage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
