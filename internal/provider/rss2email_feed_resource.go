package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &rss2emailFeedResource{}
	// _ resource.ResourceWithConfigure = &rss2emailFeedResource{}
)

func newRss2emailFeedResource() resource.Resource {
	return &rss2emailFeedResource{}
}

type rss2emailFeedResource struct {
	Id            types.String `tfsdk:"id"`
	Url           types.String `tfsdk:"url"`
	Name          types.String `tfsdk:"name"`
	LastCheck     types.String `tfsdk:"last_check"`
	LastDataFound types.String `tfsdk:"last_data_found"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	api_key       string
}

type createFeedRequest struct {
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// Create implements resource.Resource.
func (d *rss2emailFeedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rss2emailFeedResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := createFeedRequest{
		Url:     plan.Url.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create feed, got error: %s", err))
		return
	}

	url := "https://rss2email.dtz.rocks/api/2021-02-01/rss2email/feed"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create feed, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create feed, status code: %d", httpResp.StatusCode))
		return
	}

	var createResp rss2emailFeedResponse
	err = json.NewDecoder(httpResp.Body).Decode(&createResp)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse create response, got error: %s", err))
		return
	}

	// Set the resource state
	plan.Id = types.StringValue(createResp.Id)
	plan.Name = types.StringValue(createResp.Name)
	plan.LastCheck = types.StringValue(createResp.LastCheck)
	plan.LastDataFound = types.StringValue(createResp.LastDataFound)
	plan.Enabled = types.BoolValue(createResp.Enabled)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (d *rss2emailFeedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "rss2emailFeedResource delete")
	var cfg rss2emailFeedResource
	req.State.Get(ctx, &cfg)
	var url = fmt.Sprintf("https://rss2email.dtz.rocks/api/2021-02-01/rss2email/feed/%s", cfg.Id.ValueString())
	tflog.Info(ctx, "query API "+url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
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
}

// Update implements resource.Resource.
func (d *rss2emailFeedResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("unimplemented")
}

type rss2emailFeedResponse2 struct {
	Id            string `json:"id"`
	Url           string `json:"url"`
	LastCheck     string `json:"lastCheck"`
	LastDataFound string `json:"lastDataFound"`
	Enabled       bool   `json:"enabled"`
	Name          string `json:"name"`
}

func (d *rss2emailFeedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rss2email_feed"
}

func (d *rss2emailFeedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"url": schema.StringAttribute{
				Required: true,
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
				Optional: true,
			},
		},
	}
}

func (d *rss2emailFeedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *rss2emailFeedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rss2emailFeedResource
	var config_data rss2emailFeedResource
	// tflog.Info(ctx, fmt.Sprintf("read config %+v", req))
	resp.Diagnostics.Append(req.State.Get(ctx, &config_data)...)

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

	var resp_type rss2emailFeedResponse2
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
