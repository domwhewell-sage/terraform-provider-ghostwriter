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
	_ datasource.DataSource              = &serverproviderDataSource{}
	_ datasource.DataSourceWithConfigure = &serverproviderDataSource{}
)

// NewserverproviderDataSource is a helper function to simplify the provider implementation.
func NewserverproviderDataSource() datasource.DataSource {
	return &serverproviderDataSource{}
}

// serverproviderDataSource is the data source implementation.
type serverproviderDataSource struct {
	client *graphql.Client
}

// activityType maps coffees schema data.
type serverproviderDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *serverproviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_provider"
}

// Configure adds the provider configured client to the datasource.
func (d *serverproviderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *serverproviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search an existing server provider in ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The identifier of the server provider.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the server provider to be returned.",
				Required:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *serverproviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var plan serverproviderDataSourceModel
	var state serverproviderDataSourceModel
	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const queryServerProvider = `query ServerProvider ($name: String){
		serverProvider(where: {serverProvider: {_eq: $name}}) {
			id
			serverProvider
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Querying server providers: %v", plan))
	request := graphql.NewRequest(queryServerProvider)
	request.Var("name", plan.Name.ValueString())
	var respData map[string]interface{}
	if err := d.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter server providers",
			"Could not read Ghostwriter server providers: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	// Overwrite items with refreshed state
	server_providers := respData["serverProvider"].([]interface{})
	if len(server_providers) == 1 {
		server_provider := server_providers[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(server_provider["id"].(float64)))
		state.Name = types.StringValue(server_provider["serverProvider"].(string))

		// Set state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter server providers",
			"Could not read Ghostwriter server providers: Server provider not found",
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
