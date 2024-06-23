package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &cloudserverResource{}
	_ resource.ResourceWithConfigure   = &cloudserverResource{}
	_ resource.ResourceWithImportState = &cloudserverResource{}
)

// NewcloudserverResource is a helper function to simplify the provider implementation.
func NewcloudserverResource() resource.Resource {
	return &cloudserverResource{}
}

// cloudserverResource is the resource implementation.
type cloudserverResource struct {
	client *graphql.Client
}

// orderResourceModel maps the resource schema data.
type cloudserverResourceModel struct {
	ID               types.Int64    `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	ServerProviderID types.Int64    `tfsdk:"server_provider_id"`
	ActivityTypeId   types.Int64    `tfsdk:"activity_type_id"`
	IpAddress        types.String   `tfsdk:"ip_address"`
	AuxAddress       []types.String `tfsdk:"aux_address"`
	ProjectID        types.Int64    `tfsdk:"project_id"`
	Note             types.String   `tfsdk:"note"`
	OperatorID       types.Int64    `tfsdk:"operator_id"`
	ServerRoleId     types.Int64    `tfsdk:"server_role_id"`
	ForceDelete      types.Bool     `tfsdk:"force_delete"`
	LastUpdated      types.String   `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *cloudserverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_server"
}

// Configure adds the provider configured client to the resource.
func (r *cloudserverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *cloudserverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Add a cloud server to ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the cloud server.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the server typically its hostname.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"server_provider_id": schema.Int64Attribute{
				Description: "The identifier of the server hosting provider.",
				Required:    true,
			},
			"activity_type_id": schema.Int64Attribute{
				Description: "How this VPS will be used.",
				Required:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "The servers IP address.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$|^([0-9a-fA-F]{0,4}:){7}[0-9a-fA-F]{0,4}$`),
						"Must be an IPv4 or IPv6 address.",
					),
				},
			},
			"aux_address": schema.ListAttribute{
				Description: "Any additional IP addresses associated with the server.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"project_id": schema.Int64Attribute{
				Description: "The project this server is associated with.",
				Required:    true,
			},
			"note": schema.StringAttribute{
				Description: "Additional notes about the cloud server.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"operator_id": schema.Int64Attribute{
				Description: "The operator responsible for this server.",
				Required:    true,
			},
			"server_role_id": schema.Int64Attribute{
				Description: "The role of the server.",
				Required:    true,
			},
			"force_delete": schema.BoolAttribute{
				Description: "If false, the server will be soft-deleted left to expire by the ghostwriter instance. If true, the server will be hard-deleted from the ghostwriter instance. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// ImportState imports the resource state from Terraform state.
func (r *cloudserverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	tflog.Debug(ctx, fmt.Sprintf("Importing cloud server resource ID: %s", req.ID))
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
func (r *cloudserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const insertcloudserver = `mutation InsertCloudServer ($name: String, $server_provider_id: bigint, $activity_type_id: bigint, $ip: inet, $aux_address: [inet!], $project_id: bigint, $note: String, $operator_id: bigint, $server_role_id: bigint) {
		insert_cloudServer(objects: {name: $name, serverProviderId: $server_provider_id, activityTypeId: $activity_type_id, ipAddress: $ip, auxAddress: $aux_address, projectId: $project_id, note: $note, operatorId: $operator_id, serverRoleId: $server_role_id}) {
			returning {
				id,
				name,
				serverProviderId,
				activityTypeId,
				ipAddress,
				auxAddress,
				projectId,
				note,
				operatorId,
				serverRoleId
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Creating cloudserver: %v", plan))
	request := graphql.NewRequest(insertcloudserver)
	request.Var("name", plan.Name.ValueString())
	request.Var("server_provider_id", plan.ServerProviderID.ValueInt64())
	request.Var("activity_type_id", plan.ActivityTypeId.ValueInt64())
	request.Var("ip", plan.IpAddress.ValueString())
	aux_address := []interface{}{}
	for _, address := range plan.AuxAddress {
		aux_address = append(aux_address, address.ValueString())
	}
	request.Var("aux_address", aux_address)
	request.Var("project_id", plan.ProjectID.ValueInt64())
	request.Var("note", plan.Note.ValueString())
	request.Var("operator_id", plan.OperatorID.ValueInt64())
	request.Var("server_role_id", plan.ServerRoleId.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error creating cloud server",
			"Could not create cloud server, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	cloud_servers := respData["insert_cloudServer"].(map[string]interface{})["returning"].([]interface{})
	if len(cloud_servers) == 1 {
		cloud_server := cloud_servers[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(cloud_server["id"].(float64)))
		plan.Name = types.StringValue(cloud_server["name"].(string))
		plan.ServerProviderID = types.Int64Value(int64(cloud_server["serverProviderId"].(float64)))
		plan.ActivityTypeId = types.Int64Value(int64(cloud_server["activityTypeId"].(float64)))
		plan.IpAddress = types.StringValue(cloud_server["ipAddress"].(string))
		aux_address, ok := cloud_server["auxAddress"].([]interface{})
		aux_address_list := []types.String{}
		if ok {
			for _, address := range aux_address {
				aux_address_list = append(aux_address_list, types.StringValue(address.(string)))
			}
		}
		plan.AuxAddress = aux_address_list
		plan.ProjectID = types.Int64Value(int64(cloud_server["projectId"].(float64)))
		plan.Note = types.StringValue(cloud_server["note"].(string))
		plan.OperatorID = types.Int64Value(int64(cloud_server["operatorId"].(float64)))
		plan.ServerRoleId = types.Int64Value(int64(cloud_server["serverRoleId"].(float64)))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error creating cloud server",
			"Could not create cloud server: Cloud server not created",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *cloudserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state cloudserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const querycloudserver = `query CloudServer ($id: bigint){
		cloudServer(where: {id: {_eq: $id}}) {
			id,
			name,
			serverProviderId,
			activityTypeId,
			ipAddress,
			auxAddress,
			projectId,
			note,
			operatorId,
			serverRoleId
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Reading cloud server: %v", state.ID))
	request := graphql.NewRequest(querycloudserver)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Cloud Server",
			"Could not read Ghostwriter cloud server ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	cloud_servers := respData["cloudServer"].([]interface{})
	if len(cloud_servers) == 1 {
		cloud_server := cloud_servers[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(cloud_server["id"].(float64)))
		state.Name = types.StringValue(cloud_server["name"].(string))
		state.ServerProviderID = types.Int64Value(int64(cloud_server["serverProviderId"].(float64)))
		state.ActivityTypeId = types.Int64Value(int64(cloud_server["activityTypeId"].(float64)))
		state.IpAddress = types.StringValue(cloud_server["ipAddress"].(string))
		aux_address, ok := cloud_server["auxAddress"].([]interface{})
		aux_address_list := []types.String{}
		if ok {
			for _, address := range aux_address {
				aux_address_list = append(aux_address_list, types.StringValue(address.(string)))
			}
		}
		state.AuxAddress = aux_address_list
		state.ProjectID = types.Int64Value(int64(cloud_server["projectId"].(float64)))
		state.Note = types.StringValue(cloud_server["note"].(string))
		state.OperatorID = types.Int64Value(int64(cloud_server["operatorId"].(float64)))
		state.ServerRoleId = types.Int64Value(int64(cloud_server["serverRoleId"].(float64)))

		// Set refreshed state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter cloud server",
			"Could not read Ghostwriter cloud server ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": Cloud Server not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *cloudserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan cloudserverResourceModel
	var state cloudserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const updatecloudserver = `mutation UpdateCloudServer ($id: bigint, $name: String, $server_provider_id: bigint, $activity_type_id: bigint, $ip: inet, $aux_address: [inet!], $project_id: bigint, $note: String, $operator_id: bigint, $server_role_id: bigint) {
		update_cloudServer(where: {id: {_eq: $id}}, _set: {name: $name, serverProviderId: $server_provider_id, activityTypeId: $activity_type_id, ipAddress: $ip, auxAddress: $aux_address, projectId: $project_id, note: $note, operatorId: $operator_id, serverRoleId: $server_role_id}) {
			returning {
				id,
				name,
				serverProviderId,
				activityTypeId,
				ipAddress,
				auxAddress,
				projectId,
				note,
				operatorId,
				serverRoleId
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Updating cloud server: %v", plan))
	request := graphql.NewRequest(updatecloudserver)
	request.Var("id", state.ID.ValueInt64())
	request.Var("name", plan.Name.ValueString())
	request.Var("server_provider_id", plan.ServerProviderID.ValueInt64())
	request.Var("activity_type_id", plan.ActivityTypeId.ValueInt64())
	request.Var("ip", plan.IpAddress.ValueString())
	aux_address := []interface{}{}
	for _, address := range plan.AuxAddress {
		aux_address = append(aux_address, address.ValueString())
	}
	request.Var("aux_address", aux_address)
	request.Var("project_id", plan.ProjectID.ValueInt64())
	request.Var("note", plan.Note.ValueString())
	request.Var("operator_id", plan.OperatorID.ValueInt64())
	request.Var("server_role_id", plan.ServerRoleId.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Cloud Server",
			"Could not update cloud server ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	cloud_servers := respData["update_cloudServer"].(map[string]interface{})["returning"].([]interface{})
	if len(cloud_servers) == 1 {
		cloud_server := cloud_servers[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(cloud_server["id"].(float64)))
		plan.Name = types.StringValue(cloud_server["name"].(string))
		plan.ServerProviderID = types.Int64Value(int64(cloud_server["serverProviderId"].(float64)))
		plan.ActivityTypeId = types.Int64Value(int64(cloud_server["activityTypeId"].(float64)))
		plan.IpAddress = types.StringValue(cloud_server["ipAddress"].(string))
		aux_address, ok := cloud_server["auxAddress"].([]interface{})
		aux_address_list := []types.String{}
		if ok {
			for _, address := range aux_address {
				aux_address_list = append(aux_address_list, types.StringValue(address.(string)))
			}
		}
		plan.AuxAddress = aux_address_list
		plan.ProjectID = types.Int64Value(int64(cloud_server["projectId"].(float64)))
		plan.Note = types.StringValue(cloud_server["note"].(string))
		plan.OperatorID = types.Int64Value(int64(cloud_server["operatorId"].(float64)))
		plan.ServerRoleId = types.Int64Value(int64(cloud_server["serverRoleId"].(float64)))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Cloud Server",
			"Could not update cloud server ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": Cloud Server not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *cloudserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state cloudserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ForceDelete.ValueBool() {
		// Generate API request body from plan
		const deletecloudserver = `mutation DeleteCloudServer ($id: bigint){
			delete_cloudServer(where: {id: {_eq: $id}}) {
				returning {
					id
				}
			}
		}`
		request := graphql.NewRequest(deletecloudserver)
		request.Var("id", state.ID.ValueInt64())
		var respData map[string]interface{}
		if err := r.client.Run(ctx, request, &respData); err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Ghostwriter Cloud Server",
				"Could not delete cloud server ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
			)
			return
		}
	} else {
		tflog.Info(ctx, "Cowardly refusing to delete cloud server. Cloud Server expiration will be managed by ghostwriter. Set force_delete to true to delete cloud server.")
		return
	}
}
