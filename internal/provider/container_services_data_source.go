package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &containersDataSource{}
	_ datasource.DataSourceWithConfigure = &containersDataSource{}
)

// NewcontainersDataSource is a helper function to simplify the provider implementation.
func NewContainerServicesDataSource() datasource.DataSource {
	return &containersDataSource{}
}

// containersDataSource is the data source implementation.
type containersDataSource struct {
	client *Client
}

// Metadata returns the data source type name.
func (d *containersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_services"
}

// Schema defines the schema for the data source.
func (d *containersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"routing": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"prefix": schema.StringAttribute{
										Computed: true,
									},
									"service_definition": schema.SingleNestedAttribute{
										Computed: true,
										Attributes: map[string]schema.Attribute{
											"container_image": schema.StringAttribute{
												Computed: true,
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

// Configure adds the provider configured client to the data source.
func (d *containersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

// containersModel maps containers schema data.
type containerServiceModel struct {
	Domains []containersDomainsModel `tfsdk:"domains"`
}

// containersModel maps containers schema data.
type containersDomainsModel struct {
	Name    types.String             `tfsdk:"name"`
	Routing []containersRoutingModel `tfsdk:"routing"`
}

// containersModel maps containers schema data.
type containersRoutingModel struct {
	Prefix            types.String                     `tfsdk:"prefix"`
	ServiceDefinition containersServiceDefinitionModel `tfsdk:"service_definition"`
}

// containersModel maps containers schema data.
type containersServiceDefinitionModel struct {
	ContainerImage types.String `tfsdk:"container_image"`
}

// Read refreshes the Terraform state with the latest data.
func (d *containersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state containerServiceModel

	containersdomains, err := d.client.GetContainerServices(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DownToZero container services",
			err.Error(),
		)
		return
	}
	// Map response body to model
	for _, domain := range containersdomains {
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

		state.Domains = append(state.Domains, containersDomainsModel)
	}
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Getcontainers - Returns list of containers (no auth required)
func (c *Client) GetContainerServices(ctx context.Context) ([]ContainersDomains, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/service", c.ApiUrl), nil)
	if err != nil {
		return nil, err
	}

	status, body, err := c.doRequest(req)
	if err != nil || status != 200 {
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("status: %d, body: %s", status, string(body[:])))

	container_services := ContainerServices{}
	err = json.Unmarshal(body, &container_services)
	if err != nil {
		return nil, err
	}
	return container_services.Domains, nil
}

// Container -
type ContainerServices struct {
	Domains []ContainersDomains `json:"domains"`
}

// containers maps containers schema data.
type ContainersDomains struct {
	Name    string              `json:"name"`
	Routing []ContainersRouting `json:"routing"`
}

// containers maps containers schema data.
type ContainersRouting struct {
	Prefix            string                      `json:"prefix"`
	ServiceDefinition ContainersServiceDefinition `json:"serviceDefinition"`
}

// containers maps containers schema data.
type ContainersServiceDefinition struct {
	ContainerImage string `json:"containerImage"`
}
