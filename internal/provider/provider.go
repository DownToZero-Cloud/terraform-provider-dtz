package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.Provider = &dtzProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dtzProvider{
			version: version,
		}
	}
}

type dtzProvider struct {
	version string
	ApiKey  string `tfsdk:"api_key"`
}

func (p *dtzProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dtz"
	resp.Version = p.version
}

func (p *dtzProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The API key for authentication",
			},
		},
	}
}

func (p *dtzProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config dtzProvider
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.DataSourceData = config
	resp.ResourceData = config
}

func (p *dtzProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newContextDataSource,
		newRss2emailFeedDataSource,
		newRss2emailProfileDataSource,
	}
}

func (p *dtzProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
