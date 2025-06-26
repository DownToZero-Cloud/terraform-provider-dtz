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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &identityApikeyResource{}
)

func newIdentityApikeyResource() resource.Resource {
	return &identityApikeyResource{}
}

type identityApikeyResource struct {
	Apikey    types.String `tfsdk:"apikey"`
	Alias     types.String `tfsdk:"alias"`
	ContextId types.String `tfsdk:"context_id"`
	api_key   string
}

type createApikeyRequest struct {
	Alias     string `json:"alias"`
	ContextId string `json:"contextId"`
}

type authenticationResponse struct {
	IdentityId string `json:"identityId"`
	UserAuth   []struct {
	} `json:"userAuth"`
	ApiKeyAuth []struct {
		ApiKey           string `json:"apiKey"`
		DefaultContextId string `json:"defaultContextId"`
		Alias            string `json:"alias"`
	} `json:"apiKeyAuth"`
	OauthAuth []struct {
	} `json:"oauthAuth"`
}

func (d *identityApikeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_apikey"
}

func (d *identityApikeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"apikey": schema.StringAttribute{
				Computed: true,
			},
			"alias": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"context_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (d *identityApikeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan identityApikeyResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createApikey := createApikeyRequest{
		Alias:     plan.Alias.ValueString(),
		ContextId: plan.ContextId.ValueString(),
	}

	body, err := json.Marshal(createApikey)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create apikey, got error: %s", err))
		return
	}

	url := "https://identity.dtz.rocks/api/2021-02-21/me/identity/apikey"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	// Add debug log before sending the request
	tflog.Debug(ctx, "Sending create apikey request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create apikey, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, res.Body)

	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		tflog.Error(ctx, "error reading")
		return
	}

	jobResponse := string(resp_body[:])
	tflog.Info(ctx, fmt.Sprintf("status: %d, body: %s", res.StatusCode, jobResponse))

	plan.Apikey = types.StringValue(jobResponse)
	plan.Alias = types.StringValue(createApikey.Alias)
	plan.ContextId = types.StringValue(createApikey.ContextId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *identityApikeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state identityApikeyResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "https://identity.dtz.rocks/api/2021-02-21/authentication"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	// Add debug log before sending the request
	tflog.Debug(ctx, "Sending read job request", map[string]interface{}{
		"url":    url,
		"method": http.MethodGet,
	})

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read authentications, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}

	if response.StatusCode == http.StatusNotFound {
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	var authenticationResponse authenticationResponse
	err = json.Unmarshal(body, &authenticationResponse)
	if err != nil {
		statusCode := response.StatusCode
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s\nstatus code: %d, body: %s", err, statusCode, string(body)))
		return
	}
	var result identityApikeyResource
	for _, auth := range authenticationResponse.ApiKeyAuth {
		if auth.ApiKey == state.Apikey.ValueString() {
			result.Apikey = types.StringValue(auth.ApiKey)
			result.Alias = types.StringValue(auth.Alias)
			result.ContextId = types.StringValue(auth.DefaultContextId)
		}
	}

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (d *identityApikeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan identityApikeyResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: Implement update

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *identityApikeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state identityApikeyResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://identity.dtz.rocks/api/2021-02-21/apikey/%s", state.Apikey.ValueString())
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete apikey, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	if response.StatusCode == http.StatusNotFound {
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete apikey, status code: %d", response.StatusCode))
		return
	}

	// Add debug log after receiving the response
	body, _ := io.ReadAll(response.Body)
	tflog.Debug(ctx, "Received delete apikey response", map[string]interface{}{
		"statusCode": response.StatusCode,
		"body":       string(body),
	})
}

func (d *identityApikeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	dtz, ok := req.ProviderData.(dtzProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected dtzProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.api_key = dtz.ApiKey
}
