package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &activitytypeDataSource{}
	_ datasource.DataSourceWithConfigure = &activitytypeDataSource{}
)

// NewactivitytypeDataSource is a helper function to simplify the provider implementation.
func NewactivitytypeDataSource() datasource.DataSource {
	return &activitytypeDataSource{}
}

// activitytypeDataSource is the data source implementation.
type activitytypeDataSource struct {
	client *graphql.Client
}

// activityType maps coffees schema data.
type activitytypeDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *activitytypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activity_type"
}

// Configure adds the provider configured client to the datasource.
func (d *activitytypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *activitytypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search an existing activity type in ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The identifier of the activity type.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the activity type to be returned.",
				Required:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *activitytypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var plan activitytypeDataSourceModel
	var state activitytypeDataSourceModel
	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const queryActivityType = `query ActivityType ($name: String){
		activityType(where: {activity: {_eq: $name}}) {
			id
			activity
		}
	}`
	request := graphql.NewRequest(queryActivityType)
	request.Var("name", plan.Name.ValueString())
	var respData map[string]interface{}
	if err := d.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter activity types",
			"Could not read Ghostwriter activity types: "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	activityTypes := respData["activityType"].([]interface{})
	if len(activityTypes) == 1 {
		activityType := activityTypes[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(activityType["id"].(float64)))
		state.Name = types.StringValue(activityType["activity"].(string))

		// Set state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter activity types",
			"Could not read Ghostwriter activity types: Activity type not found",
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
