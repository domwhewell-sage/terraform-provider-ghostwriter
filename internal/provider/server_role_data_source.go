package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &serverroleDataSource{}
	_ datasource.DataSourceWithConfigure = &serverroleDataSource{}
)

// NewserverroleDataSource is a helper function to simplify the provider implementation.
func NewserverroleDataSource() datasource.DataSource {
	return &serverroleDataSource{}
}

// serverroleDataSource is the data source implementation.
type serverroleDataSource struct {
	client *graphql.Client
}

// activityType maps coffees schema data.
type serverroleDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *serverroleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_role"
}

// Configure adds the provider configured client to the datasource.
func (d *serverroleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*graphql.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *graphql.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Schema defines the schema for the data source.
func (d *serverroleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search an existing server role in ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The identifier of the server role.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the server role to be returned.",
				Required:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *serverroleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var plan serverroleDataSourceModel
	var state serverroleDataSourceModel
	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const queryServerRoles = `query ServerRoles ($name: String){
		serverRole(where: {serverRole: {_eq: $name}}) {
			id
			serverRole
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Querying server roles: %v", plan))
	request := graphql.NewRequest(queryServerRoles)
	request.Var("name", plan.Name.ValueString())
	var respData map[string]interface{}
	if err := d.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter server roles",
			"Could not read Ghostwriter server roles: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	// Overwrite items with refreshed state
	serverRoles := respData["serverRole"].([]interface{})
	if len(serverRoles) == 1 {
		serverRole := serverRoles[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(serverRole["id"].(float64)))
		state.Name = types.StringValue(serverRole["serverRole"].(string))

		// Set state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter server roles",
			"Could not read Ghostwriter server roles: Server role not found",
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
