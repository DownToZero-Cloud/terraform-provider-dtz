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
	_ datasource.DataSource = &rssFeedDataSource{}
)

func NewRssFeedDataSource() datasource.DataSource {
	return &rssFeedDataSource{}
}

type rssFeedDataSource struct {
	Id            types.String `tfsdk:"id"`
	Url           types.String `tfsdk:"url"`
	lastCheck     types.String `tfsdk:"lastCheck"`
	lastDataFound types.String `tfsdk:"lastDataFound"`
	enabled       types.Bool   `tfsdk:"enabled"`
	Name          types.String `tfsdk:"name"`
	api_key       string
}

type rssFeedConfig struct {
	Id   types.String `tfsdk:"id"`
	Url  types.String `tfsdk:"url"`
	Name types.String `tfsdk:"name"`
}

type rssFeedResponse struct {
	Id            string `json:"id"`
	Url           string `json:"url"`
	LastCheck     string `json:"lastCheck"`
	LastDataFound string `json:"lastDataFound"`
	Enabled       bool   `json:"enabled"`
	Name          string `json:"name"`
}

func (d *rssFeedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rss_feed"
}

func (d *rssFeedDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:  false,
				Required:  true,
				Sensitive: false,
			},
			"url": schema.StringAttribute{
				Computed:  false,
				Optional:  true,
				Sensitive: false,
			},
			"name": schema.StringAttribute{
				Computed:  true,
				Optional:  true,
				Sensitive: false,
			},
		},
	}
}

func (d *rssFeedDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *rssFeedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rssFeedDataSource
	var config_data rssFeedConfig
	tflog.Info(ctx, fmt.Sprintf("read config %+v", req))
	resp.Diagnostics.Append(req.Config.Get(ctx, &config_data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("read data %+v", config_data))
	var feed_id = config_data.Id
	var url = fmt.Sprintf("https://rss2email.dtz.rocks/api/2021-02-01/rss2email/feed/%s", feed_id.ValueString())
	tflog.Info(ctx, "query API "+url)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		tflog.Error(ctx, "error retrieving")
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		tflog.Error(ctx, "error fetching")
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		tflog.Error(ctx, "error reading")
		return
	}
	tflog.Info(ctx, fmt.Sprintf("rssFeedDataSource Read status: %d, body: %s", response.StatusCode, string(body[:])))

	var resp_type rssFeedResponse
	err = json.Unmarshal(body, &resp_type)
	if err != nil {
		tflog.Error(ctx, "error unmarshalling")
		return
	}
	tflog.Info(ctx, fmt.Sprintf("rssFeedDataSource Read response: %+v", resp_type))

	state.Id = types.StringValue(resp_type.Id)
	state.Url = types.StringValue(resp_type.Url)
	state.Name = types.StringValue(resp_type.Name)
	// set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
