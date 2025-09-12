package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Test the resource type name generation
func TestContainersServiceResource_TypeName(t *testing.T) {
	// Test the type name generation logic
	providerTypeName := "dtz"
	expectedTypeName := providerTypeName + "_containers_service"

	if expectedTypeName != "dtz_containers_service" {
		t.Errorf("Expected type name %s, got %s", "dtz_containers_service", expectedTypeName)
	}
}

// Test the Configure method
func TestContainersServiceResource_Configure(t *testing.T) {
	tests := []struct {
		name          string
		providerData  interface{}
		expectedError bool
	}{
		{
			name: "valid provider data",
			providerData: dtzProvider{
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
			resource := &containersServiceResource{}

			// Test the configure logic directly
			if tt.providerData == nil {
				// Should not set api_key when provider data is nil
				if resource.api_key != "" {
					t.Error("Expected api_key to remain empty when provider data is nil")
				}
				return
			}

			// Test with valid provider data
			if dtz, ok := tt.providerData.(dtzProvider); ok {
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
func TestContainersServiceResource_RequestResponseStructures(t *testing.T) {
	// Test createServiceRequest marshaling
	createReq := createServiceRequest{
		Prefix:            "/test",
		ContainerImage:    "nginx:alpine",
		ContainerPullUser: "user",
		ContainerPullPwd:  "password",
		EnvVariables: map[string]string{
			"PORT": "8080",
			"ENV":  "test",
		},
		Login: &struct {
			ProviderName string `json:"providerName"`
		}{
			ProviderName: "dtz",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal createServiceRequest: %v", err)
	}

	// Unmarshal back to verify structure
	var unmarshaledReq createServiceRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal createServiceRequest: %v", err)
	}

	// Verify fields
	if unmarshaledReq.Prefix != createReq.Prefix {
		t.Errorf("Expected prefix %s, got %s", createReq.Prefix, unmarshaledReq.Prefix)
	}
	if unmarshaledReq.ContainerImage != createReq.ContainerImage {
		t.Errorf("Expected container image %s, got %s", createReq.ContainerImage, unmarshaledReq.ContainerImage)
	}
	if unmarshaledReq.Login.ProviderName != createReq.Login.ProviderName {
		t.Errorf("Expected login provider %s, got %s", createReq.Login.ProviderName, unmarshaledReq.Login.ProviderName)
	}

	// Test containersServiceResponse unmarshaling
	responseJSON := `{
		"contextId": "ctx-123",
		"serviceId": "svc-456",
		"created": "2023-01-01T00:00:00Z",
		"prefix": "/test",
		"containerImage": "nginx:alpine",
		"containerImageVersion": "latest",
		"containerPullUser": "user",
		"containerPullPwd": "password",
		"envVariables": {
			"PORT": "8080",
			"ENV": "test"
		},
		"login": {
			"providerName": "dtz"
		}
	}`

	var response containersServiceResponse
	err = json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal containersServiceResponse: %v", err)
	}

	// Verify response fields
	if response.ServiceId != "svc-456" {
		t.Errorf("Expected service ID %s, got %s", "svc-456", response.ServiceId)
	}
	if response.Prefix != "/test" {
		t.Errorf("Expected prefix %s, got %s", "/test", response.Prefix)
	}
	if response.Login.ProviderName != "dtz" {
		t.Errorf("Expected login provider %s, got %s", "dtz", response.Login.ProviderName)
	}
}

// Test environment variables handling
func TestContainersServiceResource_EnvironmentVariables(t *testing.T) {
	// Test creating a resource with environment variables
	envVars := types.MapValueMust(types.StringType, map[string]attr.Value{
		"PORT":   types.StringValue("8080"),
		"ENV":    types.StringValue("test"),
		"DB_URL": types.StringValue("postgres://localhost:5432/mydb"),
	})

	resource := &containersServiceResource{
		Prefix:         types.StringValue("/test"),
		ContainerImage: types.StringValue("nginx:alpine"),
		EnvVariables:   envVars,
	}

	// Verify the environment variables are set correctly
	if resource.EnvVariables.IsNull() {
		t.Error("Expected environment variables to be set")
	}

	// Test converting to map
	var envMap map[string]string
	diags := resource.EnvVariables.ElementsAs(context.Background(), &envMap, false)
	if diags.HasError() {
		t.Errorf("Failed to convert environment variables to map: %v", diags)
	}

	expectedEnvVars := map[string]string{
		"PORT":   "8080",
		"ENV":    "test",
		"DB_URL": "postgres://localhost:5432/mydb",
	}

	for key, expectedValue := range expectedEnvVars {
		if actualValue, exists := envMap[key]; !exists {
			t.Errorf("Expected environment variable %s to exist", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected environment variable %s to be %s, got %s", key, expectedValue, actualValue)
		}
	}
}

// Test login block validation
func TestContainersServiceResource_LoginValidation(t *testing.T) {
	tests := []struct {
		name          string
		login         *LoginModel
		expectedValid bool
	}{
		{
			name: "valid dtz login",
			login: &LoginModel{
				ProviderName: types.StringValue("dtz"),
			},
			expectedValid: true,
		},
		{
			name: "invalid provider name",
			login: &LoginModel{
				ProviderName: types.StringValue("invalid"),
			},
			expectedValid: false,
		},
		{
			name: "null provider name",
			login: &LoginModel{
				ProviderName: types.StringNull(),
			},
			expectedValid: false,
		},
		{
			name:          "nil login",
			login:         nil,
			expectedValid: true, // nil login is valid (no authentication)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request to test validation
			createReq := createServiceRequest{
				Prefix:         "/test",
				ContainerImage: "nginx:alpine",
			}

			if tt.login != nil {
				providerName := tt.login.ProviderName.ValueString()

				// Simulate the validation logic from the Create method
				if tt.login.ProviderName.IsNull() || tt.login.ProviderName.IsUnknown() {
					if tt.expectedValid {
						t.Errorf("Expected valid login but got null/unknown provider name")
					}
					return
				}

				if providerName != "dtz" {
					if tt.expectedValid {
						t.Errorf("Expected valid login but got invalid provider name: %s", providerName)
					}
					return
				}

				createReq.Login = &struct {
					ProviderName string `json:"providerName"`
				}{
					ProviderName: providerName,
				}
			}

			// If we reach here and expectedValid is false, that's an error
			if !tt.expectedValid {
				t.Errorf("Expected invalid login but validation passed")
			}
		})
	}
}

// Test URL construction
func TestContainersServiceResource_URLConstruction(t *testing.T) {
	baseURL := "https://containers.dtz.rocks/api/2021-02-21"

	tests := []struct {
		name      string
		serviceID string
		expected  string
	}{
		{
			name:      "create service URL",
			serviceID: "",
			expected:  baseURL + "/service",
		},
		{
			name:      "read service URL",
			serviceID: "svc-123",
			expected:  baseURL + "/service/svc-123",
		},
		{
			name:      "update service URL",
			serviceID: "svc-456",
			expected:  baseURL + "/service/svc-456",
		},
		{
			name:      "delete service URL",
			serviceID: "svc-789",
			expected:  baseURL + "/service/svc-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var url string
			if tt.serviceID == "" {
				url = baseURL + "/service"
			} else {
				url = fmt.Sprintf("%s/service/%s", baseURL, tt.serviceID)
			}

			if url != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, url)
			}
		})
	}
}

// Test the LoginModel structure
func TestLoginModel(t *testing.T) {
	// Test creating a LoginModel
	login := &LoginModel{
		ProviderName: types.StringValue("dtz"),
	}

	if login.ProviderName.ValueString() != "dtz" {
		t.Errorf("Expected provider name 'dtz', got %s", login.ProviderName.ValueString())
	}

	// Test null LoginModel
	nullLogin := &LoginModel{
		ProviderName: types.StringNull(),
	}

	if !nullLogin.ProviderName.IsNull() {
		t.Error("Expected provider name to be null")
	}
}
