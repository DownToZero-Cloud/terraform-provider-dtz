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
	_ resource.Resource = &rss2emailProfileResource{}
)

func newRss2emailProfileResource() resource.Resource {
	return &rss2emailProfileResource{}
}

type rss2emailProfileResource struct {
	Email   types.String `tfsdk:"email"`
	Subject types.String `tfsdk:"subject"`
	Body    types.String `tfsdk:"body"`
	api_key string
}

type rss2emailProfileResponse struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (d *rss2emailProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rss2email_profile"
}

func (d *rss2emailProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Required: true,
			},
			"subject": schema.StringAttribute{
				Optional: true,
			},
			"body": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *rss2emailProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rss2emailProfileResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createProfile := rss2emailProfileResource{
		Email:   plan.Email,
		Subject: plan.Subject,
		Body:    plan.Body,
	}

	body, err := json.Marshal(createProfile)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create profile, got error: %s", err))
		return
	}

	url := "https://rss2email.dtz.rocks/api/2021-02-01/rss2email/profile"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending create profile request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create profile, got error: %s", err))
		return
	}
	defer res.Body.Close()

	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}

	var profileResponse rss2emailProfileResponse
	err = json.Unmarshal(resp_body, &profileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.Email = types.StringValue(profileResponse.Email)
	plan.Subject = types.StringValue(profileResponse.Subject)
	plan.Body = types.StringValue(profileResponse.Body)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *rss2emailProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rss2emailProfileResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "https://rss2email.dtz.rocks/api/2021-02-01/rss2email/profile"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read profile, got error: %s", err))
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	var profileResponse rss2emailProfileResponse
	err = json.Unmarshal(body, &profileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	state.Email = types.StringValue(profileResponse.Email)
	state.Subject = types.StringValue(profileResponse.Subject)
	state.Body = types.StringValue(profileResponse.Body)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *rss2emailProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rss2emailProfileResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateProfile := rss2emailProfileResource{
		Email:   plan.Email,
		Subject: plan.Subject,
		Body:    plan.Body,
	}

	body, err := json.Marshal(updateProfile)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update profile, got error: %s", err))
		return
	}

	url := "https://rss2email.dtz.rocks/api/2021-02-01/rss2email/profile"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending update profile request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update profile, got error: %s", err))
		return
	}
	defer res.Body.Close()

	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}

	var profileResponse rss2emailProfileResponse
	err = json.Unmarshal(resp_body, &profileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.Email = types.StringValue(profileResponse.Email)
	plan.Subject = types.StringValue(profileResponse.Subject)
	plan.Body = types.StringValue(profileResponse.Body)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *rss2emailProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The API doesn't provide a delete endpoint for the profile, so we'll just remove it from the state
	resp.State.RemoveResource(ctx)
}

func (d *rss2emailProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
