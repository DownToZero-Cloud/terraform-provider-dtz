package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	version                        string
	ApiKey                         string     `tfsdk:"api_key"`
	EnableServiceContainers        types.Bool `tfsdk:"enable_service_containers"`
	EnableServiceObjectstore       types.Bool `tfsdk:"enable_service_objectstore"`
	EnableServiceContainerregistry types.Bool `tfsdk:"enable_service_containerregistry"`
	EnableServiceRss2email         types.Bool `tfsdk:"enable_service_rss2email"`
	EnableServiceObservability     types.Bool `tfsdk:"enable_service_observability"`
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
				Validators: []validator.String{
					stringvalidator.LengthBetween(30, 43),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^apikey-`),
						"must start with 'apikey-'",
					),
				},
			},
			"enable_service_containers": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable the containers service",
			},
			"enable_service_objectstore": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable the object store service",
			},
			"enable_service_containerregistry": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable the container registry service",
			},
			"enable_service_rss2email": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable the RSS2Email service",
			},
			"enable_service_observability": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable the observability service",
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

	if config.EnableServiceRss2email.ValueBool() {
		err := enableRss2EmailService(ctx, config.ApiKey)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to enable RSS2Email service",
				fmt.Sprintf("An error occurred when calling the RSS2Email service endpoint: %s", err),
			)
			return
		}
	}

	if config.EnableServiceContainers.ValueBool() {
		err := enableContainersService(ctx, config.ApiKey)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to enable Containers service",
				fmt.Sprintf("An error occurred when calling the Containers service endpoint: %s", err),
			)
			return
		}
	}

	if config.EnableServiceObjectstore.ValueBool() {
		err := enableObjectstoreService(ctx, config.ApiKey)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to enable Objectstore service",
				fmt.Sprintf("An error occurred when calling the Objectstore service endpoint: %s", err),
			)
			return
		}
	}

	if config.EnableServiceContainerregistry.ValueBool() {
		err := enableContainerregistryService(ctx, config.ApiKey)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to enable Container Registry service",
				fmt.Sprintf("An error occurred when calling the Container Registry service endpoint: %s", err),
			)
			return
		}
	}

	if config.EnableServiceObservability.ValueBool() {
		err := enableObservabilityService(ctx, config.ApiKey)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to enable Observability service",
				fmt.Sprintf("An error occurred when calling the Observability service endpoint: %s", err),
			)
			return
		}
	}

	resp.DataSourceData = config
	resp.ResourceData = config
}

func (p *dtzProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newContainerRegistryDataSource,
		newContextDataSource,
		newContainersDomainDataSource,
		newRss2emailFeedDataSource,
		newRss2emailProfileDataSource,
	}
}

func (p *dtzProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newIdentityApikeyResource,
		newRss2emailFeedResource,
		newRss2emailProfileResource,
		newContainersJobResource,
		newContainersDomainResource,
		newContainersServiceResource,
	}
}

func enableRss2EmailService(ctx context.Context, apiKey string) error {
	url := "https://rss2email.dtz.rocks/api/2021-02-01/enable"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	defer deferredCloseResponseBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func enableContainersService(ctx context.Context, apiKey string) error {
	url := "https://containers.dtz.rocks/api/2021-02-21/enable"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer deferredCloseResponseBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func enableObjectstoreService(ctx context.Context, apiKey string) error {
	url := "https://objectstore.dtz.rocks/api/2022-11-28/enable"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer deferredCloseResponseBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func enableContainerregistryService(ctx context.Context, apiKey string) error {
	url := "https://cr.dtz.rocks/api/2023-12-28/enable"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer deferredCloseResponseBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func enableObservabilityService(ctx context.Context, apiKey string) error {
	url := "https://observability.dtz.rocks/api/2021-02-01/enable"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer deferredCloseResponseBody(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
