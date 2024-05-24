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
	_ datasource.DataSource = &rss2emailProfileDataSource{}
)

func newRss2emailProfileDataSource() datasource.DataSource {
	return &rss2emailProfileDataSource{}
}

type rss2emailProfileDataSource struct {
	Email   types.String `tfsdk:"email" json:"email"`
	Subject types.String `tfsdk:"subject" json:"subject"`
	Body    types.String `tfsdk:"body" json:"body"`
	api_key string
}

func (d *rss2emailProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rss2email_profile"
}

func (d *rss2emailProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Computed: true,
			},
			"subject": schema.StringAttribute{
				Computed: true,
			},
			"body": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *rss2emailProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		tflog.Error(ctx, "configure: provider data is nil")
		return
	}
	dtz := req.ProviderData.(dtzProvider)
	d.api_key = dtz.ApiKey
}

func (d *rss2emailProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rss2emailProfileDataSource
	var url = "https://rss2email.dtz.rocks/api/2021-02-01/rss2email/profile"
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
	tflog.Info(ctx, fmt.Sprintf("rssProfileDataSource Read status: %d, body: %s", response.StatusCode, string(body[:])))

	var resp_type rss2emailProfileDataSource
	err = json.Unmarshal(body, &resp_type)
	if err != nil {
		tflog.Error(ctx, "error unmarshalling")
		return
	}
	tflog.Info(ctx, fmt.Sprintf("rssProfileDataSource Read response: %+v", resp_type))

	state.Email = resp_type.Email
	state.Subject = resp_type.Subject
	state.Body = resp_type.Body
	// set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
