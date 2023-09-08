// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure DowntozeroProvider satisfies various provider interfaces.
var _ provider.Provider = &DowntozeroProvider{}

// DowntozeroProvider defines the provider implementation.
type DowntozeroProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DowntozeroProviderModel describes the provider data model.
type DowntozeroProviderModel struct {
	Apikey types.String `tfsdk:"apikey"`
}

func (p *DowntozeroProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "downtozero"
	resp.Version = p.version
}

func (p *DowntozeroProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"apikey": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *DowntozeroProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config DowntozeroProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Apikey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("apikey"),
			"Unknown DownToZero API Apikey",
			"The provider cannot create the DownToZero API client as there is an unknown configuration value for the DownToZero API apikey. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOWNTOZERO_APIKEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apikey := os.Getenv("DOWNTOZERO_APIKEY")
	apiname := "containers"
	apiversion := "2021-02-21"

	if !config.Apikey.IsNull() {
		apikey = config.Apikey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apikey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apikey"),
			"Missing DownToZero API Apikey",
			"The provider cannot create the DownToZero API client as there is a missing or empty value for the DownToZero API apikey. "+
				"Set the apikey value in the configuration or use the DOWNTOZERO_APIKEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Downtozero client using the configuration values
	client, err := NewClient(&apiname, &apiversion, &apikey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Downtozero API Client",
			"An unexpected error occurred when creating the Downtozero API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Downtozero Client Error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DowntozeroProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *DowntozeroProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContainersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DowntozeroProvider{
			version: version,
		}
	}
}
