package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource = &containerRegistryDataSource{}
)

func newContainerRegistryDataSource() datasource.DataSource {
	return &containerRegistryDataSource{}
}

type containerRegistryDataSource struct {
	Url        types.String `tfsdk:"url"`
	ImageCount types.Int64  `tfsdk:"image_count"`
	api_key    string
}

type containerRegistryResponse struct {
	Url        string `json:"serverUrl"`
	ImageCount int64  `json:"imageCount"`
}

func (d *containerRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (d *containerRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the container registry server",
			},
			"image_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of images in the container registry",
			},
		},
	}
}

func (d *containerRegistryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *containerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state containerRegistryDataSource
	tflog.Info(ctx, "query container registry API")

	request, err := http.NewRequest(http.MethodGet, "https://cr.dtz.rocks/api/2023-12-28/stats", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read container registry stats, got error: %s", err))
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	tflog.Info(ctx, fmt.Sprintf("status: %d, body: %s", response.StatusCode, string(body)))

	var resp_type containerRegistryResponse
	err = json.Unmarshal(body, &resp_type)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	state.Url = types.StringValue(resp_type.Url)
	state.ImageCount = types.Int64Value(resp_type.ImageCount)

	// set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
