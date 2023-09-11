package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &containersResource{}
	_ resource.ResourceWithConfigure = &containersResource{}
)

// NewContainerServicesResource is a helper function to simplify the provider implementation.
func NewContainerServicesResource() resource.Resource {
	return &containersResource{}
}

// containersResource is the resource implementation.
type containersResource struct {
	client *Client
}

// Metadata returns the resource type name.
func (r *containersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_service"
}

// Configure adds the provider configured client to the resource.
func (r *containersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the data source.
func (r *containersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"routing": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"prefix": schema.StringAttribute{
										Required: true,
									},
									"service_definition": schema.SingleNestedAttribute{
										Required: true,
										Attributes: map[string]schema.Attribute{
											"container_image": schema.StringAttribute{
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create a new resource.
func (r *containersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan containerServiceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var items []ContainersDomains
	for _, item := range plan.Domains {
		var containersDomains ContainersDomains
		containersDomains.Name = item.Name.ValueString()

		for _, routing := range item.Routing {
			var containersRouting ContainersRouting
			containersRouting.Prefix = routing.Prefix.ValueString()

			var serviceDefinition ContainersServiceDefinition
			serviceDefinition.ContainerImage = routing.ServiceDefinition.ContainerImage.ValueString()

			containersRouting.ServiceDefinition = serviceDefinition
			containersDomains.Routing = append(containersDomains.Routing, containersRouting)

		}

		items = append(items, containersDomains)
	}
	var wrapper ContainerServices
	wrapper.Domains = append(wrapper.Domains, items...)

	// Create new containerservice

	new_domains, err := r.client.CreateDomain(ctx, wrapper)
	tflog.Error(ctx, err.Error())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating containerservice",
			"Could not create containerservice, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to model
	for _, domain := range new_domains {
		var containersDomainsModel containersDomainsModel
		containersDomainsModel.Name = types.StringValue(domain.Name)

		for _, routing := range domain.Routing {
			var containersRoutingModel containersRoutingModel
			containersRoutingModel.Prefix = types.StringValue(routing.Prefix)

			var serviceDefinition containersServiceDefinitionModel
			serviceDefinition.ContainerImage = types.StringValue(routing.ServiceDefinition.ContainerImage)

			containersRoutingModel.ServiceDefinition = serviceDefinition
			containersDomainsModel.Routing = append(containersDomainsModel.Routing, containersRoutingModel)
		}

		plan.Domains = append(plan.Domains, containersDomainsModel)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *containersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *containersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *containersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// CreateDomain - Create new containerservice
func (c *Client) CreateDomain(ctx context.Context, containerServices ContainerServices) ([]ContainersDomains, error) {
	rb, err := json.Marshal(containerServices)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, string(rb[:]))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/service", c.ApiUrl), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	// var body []byte
	status, body, err := c.doRequest(req)
	tflog.Debug(ctx, fmt.Sprintf("status: %d, body: %s", status, string(body[:])))
	// for {
	// 	var status int
	// 	status, body, err = c.doRequest(req)
	// 	if b := string(body[:]); status == 500 && b == "\"issue certificate failed\"" {
	// 		tflog.Info(ctx, "Waiting for Certificate Request to be done...")
	// 		tflog.Debug(ctx, fmt.Sprintf("status: %d, body: %s", status, string(body[:])))
	// 		time.Sleep(10 * time.Second)
	// 		continue
	// 	} else {
	// 		tflog.Debug(ctx, fmt.Sprintf("+++++++++++status: %d, body: %s", status, string(body[:])))
	// 		break
	// 	}
	// }

	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("%+v", string(body[:])))

	container_services := ContainerServices{}
	err = json.Unmarshal(body, &container_services)
	if err != nil {
		return nil, err
	}
	return container_services.Domains, nil
}
