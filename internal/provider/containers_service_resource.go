package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &containersServiceResource{}
)

func newContainersServiceResource() resource.Resource {
	return &containersServiceResource{}
}

type containersServiceResource struct {
	Id                types.String `tfsdk:"id"`
	Prefix            types.String `tfsdk:"prefix"`
	ContainerImage    types.String `tfsdk:"container_image"`
	ContainerPullUser types.String `tfsdk:"container_pull_user"`
	ContainerPullPwd  types.String `tfsdk:"container_pull_pwd"`
	EnvVariables      types.Map    `tfsdk:"env_variables"`
	Login             types.Object `tfsdk:"login"`
	api_key           string
}

type containersServiceResponse struct {
	ContextId         string            `json:"contextId"`
	ServiceId         string            `json:"serviceId"`
	Created           string            `json:"created"`
	Prefix            string            `json:"prefix"`
	ContainerImage    string            `json:"containerImage"`
	ContainerPullUser *string           `json:"containerPullUser"`
	ContainerPullPwd  *string           `json:"containerPullPwd"`
	EnvVariables      map[string]string `json:"envVariables"`
	Login             *struct {
		ProviderName string `json:"providerName"`
	} `json:"login"`
}

type createServiceRequest struct {
	Prefix            string            `json:"prefix"`
	ContainerImage    string            `json:"containerImage"`
	ContainerPullUser string            `json:"containerPullUser,omitempty"`
	ContainerPullPwd  string            `json:"containerPullPwd,omitempty"`
	EnvVariables      map[string]string `json:"envVariables,omitempty"`
	Login             *struct {
		ProviderName string `json:"providerName"`
	} `json:"login,omitempty"`
}

func (d *containersServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers_service"
}

func (d *containersServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"prefix": schema.StringAttribute{
				Required: true,
			},
			"container_image": schema.StringAttribute{
				Required: true,
			},
			"container_pull_user": schema.StringAttribute{
				Optional: true,
			},
			"container_pull_pwd": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"env_variables": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"login": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"provider_name": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (d *containersServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan containersServiceResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createService := createServiceRequest{
		Prefix:            plan.Prefix.ValueString(),
		ContainerImage:    plan.ContainerImage.ValueString(),
		ContainerPullUser: plan.ContainerPullUser.ValueString(),
		ContainerPullPwd:  plan.ContainerPullPwd.ValueString(),
	}

	if !plan.EnvVariables.IsNull() {
		envVars := make(map[string]string)
		diags = plan.EnvVariables.ElementsAs(ctx, &envVars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createService.EnvVariables = envVars
	}

	if !plan.Login.IsNull() {
		var login struct {
			ProviderName string `tfsdk:"provider_name"`
		}
		diags = plan.Login.As(ctx, &login, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createService.Login = &struct {
			ProviderName string `json:"providerName"`
		}{
			ProviderName: login.ProviderName,
		}
	}

	body, err := json.Marshal(createService)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create service, got error: %s", err))
		return
	}

	url := "https://containers.dtz.rocks/api/2021-02-21/service"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending create service request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create service, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, res.Body)

	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("status: %d, body: %s", res.StatusCode, string(resp_body[:])))

	var serviceResponse containersServiceResponse
	err = json.Unmarshal(resp_body, &serviceResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.Id = types.StringValue(serviceResponse.ServiceId)
	plan.Prefix = types.StringValue(serviceResponse.Prefix)
	plan.ContainerImage = types.StringValue(serviceResponse.ContainerImage)
	plan.ContainerPullUser = types.StringPointerValue(serviceResponse.ContainerPullUser)
	plan.ContainerPullPwd = types.StringPointerValue(serviceResponse.ContainerPullPwd)

	envVars, diags := types.MapValueFrom(ctx, types.StringType, serviceResponse.EnvVariables)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.EnvVariables = envVars

	if serviceResponse.Login != nil {
		login, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"provider_name": types.StringType,
		}, map[string]attr.Value{
			"provider_name": types.StringValue(serviceResponse.Login.ProviderName),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Login = login
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *containersServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containersServiceResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/service/%s", state.Id.ValueString())
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending read service request", map[string]interface{}{
		"url":    url,
		"method": http.MethodGet,
	})

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service, got error: %s", err))
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

	var serviceResponse containersServiceResponse
	err = json.Unmarshal(body, &serviceResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	state.Id = types.StringValue(serviceResponse.ServiceId)
	state.Prefix = types.StringValue(serviceResponse.Prefix)
	state.ContainerImage = types.StringValue(serviceResponse.ContainerImage)
	state.ContainerPullUser = types.StringPointerValue(serviceResponse.ContainerPullUser)
	state.ContainerPullPwd = types.StringPointerValue(serviceResponse.ContainerPullPwd)

	envVars, diags := types.MapValueFrom(ctx, types.StringType, serviceResponse.EnvVariables)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.EnvVariables = envVars

	if serviceResponse.Login != nil {
		login, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"provider_name": types.StringType,
		}, map[string]attr.Value{
			"provider_name": types.StringValue(serviceResponse.Login.ProviderName),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Login = login
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *containersServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containersServiceResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateService := createServiceRequest{
		Prefix:            plan.Prefix.ValueString(),
		ContainerImage:    plan.ContainerImage.ValueString(),
		ContainerPullUser: plan.ContainerPullUser.ValueString(),
		ContainerPullPwd:  plan.ContainerPullPwd.ValueString(),
	}

	if !plan.EnvVariables.IsNull() {
		envVars := make(map[string]string)
		diags = plan.EnvVariables.ElementsAs(ctx, &envVars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateService.EnvVariables = envVars
	}

	if !plan.Login.IsNull() {
		var login struct {
			ProviderName string `tfsdk:"provider_name"`
		}
		diags = plan.Login.As(ctx, &login, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateService.Login = &struct {
			ProviderName string `json:"providerName"`
		}{
			ProviderName: login.ProviderName,
		}
	}

	body, err := json.Marshal(updateService)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", err))
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/service/%s", plan.Id.ValueString())
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending update service request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, res.Body)

	if res.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update service, status code: %d", res.StatusCode))
		return
	}

	var serviceResponse containersServiceResponse
	err = json.NewDecoder(res.Body).Decode(&serviceResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.Id = types.StringValue(serviceResponse.ServiceId)
	plan.Prefix = types.StringValue(serviceResponse.Prefix)
	plan.ContainerImage = types.StringValue(serviceResponse.ContainerImage)
	plan.ContainerPullUser = types.StringPointerValue(serviceResponse.ContainerPullUser)
	plan.ContainerPullPwd = types.StringPointerValue(serviceResponse.ContainerPullPwd)

	envVars, diags := types.MapValueFrom(ctx, types.StringType, serviceResponse.EnvVariables)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.EnvVariables = envVars

	if serviceResponse.Login != nil {
		login, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"provider_name": types.StringType,
		}, map[string]attr.Value{
			"provider_name": types.StringValue(serviceResponse.Login.ProviderName),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Login = login
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *containersServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state containersServiceResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/service/%s", state.Id.ValueString())
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	tflog.Debug(ctx, "Sending delete service request", map[string]interface{}{
		"url":    url,
		"method": http.MethodDelete,
	})

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	if response.StatusCode == http.StatusNotFound {
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete service, status code: %d", response.StatusCode))
		return
	}
}

func (d *containersServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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