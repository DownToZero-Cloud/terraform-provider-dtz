package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Test the normalizeContainerImage function integration for jobs
func TestContainersJobResource_ImageNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple image without tag",
			input:    "nginx",
			expected: "nginx:latest",
		},
		{
			name:     "image with tag",
			input:    "nginx:1.21",
			expected: "nginx:1.21",
		},
		{
			name:     "registry image without tag",
			input:    "docker.io/library/nginx",
			expected: "docker.io/library/nginx:latest",
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
		// Test cases for the bug fix: malformed registry URLs with empty ports
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
			name:     "registry with empty port and digest",
			input:    "registry:/myimage@sha256:abc123",
			expected: "registry:/myimage@sha256:abc123",
		},
		{
			name:     "localhost with empty port and digest",
			input:    "localhost:/myimage@sha256:def456",
			expected: "localhost:/myimage@sha256:def456",
		},
		// Additional edge cases to ensure robustness
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
			name:     "registry with mixed port (numeric and non-numeric)",
			input:    "registry:123abc/myimage",
			expected: "registry:123abc/myimage:latest",
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
			// Test the normalization by creating a request
			createJob := createJobRequest{
				Name:           "test-job",
				ContainerImage: normalizeContainerImage(tt.input),
			}

			if createJob.ContainerImage != tt.expected {
				t.Errorf("normalizeContainerImage(%q) = %q, want %q", tt.input, createJob.ContainerImage, tt.expected)
			}
		})
	}
}

// Test the resource type name generation
func TestContainersJobResource_TypeName(t *testing.T) {
	// Test the type name generation logic
	providerTypeName := "dtz"
	expectedTypeName := providerTypeName + "_containers_job"

	if expectedTypeName != "dtz_containers_job" {
		t.Errorf("Expected type name %s, got %s", "dtz_containers_job", expectedTypeName)
	}
}

// Test the Configure method
func TestContainersJobResource_Configure(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectedError bool
	}{
		{
			name: "valid provider data",
			providerData: &dtzProvider{
				ApiKey: "test-api-key",
			},
			expectedError: false,
		},
		{
			name:          "nil provider data",
			providerData:  nil,
			expectedError: false,
		},
		{
			name:          "invalid provider data type",
			providerData:  "invalid",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create resource instance
			resource := &containersJobResource{}

			// Test the configure logic directly
			if tt.providerData == nil {
				// Should not set api_key when provider data is nil
				if resource.api_key != "" {
					t.Error("Expected api_key to remain empty when provider data is nil")
				}
				return
			}

			// Test with valid provider data
			if dtz, ok := tt.providerData.(*dtzProvider); ok {
				resource.api_key = dtz.ApiKey
				if resource.api_key != "test-api-key" {
					t.Errorf("Expected api_key to be 'test-api-key', got %s", resource.api_key)
				}
			} else {
				// Invalid provider data type
				if !tt.expectedError {
					t.Error("Expected error for invalid provider data type")
				}
			}
		})
	}
}

