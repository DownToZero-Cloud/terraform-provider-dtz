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
	_ resource.Resource = &containersDomainResource{}
)

func newContainersDomainResource() resource.Resource {
	return &containersDomainResource{}
}

type containersDomainResource struct {
	ContextId types.String `tfsdk:"context_id"`
	Name      types.String `tfsdk:"name"`
	Verified  types.Bool   `tfsdk:"verified"`
	Created   types.String `tfsdk:"created"`
	api_key   string
}

type containersDomainResponse struct {
	ContextId string `json:"contextId"`
	Name      string `json:"name"`
	Verified  bool   `json:"verified"`
	Created   string `json:"created"`
}

type createDomainRequest struct {
	Name string `json:"name"`
}

func (d *containersDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers_domain"
}

func (d *containersDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"context_id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
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

func (d *containersDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan containersDomainResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createDomain := createDomainRequest{
		Name: plan.Name.ValueString(),
	}

	body, err := json.Marshal(createDomain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create domain, got error: %s", err))
		return
	}

	url := "https://containers.dtz.rocks/api/2021-02-21/domain"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending create domain request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create domain, got error: %s", err))
		return
	}
	defer res.Body.Close()

	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		tflog.Error(ctx, "error reading")
		return
	}
	tflog.Info(ctx, fmt.Sprintf("status: %d, body: %s", res.StatusCode, string(resp_body[:])))

	var domainResponse containersDomainResponse
	err = json.Unmarshal(resp_body, &domainResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.ContextId = types.StringValue(domainResponse.ContextId)
	plan.Name = types.StringValue(domainResponse.Name)
	plan.Verified = types.BoolValue(domainResponse.Verified)
	plan.Created = types.StringValue(domainResponse.Created)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *containersDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containersDomainResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/domain/%s", state.Name.ValueString())
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending read domain request", map[string]interface{}{
		"url":    url,
		"method": http.MethodGet,
	})

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domain, got error: %s", err))
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}

	var domainResponse containersDomainResponse
	err = json.Unmarshal(body, &domainResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	state.ContextId = types.StringValue(domainResponse.ContextId)
	state.Name = types.StringValue(domainResponse.Name)
	state.Verified = types.BoolValue(domainResponse.Verified)
	state.Created = types.StringValue(domainResponse.Created)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *containersDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Create a new ReadResponse
	readResp := &resource.ReadResponse{
		State:       resp.State,
		Private:     resp.Private,
		Diagnostics: resp.Diagnostics,
	}

	// Call Read method
	d.Read(ctx, resource.ReadRequest{State: req.State}, readResp)

	// Copy relevant fields back to UpdateResponse
	resp.State = readResp.State
	resp.Private = readResp.Private
	resp.Diagnostics = readResp.Diagnostics
}

func (d *containersDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state containersDomainResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/domain/%s", state.Name.ValueString())
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete domain, got error: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete domain, status code: %d", response.StatusCode))
		return
	}
}

func (d *containersDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
