package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource = &containersDomainDataSource{}
)

func newContainersDomainDataSource() datasource.DataSource {
	return &containersDomainDataSource{}
}

type containersDomainDataSource struct {
	Name      types.String `tfsdk:"name"`
	ContextId types.String `tfsdk:"context_id"`
	Verified  types.Bool   `tfsdk:"verified"`
	Created   types.String `tfsdk:"created"`
	api_key   string
}

func (d *containersDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers_domain"
}

func (d *containersDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Domain name to fetch. If omitted, the first domain is returned.",
			},
			"context_id": schema.StringAttribute{
				Computed: true,
			},
			"verified": schema.BoolAttribute{
				Computed: true,
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *containersDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *containersDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config containersDomainDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if name provided, fetch specific domain
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/domain/%s", config.Name.ValueString())
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
			return
		}
		request.Header.Set("X-API-KEY", d.api_key)

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domain, got error: %s", err))
			return
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
			return
		}
		defer deferredCloseResponseBody(ctx, response.Body)

		if response.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Domain '%s' not found", config.Name.ValueString()))
			return
		}

		var domain containersDomainResponse
		if err := json.Unmarshal(body, &domain); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
			return
		}

		state := containersDomainDataSource{
			Name:      types.StringValue(domain.Name),
			ContextId: types.StringValue(domain.ContextId),
			Verified:  types.BoolValue(domain.Verified),
			Created:   types.StringValue(domain.Created),
		}
		diags := resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// otherwise, fetch all domains and return the system-generated one if present
	url := "https://containers.dtz.rocks/api/2021-02-21/domain"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list domains, got error: %s", err))
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	var domains []containersDomainResponse
	if err := json.Unmarshal(body, &domains); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	if len(domains) == 0 {
		resp.Diagnostics.AddError("Not Found", "No domains found in this context")
		return
	}

	// Prefer the system-generated domain ending with '.containers.dtz.dev'
	var selected *containersDomainResponse
	for i := range domains {
		if strings.HasSuffix(domains[i].Name, ".containers.dtz.dev") {
			selected = &domains[i]
			break
		}
	}
	if selected == nil {
		selected = &domains[0]
	}
	state := containersDomainDataSource{
		Name:      types.StringValue(selected.Name),
		ContextId: types.StringValue(selected.ContextId),
		Verified:  types.BoolValue(selected.Verified),
		Created:   types.StringValue(selected.Created),
	}
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
