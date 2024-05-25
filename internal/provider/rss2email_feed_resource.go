package provider

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"

// 	"github.com/hashicorp/terraform-plugin-framework/resource"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/hashicorp/terraform-plugin-log/tflog"
// )

// var (
// 	_ resource.Resource = &rss2emailFeedResource{}
// 	// _ resource.ResourceWithConfigure = &rss2emailFeedResource{}
// )

// func newRss2emailFeedResource() resource.Resource {
// 	return &rss2emailFeedResource{}
// }

// type rss2emailFeedResource struct {
// 	Id            types.String `tfsdk:"id"`
// 	Url           types.String `tfsdk:"url"`
// 	Name          types.String `tfsdk:"name"`
// 	LastCheck     types.String `tfsdk:"last_check"`
// 	LastDataFound types.String `tfsdk:"last_data_found"`
// 	Enabled       types.Bool   `tfsdk:"enabled"`
// 	api_key       string
// }

// // Create implements resource.Resource.
// func (d *rss2emailFeedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
// 	tflog.Info(ctx, "rss2emailFeedResource create")
// 	panic("unimplemented")
// }

// // Delete implements resource.Resource.
// func (d *rss2emailFeedResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
// 	panic("unimplemented")
// }

// // Update implements resource.Resource.
// func (d *rss2emailFeedResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
// 	panic("unimplemented")
// }

// type rss2emailFeedResponse2 struct {
// 	Id            string `json:"id"`
// 	Url           string `json:"url"`
// 	LastCheck     string `json:"lastCheck"`
// 	LastDataFound string `json:"lastDataFound"`
// 	Enabled       bool   `json:"enabled"`
// 	Name          string `json:"name"`
// }

// func (d *rss2emailFeedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = req.ProviderTypeName + "_rss2email_feed"
// }

// func (d *rss2emailFeedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
// 	resp.Schema = schema.Schema{
// 		Attributes: map[string]schema.Attribute{
// 			"id": schema.StringAttribute{
// 				Computed: true,
// 			},
// 			"url": schema.StringAttribute{
// 				Required: true,
// 			},
// 			"name": schema.StringAttribute{
// 				Computed: true,
// 			},
// 			"last_check": schema.StringAttribute{
// 				Computed: true,
// 			},
// 			"last_data_found": schema.StringAttribute{
// 				Computed: true,
// 			},
// 			"enabled": schema.BoolAttribute{
// 				Computed: true,
// 				Optional: true,
// 			},
// 		},
// 	}
// }

// func (d *rss2emailFeedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
// 	if req.ProviderData == nil {
// 		tflog.Error(ctx, "configure: provider data is nil")
// 		return
// 	}
// 	dtz := req.ProviderData.(dtzProvider)
// 	d.api_key = dtz.ApiKey
// }

// func (d *rss2emailFeedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
// 	var state rss2emailFeedResource
// 	var config_data rss2emailFeedResource
// 	// tflog.Info(ctx, fmt.Sprintf("read config %+v", req))
// 	resp.Diagnostics.Append(req.State.Get(ctx, &config_data)...)

// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	tflog.Info(ctx, fmt.Sprintf("read data %+v", config_data))
// 	var feed_id = config_data.Id
// 	var url = fmt.Sprintf("https://rss2email.dtz.rocks/api/2021-02-01/rss2email/feed/%s", feed_id.ValueString())
// 	tflog.Info(ctx, "query API "+url)
// 	request, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		tflog.Error(ctx, "error retrieving")
// 		return
// 	}
// 	request.Header.Set("X-API-KEY", d.api_key)
// 	client := &http.Client{}
// 	response, err := client.Do(request)
// 	if err != nil {
// 		tflog.Error(ctx, "error fetching")
// 		return
// 	}
// 	defer response.Body.Close()
// 	body, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		tflog.Error(ctx, "error reading")
// 		return
// 	}
// 	tflog.Info(ctx, fmt.Sprintf("rssFeedDataSource Read status: %d, body: %s", response.StatusCode, string(body[:])))

// 	var resp_type rss2emailFeedResponse
// 	err = json.Unmarshal(body, &resp_type)
// 	if err != nil {
// 		tflog.Error(ctx, "error unmarshalling")
// 		return
// 	}
// 	tflog.Info(ctx, fmt.Sprintf("rssFeedDataSource Read response: %+v", resp_type))

// 	state.Id = types.StringValue(resp_type.Id)
// 	state.Url = types.StringValue(resp_type.Url)
// 	state.Name = types.StringValue(resp_type.Name)
// 	state.Enabled = types.BoolValue(resp_type.Enabled)
// 	state.LastCheck = types.StringValue(resp_type.LastDataFound)
// 	state.LastDataFound = types.StringValue(resp_type.LastDataFound)
// 	// set state
// 	diags := resp.State.Set(ctx, &state)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// }
