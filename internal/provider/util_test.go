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