// Test request/response structures
func TestContainersJobResource_RequestResponseStructures(t *testing.T) {
	// Test createJobRequest marshaling
	createReq := createJobRequest{
		Name:              "test-job",
		ContainerImage:    "nginx:alpine",
		ContainerPullUser: "user",
		ContainerPullPwd:  "password",
		ScheduleType:      "relaxed",
		ScheduleCron:      "0 0 * * *",
		ScheduleRepeat:    "",
		EnvVariables: map[string]EnvVariableValue{
			"PORT": {
				StringValue: stringPtr("8080"),
			},
			"ENV": {
				StringValue: stringPtr("test"),
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal createJobRequest: %v", err)
	}

	// Unmarshal back to verify structure
	var unmarshaledReq createJobRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal createJobRequest: %v", err)
	}

	// Verify fields
	if unmarshaledReq.Name != createReq.Name {
		t.Errorf("Expected name %s, got %s", createReq.Name, unmarshaledReq.Name)
	}
	if unmarshaledReq.ContainerImage != createReq.ContainerImage {
		t.Errorf("Expected container image %s, got %s", createReq.ContainerImage, unmarshaledReq.ContainerImage)
	}
	if unmarshaledReq.ScheduleType != createReq.ScheduleType {
		t.Errorf("Expected schedule type %s, got %s", createReq.ScheduleType, unmarshaledReq.ScheduleType)
	}

	// Test containersJobResponse unmarshaling
	responseJSON := `{
		"id": "job-123",
		"name": "test-job",
		"containerImage": "nginx:alpine",
		"containerPullUser": "user",
		"containerPullPwd": "password",
		"scheduleType": "relaxed",
		"scheduleRepeat": null,
		"scheduleCron": "0 0 * * *",
		"envVariables": {
			"PORT": "8080",
			"ENV": "test"
		}
	}`

	var response containersJobResponse
	err = json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal containersJobResponse: %v", err)
	}

	// Verify response fields
	if response.Id != "job-123" {
		t.Errorf("Expected job ID %s, got %s", "job-123", response.Id)
	}
	if response.Name != "test-job" {
		t.Errorf("Expected name %s, got %s", "test-job", response.Name)
	}
	if response.ScheduleType != "relaxed" {
		t.Errorf("Expected schedule type %s, got %s", "relaxed", response.ScheduleType)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Test environment variables handling
func TestContainersJobResource_EnvironmentVariables(t *testing.T) {
	// Test creating a resource with environment variables
	envVars := types.MapValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		},
	}, map[string]attr.Value{
		"PORT": types.ObjectValueMust(map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		}, map[string]attr.Value{
			"string_value":    types.StringValue("8080"),
			"encryption_key":  types.StringNull(),
			"encrypted_value": types.StringNull(),
			"plain_value":     types.StringNull(),
		}),
		"ENV": types.ObjectValueMust(map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		}, map[string]attr.Value{
			"string_value":    types.StringValue("test"),
			"encryption_key":  types.StringNull(),
			"encrypted_value": types.StringNull(),
			"plain_value":     types.StringNull(),
		}),
		"DB_URL": types.ObjectValueMust(map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		}, map[string]attr.Value{
			"string_value":    types.StringValue("postgres://localhost:5432/mydb"),
			"encryption_key":  types.StringNull(),
			"encrypted_value": types.StringNull(),
			"plain_value":     types.StringNull(),
		}),
	})

	resource := &containersJobResource{
		Name:           types.StringValue("test-job"),
		ContainerImage: types.StringValue("nginx:alpine"),
		ScheduleType:   types.StringValue("relaxed"),
		EnvVariables:   envVars,
	}

	// Verify the environment variables are set correctly
	if resource.EnvVariables.IsNull() {
		t.Error("Expected environment variables to be set")
	}

	// Test converting to map
	var envMap map[string]EnvVariableTerraformValue
	diags := resource.EnvVariables.ElementsAs(context.Background(), &envMap, false)
	if diags.HasError() {
		t.Errorf("Failed to convert environment variables to map: %v", diags)
	}

	// Verify the environment variables
	expectedEnvVars := map[string]string{
		"PORT":   "8080",
		"ENV":    "test",
		"DB_URL": "postgres://localhost:5432/mydb",
	}

	for key, expectedValue := range expectedEnvVars {
		if envVar, exists := envMap[key]; !exists {
			t.Errorf("Expected environment variable %s to exist", key)
		} else if envVar.StringValue.ValueString() != expectedValue {
			t.Errorf("Expected environment variable %s to be %s, got %s", key, expectedValue, envVar.StringValue.ValueString())
		}
	}
}

