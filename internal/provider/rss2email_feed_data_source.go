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
	_ datasource.DataSource = &rss2emailFeedDataSource{}
)

func newRss2emailFeedDataSource() datasource.DataSource {
	return &rss2emailFeedDataSource{}
}

type rss2emailFeedDataSource struct {
	Id            types.String `tfsdk:"id"`
	Url           types.String `tfsdk:"url"`
	Name          types.String `tfsdk:"name"`
	LastCheck     types.String `tfsdk:"last_check"`
	LastDataFound types.String `tfsdk:"last_data_found"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	api_key       string
}

type rss2emailFeedResponse struct {
	Id            string `json:"id"`
	Url           string `json:"url"`
	LastCheck     string `json:"lastCheck"`
	LastDataFound string `json:"lastDataFound"`
	Enabled       bool   `json:"enabled"`
	Name          string `json:"name"`
}

func (d *rss2emailFeedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rss2email_feed"
}

func (d *rss2emailFeedDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"last_check": schema.StringAttribute{
				Computed: true,
			},
			"last_data_found": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *rss2emailFeedDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *rss2emailFeedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rss2emailFeedDataSource
	var config_data rss2emailFeedDataSource
	// tflog.Info(ctx, fmt.Sprintf("read config %+v", req))
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

	var resp_type rss2emailFeedResponse
	err = json.Unmarshal(body, &resp_type)
	if err != nil {
		tflog.Error(ctx, "error unmarshalling")
		return
	}
	tflog.Info(ctx, fmt.Sprintf("rssFeedDataSource Read response: %+v", resp_type))

	state.Id = types.StringValue(resp_type.Id)
	state.Url = types.StringValue(resp_type.Url)
	state.Name = types.StringValue(resp_type.Name)
	state.Enabled = types.BoolValue(resp_type.Enabled)
	state.LastCheck = types.StringValue(resp_type.LastDataFound)
	state.LastDataFound = types.StringValue(resp_type.LastDataFound)
	// set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
