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
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

// NewprojectDataSource is a helper function to simplify the provider implementation.
func NewprojectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// projectDataSource is the data source implementation.
type projectDataSource struct {
	client *graphql.Client
}

// projectType maps coffees schema data.
type projectDataSourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ClientID      types.Int64  `tfsdk:"client_id"`
	ProjectTypeID types.Int64  `tfsdk:"project_type_id"`
	OperatorID    types.Int64  `tfsdk:"operator_id"`
	CodeName      types.String `tfsdk:"code_name"`
	Complete      types.Bool   `tfsdk:"complete"`
	StartDate     types.String `tfsdk:"start_date"`
	StartTime     types.String `tfsdk:"start_time"`
	EndDate       types.String `tfsdk:"end_date"`
	EndTime       types.String `tfsdk:"end_time"`
	Timezone      types.String `tfsdk:"timezone"`
	Note          types.String `tfsdk:"note"`
	SlackChannel  types.String `tfsdk:"slack_channel"`
}

// Metadata returns the data source type name.
func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Configure adds the provider configured client to the datasource.
func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search an existing project in ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The identifier of the project.",
				Computed:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "The ID of the client",
				Computed:    true,
			},
			"project_type_id": schema.StringAttribute{
				Description: "The ID of the project type",
				Computed:    true,
			},
			"operator_id": schema.StringAttribute{
				Description: "The ID of the assigned operator",
				Computed:    true,
			},
			"code_name": schema.StringAttribute{
				Description: "The project codename",
				Required:    true,
			},
			"complete": schema.StringAttribute{
				Description: "If the project is complete",
				Computed:    true,
			},
			"start_date": schema.StringAttribute{
				Description: "The start date of the project",
				Computed:    true,
			},
			"start_time": schema.StringAttribute{
				Description: "The start time of the project",
				Computed:    true,
			},
			"end_date": schema.StringAttribute{
				Description: "The end date of the project",
				Computed:    true,
			},
			"end_time": schema.StringAttribute{
				Description: "The end time of the project",
				Computed:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "The projects timezone",
				Computed:    true,
			},
			"note": schema.StringAttribute{
				Description: "The note asociated with the project",
				Computed:    true,
			},
			"slack_channel": schema.StringAttribute{
				Description: "The projects slack channel",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var plan projectDataSourceModel
	var state projectDataSourceModel
	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const queryProject = `query Project($name: String) {
		project(where: {codename: {_eq: $name}}) {
			id
			clientId
			operatorId
			projectTypeId
			codename
			complete
			startDate
			startTime
			endDate
			endTime
			timezone
			note
			slackChannel
		}
	}`
	request := graphql.NewRequest(queryProject)
	request.Var("name", plan.CodeName.ValueString())
	var respData map[string]interface{}
	if err := d.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter project",
			"Could not read Ghostwriter project: "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	projects := respData["project"].([]interface{})
	if len(projects) == 1 {
		project := projects[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(project["id"].(float64)))
		state.ClientID = types.Int64Value(int64(project["clientId"].(float64)))
		state.ProjectTypeID = types.Int64Value(int64(project["projectTypeId"].(float64)))
		state.OperatorID = types.Int64Value(int64(project["operatorId"].(float64)))
		state.CodeName = types.StringValue(project["codename"].(string))
		state.Complete = types.BoolValue(project["complete"].(bool))
		state.StartDate = types.StringValue(project["startDate"].(string))
		state.StartTime = types.StringValue(project["startTime"].(string))
		state.EndDate = types.StringValue(project["endDate"].(string))
		state.EndTime = types.StringValue(project["endTime"].(string))
		state.Timezone = types.StringValue(project["timezone"].(string))
		state.Note = types.StringValue(project["note"].(string))
		state.SlackChannel = types.StringValue(project["slackChannel"].(string))

		// Set state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter projects",
			"Could not read Ghostwriter projects: Project not found or multiple projects found with the same name.",
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