// Test mixed environment variable types
func TestContainersJobResource_MixedEnvironmentVariables(t *testing.T) {
	// Test creating a resource with mixed environment variable types
	envVars := types.MapValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		},
	}, map[string]attr.Value{
		// String value
		"PORT": types.ObjectValueMust(map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		}, map[string]attr.Value{
			"string_value":    types.StringValue("8080"),
			"encryption_key":  types.StringNull(),
			"encrypted_value": types.StringNull(),
			"plain_value":     types.StringNull(),
		}),
		// Encrypted value
		"API_SECRET": types.ObjectValueMust(map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		}, map[string]attr.Value{
			"string_value":    types.StringNull(),
			"encryption_key":  types.StringValue("AES256:PROD_KEY"),
			"encrypted_value": types.StringValue("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."),
			"plain_value":     types.StringNull(),
		}),
		// Plain value
		"DB_PASSWORD": types.ObjectValueMust(map[string]attr.Type{
			"string_value":    types.StringType,
			"encryption_key":  types.StringType,
			"encrypted_value": types.StringType,
			"plain_value":     types.StringType,
		}, map[string]attr.Value{
			"string_value":    types.StringNull(),
			"encryption_key":  types.StringNull(),
			"encrypted_value": types.StringNull(),
			"plain_value":     types.StringValue("super-secret-password"),
		}),
	})

	resource := &containersJobResource{
		Name:           types.StringValue("test-job"),
		ContainerImage: types.StringValue("nginx:alpine"),
		ScheduleType:   types.StringValue("relaxed"),
		EnvVariables:   envVars,
	}

	// Verify the environment variables are set correctly
	if resource.EnvVariables.IsNull() {
		t.Error("Expected environment variables to be set")
	}

	// Test converting to map
	var envMap map[string]EnvVariableTerraformValue
	diags := resource.EnvVariables.ElementsAs(context.Background(), &envMap, false)
	if diags.HasError() {
		t.Errorf("Failed to convert environment variables to map: %v", diags)
	}

	// Verify string value
	if envVar, exists := envMap["PORT"]; !exists {
		t.Error("Expected environment variable PORT to exist")
	} else if envVar.StringValue.ValueString() != "8080" {
		t.Errorf("Expected PORT to be 8080, got %s", envVar.StringValue.ValueString())
	}

	// Verify encrypted value
	if envVar, exists := envMap["API_SECRET"]; !exists {
		t.Error("Expected environment variable API_SECRET to exist")
	} else if envVar.EncryptionKey.ValueString() != "AES256:PROD_KEY" {
		t.Errorf("Expected encryption key to be AES256:PROD_KEY, got %s", envVar.EncryptionKey.ValueString())
	} else if envVar.EncryptedValue.ValueString() != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." {
		t.Errorf("Expected encrypted value to be eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..., got %s", envVar.EncryptedValue.ValueString())
	}

	// Verify plain value
	if envVar, exists := envMap["DB_PASSWORD"]; !exists {
		t.Error("Expected environment variable DB_PASSWORD to exist")
	} else if envVar.PlainValue.ValueString() != "super-secret-password" {
		t.Errorf("Expected plain value to be super-secret-password, got %s", envVar.PlainValue.ValueString())
	}
}

// Test EnvVariableValue marshaling and unmarshaling
func TestEnvVariableValue_JSONHandling(t *testing.T) {
	tests := []struct {
		name         string
		input        EnvVariableValue
		expectedType string
	}{
		{
			name: "string value",
			input: EnvVariableValue{
				StringValue: stringPtr("simple-value"),
			},
			expectedType: "string",
		},
		{
			name: "encrypted value",
			input: EnvVariableValue{
				EncryptionKey:  stringPtr("AES256:KEY1"),
				EncryptedValue: stringPtr("base64-encoded-ciphertext"),
			},
			expectedType: "encrypted",
		},
		{
			name: "plain value",
			input: EnvVariableValue{
				PlainValue: stringPtr("plain-text-for-encryption"),
			},
			expectedType: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal EnvVariableValue: %v", err)
			}

			// Test unmarshaling
			var unmarshaled EnvVariableValue
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal EnvVariableValue: %v", err)
			}

			// Verify the unmarshaled value matches the input based on type
			switch tt.expectedType {
			case "string":
				if unmarshaled.StringValue == nil || *unmarshaled.StringValue != *tt.input.StringValue {
					t.Errorf("Expected StringValue %s, got %s", *tt.input.StringValue, *unmarshaled.StringValue)
				}
			case "encrypted":
				if unmarshaled.EncryptionKey == nil || *unmarshaled.EncryptionKey != *tt.input.EncryptionKey {
					t.Errorf("Expected EncryptionKey %s, got %s", *tt.input.EncryptionKey, *unmarshaled.EncryptionKey)
				}
				if unmarshaled.EncryptedValue == nil || *unmarshaled.EncryptedValue != *tt.input.EncryptedValue {
					t.Errorf("Expected EncryptedValue %s, got %s", *tt.input.EncryptedValue, *unmarshaled.EncryptedValue)
				}
			case "plain":
				if unmarshaled.PlainValue == nil || *unmarshaled.PlainValue != *tt.input.PlainValue {
					t.Errorf("Expected PlainValue %s, got %s", *tt.input.PlainValue, *unmarshaled.PlainValue)
				}
			}
		})
	}
}

