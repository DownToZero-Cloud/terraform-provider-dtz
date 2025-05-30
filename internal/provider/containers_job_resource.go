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
	_ resource.Resource = &containersJobResource{}
)

func newContainersJobResource() resource.Resource {
	return &containersJobResource{}
}

type containersJobResource struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ContainerImage    types.String `tfsdk:"container_image"`
	ContainerPullUser types.String `tfsdk:"container_pull_user"`
	ContainerPullPwd  types.String `tfsdk:"container_pull_pwd"`
	ScheduleType      types.String `tfsdk:"schedule_type"`
	ScheduleRepeat    types.String `tfsdk:"schedule_repeat"`
	ScheduleCron      types.String `tfsdk:"schedule_cron"`
	api_key           string
}

type containersJobResponse struct {
	Id                string  `json:"id"`
	Name              string  `json:"name"`
	ContainerImage    string  `json:"containerImage"`
	ContainerPullUser *string `json:"containerPullUser"`
	ContainerPullPwd  *string `json:"containerPullPwd"`
	ScheduleType      string  `json:"scheduleType"`
	ScheduleRepeat    *string `json:"scheduleRepeat"`
	ScheduleCron      *string `json:"scheduleCron"`
}

type createJobRequest struct {
	Name              string `json:"name"`
	ContainerImage    string `json:"containerImage"`
	ContainerPullUser string `json:"containerPullUser,omitempty"`
	ContainerPullPwd  string `json:"containerPullPwd,omitempty"`
	ScheduleType      string `json:"scheduleType"`
	ScheduleCron      string `json:"scheduleCron,omitempty"`
	ScheduleRepeat    string `json:"scheduleRepeat,omitempty"`
}

func (d *containersJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers_job"
}

func (d *containersJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
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
			"schedule_type": schema.StringAttribute{
				Required: true,
			},
			"schedule_repeat": schema.StringAttribute{
				Optional: true,
			},
			"schedule_cron": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *containersJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan containersJobResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createJob := createJobRequest{
		Name:              plan.Name.ValueString(),
		ContainerImage:    plan.ContainerImage.ValueString(),
		ContainerPullUser: plan.ContainerPullUser.ValueString(),
		ContainerPullPwd:  plan.ContainerPullPwd.ValueString(),
		ScheduleType:      plan.ScheduleType.ValueString(),
		ScheduleCron:      plan.ScheduleCron.ValueString(),
		ScheduleRepeat:    plan.ScheduleRepeat.ValueString(),
	}

	body, err := json.Marshal(createJob)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create job, got error: %s", err))
		return
	}

	url := "https://containers.dtz.rocks/api/2021-02-21/job"
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	// Add debug log before sending the request
	tflog.Debug(ctx, "Sending create job request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create job, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, res.Body)

	resp_body, err := io.ReadAll(res.Body)
	if err != nil {
		tflog.Error(ctx, "error reading")
		return
	}
	tflog.Info(ctx, fmt.Sprintf("status: %d, body: %s", res.StatusCode, string(resp_body[:])))

	var jobResponse containersJobResponse
	err = json.Unmarshal(resp_body, &jobResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.Id = types.StringValue(jobResponse.Id)
	plan.Name = types.StringValue(jobResponse.Name)
	plan.ContainerImage = types.StringValue(jobResponse.ContainerImage)
	plan.ContainerPullUser = types.StringPointerValue(jobResponse.ContainerPullUser)
	plan.ContainerPullPwd = types.StringPointerValue(jobResponse.ContainerPullPwd)
	plan.ScheduleType = types.StringValue(jobResponse.ScheduleType)
	plan.ScheduleRepeat = types.StringPointerValue(jobResponse.ScheduleRepeat)
	plan.ScheduleCron = types.StringPointerValue(jobResponse.ScheduleCron)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *containersJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containersJobResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/job/%s", state.Id.ValueString())
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read job, got error: %s", err))
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

	var jobResponse containersJobResponse
	err = json.Unmarshal(body, &jobResponse)
	if err != nil {
		statusCode := response.StatusCode
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s\nstatus code: %d, body: %s", err, statusCode, string(body)))
		return
	}
	var result containersJobResource
	result.Id = types.StringValue(jobResponse.Id)
	result.Name = types.StringValue(jobResponse.Name)
	result.ContainerImage = types.StringValue(jobResponse.ContainerImage)
	result.ContainerPullUser = types.StringPointerValue(jobResponse.ContainerPullUser)
	result.ContainerPullPwd = types.StringPointerValue(jobResponse.ContainerPullPwd)
	result.ScheduleType = types.StringValue(jobResponse.ScheduleType)
	result.ScheduleRepeat = types.StringPointerValue(jobResponse.ScheduleRepeat)
	result.ScheduleCron = types.StringPointerValue(jobResponse.ScheduleCron)

	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (d *containersJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containersJobResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateJob := createJobRequest{
		Name:              plan.Name.ValueString(),
		ContainerImage:    plan.ContainerImage.ValueString(),
		ContainerPullUser: plan.ContainerPullUser.ValueString(),
		ContainerPullPwd:  plan.ContainerPullPwd.ValueString(),
		ScheduleType:      plan.ScheduleType.ValueString(),
		ScheduleCron:      plan.ScheduleCron.ValueString(),
		ScheduleRepeat:    plan.ScheduleRepeat.ValueString(),
	}

	body, err := json.Marshal(updateJob)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update job, got error: %s", err))
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/job/%s", plan.Id.ValueString())
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", d.api_key)

	// Add debug log before sending the request
	tflog.Debug(ctx, "Sending update job request", map[string]interface{}{
		"url":    url,
		"method": http.MethodPost,
		"body":   string(body),
	})

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update job, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, res.Body)

	if res.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update job, status code: %d", res.StatusCode))
		return
	}

	var jobResponse containersJobResponse
	err = json.NewDecoder(res.Body).Decode(&jobResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	plan.Id = types.StringValue(jobResponse.Id)
	plan.Name = types.StringValue(jobResponse.Name)
	plan.ContainerImage = types.StringValue(jobResponse.ContainerImage)
	plan.ContainerPullUser = types.StringPointerValue(jobResponse.ContainerPullUser)
	plan.ContainerPullPwd = types.StringPointerValue(jobResponse.ContainerPullPwd)
	plan.ScheduleType = types.StringValue(jobResponse.ScheduleType)
	plan.ScheduleRepeat = types.StringPointerValue(jobResponse.ScheduleRepeat)
	plan.ScheduleCron = types.StringPointerValue(jobResponse.ScheduleCron)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	// Add debug log after receiving the response
	body, _ = io.ReadAll(res.Body)
	tflog.Debug(ctx, "Received update job response", map[string]interface{}{
		"statusCode": res.StatusCode,
		"body":       string(body),
	})
}

func (d *containersJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state containersJobResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://containers.dtz.rocks/api/2021-02-21/job/%s", state.Id.ValueString())
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	request.Header.Set("X-API-KEY", d.api_key)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete job, got error: %s", err))
		return
	}
	defer deferredCloseResponseBody(ctx, response.Body)

	if response.StatusCode == http.StatusNotFound {
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete job, status code: %d", response.StatusCode))
		return
	}

	// Add debug log after receiving the response
	body, _ := io.ReadAll(response.Body)
	tflog.Debug(ctx, "Received delete job response", map[string]interface{}{
		"statusCode": response.StatusCode,
		"body":       string(body),
	})
}

func (d *containersJobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
