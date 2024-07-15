package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &oplogResource{}
	_ resource.ResourceWithConfigure   = &oplogResource{}
	_ resource.ResourceWithImportState = &oplogResource{}
)

// NewoplogResource is a helper function to simplify the provider implementation.
func NewoplogResource() resource.Resource {
	return &oplogResource{}
}

// oplogResource is the resource implementation.
type oplogResource struct {
	client *graphql.Client
}

// orderResourceModel maps the resource schema data.
type oplogResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.Int64  `tfsdk:"project_id"`
	ForceDelete types.Bool   `tfsdk:"force_delete"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *oplogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oplog"
}

// Configure adds the provider configured client to the resource.
func (r *oplogResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Schema defines the schema for the resource.
func (r *oplogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Create an operations log.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the oplog.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the operation log",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"project_id": schema.Int64Attribute{
				Description: "The unique identifier of the project the oplog should be created for.",
				Required:    true,
			},
			"force_delete": schema.BoolAttribute{
				Description: "If false, will not be deleted from the ghostwriter instance when not managed by terraform. If true, the oplog will be hard-deleted from the ghostwriter instance. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// ImportState imports the resource state from Terraform state.
func (r *oplogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	tflog.Debug(ctx, fmt.Sprintf("Importing oplog resource ID: %s", req.ID))
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Import ID",
			"Could not parse import ID: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	force_delete := types.BoolValue(false)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("force_delete"), &force_delete)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *oplogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oplogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const insertoplog = `mutation InsertOplog ($name: String, $project_id: bigint){
		insert_oplog(objects: {name: $name, projectId: $project_id}) {
			returning {
				id
				name
				projectId
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Creating oplog: %v", plan))
	request := graphql.NewRequest(insertoplog)
	request.Var("name", plan.Name.ValueString())
	request.Var("project_id", plan.ProjectID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error creating oplog",
			"Could not create oplog, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	oplogs := respData["insert_oplog"].(map[string]interface{})["returning"].([]interface{})
	if len(oplogs) == 1 {
		oplog := oplogs[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(oplog["id"].(float64)))
		plan.Name = types.StringValue(oplog["name"].(string))
		plan.ProjectID = types.Int64Value(int64(oplog["projectId"].(float64)))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error creating oplog",
			"Could not create oplog: Oplog not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *oplogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state oplogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const queryoplog = `query QueryOplog ($id: bigint){
		oplog(where: {id: {_eq: $id}}) {
			id
			name
			projectId
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Reading oplog: %v", state.ID))
	request := graphql.NewRequest(queryoplog)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Oplog",
			"Could not read Ghostwriter oplog ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	oplogs := respData["oplog"].([]interface{})
	if len(oplogs) == 1 {
		oplog := oplogs[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(oplog["id"].(float64)))
		state.Name = types.StringValue(oplog["name"].(string))
		state.ProjectID = types.Int64Value(int64(oplog["projectId"].(float64)))

		// Set refreshed state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Oplog",
			"Could not read Ghostwriter oplog ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": oplog not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *oplogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan oplogResourceModel
	var state oplogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const updateoplog = `mutation UpdateOplog ($id: bigint, $name: String, $project_id: bigint){
		update_oplog(where: {id: {_eq: $id}}, _set: {name: $name, projectId: $project_id}) {
			returning {
				id
				name
				projectId
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Updating oplog: %v", plan))
	request := graphql.NewRequest(updateoplog)
	request.Var("id", state.ID.ValueInt64())
	request.Var("name", plan.Name.ValueString())
	request.Var("project_id", plan.ProjectID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Oplog",
			"Could not update oplog ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	updated_oplogs := respData["update_oplog"].(map[string]interface{})["returning"].([]interface{})
	if len(updated_oplogs) == 1 {
		oplog := updated_oplogs[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(oplog["id"].(float64)))
		plan.Name = types.StringValue(oplog["name"].(string))
		plan.ProjectID = types.Int64Value(int64(oplog["projectId"].(float64)))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Oplog",
			"Could not update oplog ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": oplog not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *oplogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state oplogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ForceDelete.ValueBool() {
		// Generate API request body from plan
		const deleteoplog = `mutation DeleteOplog ($id: bigint){
			delete_oplog(where: {id: {_eq: $id}}) {
				returning {
					id
				}
			}
		}`
		request := graphql.NewRequest(deleteoplog)
		request.Var("id", state.ID.ValueInt64())
		var respData map[string]interface{}
		if err := r.client.Run(ctx, request, &respData); err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Ghostwriter Oplog",
				"Could not delete oplog ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
			)
			return
		}
	} else {
		tflog.Info(ctx, "Cowardly refusing to delete oplog. Oplog expiration will be managed by ghostwriter. Set force_delete to true to delete oplog.")
		return
	}
}