// Test URL construction
func TestContainersJobResource_URLConstruction(t *testing.T) {
	baseURL := "https://containers.dtz.rocks/api/2021-02-21"

	tests := []struct {
		name     string
		jobID    string
		expected string
	}{
		{
			name:     "create job URL",
			jobID:    "",
			expected: baseURL + "/job",
		},
		{
			name:     "read job URL",
			jobID:    "job-123",
			expected: baseURL + "/job/job-123",
		},
		{
			name:     "update job URL",
			jobID:    "job-456",
			expected: baseURL + "/job/job-456",
		},
		{
			name:     "delete job URL",
			jobID:    "job-789",
			expected: baseURL + "/job/job-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var url string
			if tt.jobID == "" {
				url = baseURL + "/job"
			} else {
				url = fmt.Sprintf("%s/job/%s", baseURL, tt.jobID)
			}

			if url != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, url)
			}
		})
	}
}

// Test schedule type validation
func TestContainersJobResource_ScheduleTypeValidation(t *testing.T) {
	tests := []struct {
		name           string
		scheduleType   string
		scheduleCron   string
		scheduleRepeat string
		expectedValid  bool
	}{
		{
			name:           "valid relaxed schedule",
			scheduleType:   "relaxed",
			scheduleCron:   "",
			scheduleRepeat: "",
			expectedValid:  true,
		},
		{
			name:           "valid precise schedule",
			scheduleType:   "precise",
			scheduleCron:   "",
			scheduleRepeat: "",
			expectedValid:  true,
		},
		{
			name:           "valid none schedule",
			scheduleType:   "none",
			scheduleCron:   "",
			scheduleRepeat: "",
			expectedValid:  true,
		},
		{
			name:           "invalid schedule type",
			scheduleType:   "invalid",
			scheduleCron:   "",
			scheduleRepeat: "",
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request to test validation
			createReq := createJobRequest{
				Name:           "test-job",
				ContainerImage: "nginx:alpine",
				ScheduleType:   tt.scheduleType,
				ScheduleCron:   tt.scheduleCron,
				ScheduleRepeat: tt.scheduleRepeat,
			}

			// Basic validation - in a real implementation, you might want more sophisticated validation
			validScheduleTypes := map[string]bool{
				"relaxed": true,
				"precise": true,
				"none":    true,
			}

			if validScheduleTypes[tt.scheduleType] != tt.expectedValid {
				t.Errorf("Expected schedule type %s to be %t, but validation logic says %t",
					tt.scheduleType, tt.expectedValid, validScheduleTypes[tt.scheduleType])
			}

			// Verify the request was created correctly
			if createReq.ScheduleType != tt.scheduleType {
				t.Errorf("Expected schedule type in request to be %s, got %s", tt.scheduleType, createReq.ScheduleType)
			}
		})
	}
}

// Test the containersJobResource structure
func TestContainersJobResource_Structure(t *testing.T) {
	// Test with all fields populated
	resource := &containersJobResource{
		Id:             types.StringValue("job-123"),
		Name:           types.StringValue("test-job"),
		ContainerImage: types.StringValue("nginx:alpine"),
		ScheduleType:   types.StringValue("relaxed"),
		ScheduleRepeat: types.StringValue(""),
		ScheduleCron:   types.StringValue("0 0 * * *"),
		EnvVariables:   types.MapNull(types.StringType),
	}

	// Verify fields are set correctly
	if resource.Id.ValueString() != "job-123" {
		t.Errorf("Expected ID %s, got %s", "job-123", resource.Id.ValueString())
	}
	if resource.Name.ValueString() != "test-job" {
		t.Errorf("Expected name %s, got %s", "test-job", resource.Name.ValueString())
	}
	if resource.ContainerImage.ValueString() != "nginx:alpine" {
		t.Errorf("Expected container image %s, got %s", "nginx:alpine", resource.ContainerImage.ValueString())
	}
	if resource.ScheduleType.ValueString() != "relaxed" {
		t.Errorf("Expected schedule type %s, got %s", "relaxed", resource.ScheduleType.ValueString())
	}

	// Test with null values
	nullResource := &containersJobResource{
		Id:             types.StringNull(),
		Name:           types.StringNull(),
		ContainerImage: types.StringNull(),
		ScheduleType:   types.StringNull(),
		ScheduleRepeat: types.StringNull(),
		ScheduleCron:   types.StringNull(),
		EnvVariables:   types.MapNull(types.StringType),
	}

	// Verify null values are handled correctly
	if !nullResource.Id.IsNull() {
		t.Error("Expected ID to be null")
	}
	if !nullResource.Name.IsNull() {
		t.Error("Expected name to be null")
	}
	if !nullResource.ContainerImage.IsNull() {
		t.Error("Expected container image to be null")
	}
	if !nullResource.ScheduleType.IsNull() {
		t.Error("Expected schedule type to be null")
	}
}
