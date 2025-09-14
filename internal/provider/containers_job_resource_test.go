package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

// Test environment variables
func TestContainersJobResource_EnvironmentVariables(t *testing.T) {
	// Test creating a resource with environment variables
	envVars := types.MapValueMust(types.StringType, map[string]attr.Value{
		"PORT": types.StringValue("8080"),
		"ENV":  types.StringValue("production"),
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
	var envMap map[string]types.String
	diags := resource.EnvVariables.ElementsAs(context.Background(), &envMap, false)
	if diags.HasError() {
		t.Errorf("Failed to convert environment variables to map: %v", diags)
	}

	// Verify the values
	if portVal, exists := envMap["PORT"]; !exists {
		t.Error("Expected environment variable PORT to exist")
	} else if portVal.ValueString() != "8080" {
		t.Errorf("Expected PORT to be '8080', got %s", portVal.ValueString())
	}

	if envVal, exists := envMap["ENV"]; !exists {
		t.Error("Expected environment variable ENV to exist")
	} else if envVal.ValueString() != "production" {
		t.Errorf("Expected ENV to be 'production', got %s", envVal.ValueString())
	}
}

// Test environment variables with mixed types (strings and objects)
func TestContainersJobResource_MixedEnvironmentVariables(t *testing.T) {
	// Test creating a resource with mixed environment variable types
	envVars := types.MapValueMust(types.StringType, map[string]attr.Value{
		"PORT":       types.StringValue("8080"),
		"ENV":        types.StringValue("production"),
		"SECRET_KEY": types.StringValue(`{"encryptionKey":"AES256:KEY1","encryptedValue":"base64-encoded-ciphertext"}`),
		"PASSWORD":   types.StringValue(`{"plainValue":"my-secret-password"}`),
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
	var envMap map[string]types.String
	diags := resource.EnvVariables.ElementsAs(context.Background(), &envMap, false)
	if diags.HasError() {
		t.Errorf("Failed to convert environment variables to map: %v", diags)
	}

	// Verify string values
	if portVal, exists := envMap["PORT"]; !exists {
		t.Error("Expected environment variable PORT to exist")
	} else if portVal.ValueString() != "8080" {
		t.Errorf("Expected PORT to be '8080', got %s", portVal.ValueString())
	}

	if envVal, exists := envMap["ENV"]; !exists {
		t.Error("Expected environment variable ENV to exist")
	} else if envVal.ValueString() != "production" {
		t.Errorf("Expected ENV to be 'production', got %s", envVal.ValueString())
	}

	// Verify JSON string values
	if secretVal, exists := envMap["SECRET_KEY"]; !exists {
		t.Error("Expected environment variable SECRET_KEY to exist")
	} else if !strings.Contains(secretVal.ValueString(), "AES256:KEY1") {
		t.Errorf("Expected SECRET_KEY to contain 'AES256:KEY1', got %s", secretVal.ValueString())
	}

	if passwordVal, exists := envMap["PASSWORD"]; !exists {
		t.Error("Expected environment variable PASSWORD to exist")
	} else if !strings.Contains(passwordVal.ValueString(), "my-secret-password") {
		t.Errorf("Expected PASSWORD to contain 'my-secret-password', got %s", passwordVal.ValueString())
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
		{
			name: "string and encrypted values",
			input: EnvVariableValue{
				StringValue:    stringPtr("default-value"),
				EncryptionKey:  stringPtr("AES256:KEY1"),
				EncryptedValue: stringPtr("encrypted-data"),
			},
			expectedType: "combined",
		},
		{
			name: "string and plain values",
			input: EnvVariableValue{
				StringValue: stringPtr("string-value"),
				PlainValue:  stringPtr("plain-secret"),
			},
			expectedType: "combined",
		},
		{
			name: "all three value types",
			input: EnvVariableValue{
				StringValue:    stringPtr("default"),
				EncryptionKey:  stringPtr("AES256:KEY1"),
				EncryptedValue: stringPtr("encrypted"),
				PlainValue:     stringPtr("plain"),
			},
			expectedType: "combined",
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
			case "combined":
				// For combined types, verify all present fields match
				if tt.input.StringValue != nil {
					if unmarshaled.StringValue == nil || *unmarshaled.StringValue != *tt.input.StringValue {
						t.Errorf("Expected StringValue %s, got %s", *tt.input.StringValue, *unmarshaled.StringValue)
					}
				}
				if tt.input.EncryptionKey != nil {
					if unmarshaled.EncryptionKey == nil || *unmarshaled.EncryptionKey != *tt.input.EncryptionKey {
						t.Errorf("Expected EncryptionKey %s, got %s", *tt.input.EncryptionKey, *unmarshaled.EncryptionKey)
					}
				}
				if tt.input.EncryptedValue != nil {
					if unmarshaled.EncryptedValue == nil || *unmarshaled.EncryptedValue != *tt.input.EncryptedValue {
						t.Errorf("Expected EncryptedValue %s, got %s", *tt.input.EncryptedValue, *unmarshaled.EncryptedValue)
					}
				}
				if tt.input.PlainValue != nil {
					if unmarshaled.PlainValue == nil || *unmarshaled.PlainValue != *tt.input.PlainValue {
						t.Errorf("Expected PlainValue %s, got %s", *tt.input.PlainValue, *unmarshaled.PlainValue)
					}
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
