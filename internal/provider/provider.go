// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
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
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *DowntozeroProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "downtozero"
	resp.Version = p.version
}

func (p *DowntozeroProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
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

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown DownToZero API Endpoint",
			"The provider cannot create the DownToZero API client as there is an unknown configuration value for the DownToZero API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOWNTOZERO_ENDPOINT environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown DownToZero API Username",
			"The provider cannot create the DownToZero API client as there is an unknown configuration value for the DownToZero API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOWNTOZERO_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown DownToZero API Password",
			"The provider cannot create the DownToZero API client as there is an unknown configuration value for the DownToZero API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DOWNTOZERO_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	endpoint := os.Getenv("DOWNTOZERO_ENDPOINT")
	username := os.Getenv("DOWNTOZERO_USERNAME")
	password := os.Getenv("DOWNTOZERO_PASSWORD")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing DownToZero API Endpoint",
			"The provider cannot create the DownToZero API client as there is a missing or empty value for the DownToZero API endpoint. "+
				"Set the endpoint value in the configuration or use the DOWNTOZERO_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing DownToZero API Username",
			"The provider cannot create the DownToZero API client as there is a missing or empty value for the DownToZero API username. "+
				"Set the username value in the configuration or use the DOWNTOZERO_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing DownToZero API Password",
			"The provider cannot create the DownToZero API client as there is a missing or empty value for the DownToZero API password. "+
				"Set the password value in the configuration or use the DOWNTOZERO_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Make the client available during DataSource and Resource
	// type Configure methods.
	client := http.DefaultClient
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
		NewCoffeesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DowntozeroProvider{
			version: version,
		}
	}
}
