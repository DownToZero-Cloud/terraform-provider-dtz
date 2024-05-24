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
	_ datasource.DataSource = &contextDataSource{}
)

func NewContextDataSource() datasource.DataSource {
	return &contextDataSource{}
}

type contextDataSource struct {
	Id      types.String `tfsdk:"id"`
	Alias   types.String `tfsdk:"alias"`
	api_key string
}

type contextResponse struct {
	ContextId string `json:"contextId"`
	Alias     string `json:"alias"`
	Created   string `json:"created"`
}

func (d *contextDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context"
}

func (d *contextDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:  true,
				Required:  false,
				Sensitive: false,
			},
			"alias": schema.StringAttribute{
				Computed:  true,
				Required:  false,
				Sensitive: false,
			},
		},
	}
}

func (d *contextDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *contextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state contextDataSource
	tflog.Info(ctx, "query API")

	request, err := http.NewRequest(http.MethodGet, "https://dtz.rocks/api/2021-12-09/context", nil)
	if err != nil {
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	tflog.Info(ctx, fmt.Sprintf("status: %d, body: %s", response.StatusCode, string(body[:])))

	var resp_type contextResponse
	err = json.Unmarshal(body, &resp_type)
	if err != nil {
		return
	}

	state.Alias = types.StringValue(resp_type.Alias)
	state.Id = types.StringValue(resp_type.ContextId)

	// set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
