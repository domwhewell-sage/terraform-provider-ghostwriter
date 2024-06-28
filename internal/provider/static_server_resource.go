package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &staticserverResource{}
	_ resource.ResourceWithConfigure   = &staticserverResource{}
	_ resource.ResourceWithImportState = &staticserverResource{}
)

// NewstaticserverResource is a helper function to simplify the provider implementation.
func NewstaticserverResource() resource.Resource {
	return &staticserverResource{}
}

// staticserverResource is the resource implementation.
type staticserverResource struct {
	client *graphql.Client
}

// orderResourceModel maps the resource schema data.
type staticserverResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ServerProviderID types.Int64  `tfsdk:"server_provider_id"`
	ServerStatusId   types.Int64  `tfsdk:"server_status_id"`
	IpAddress        types.String `tfsdk:"ip_address"`
	Note             types.String `tfsdk:"note"`
	LastUpdated      types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *staticserverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_server"
}

// Configure adds the provider configured client to the resource.
func (r *staticserverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *staticserverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Register a static server in Ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the server.",
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
			"server_status_id": schema.Int64Attribute{
				Description: "The identifier of the server status.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
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
			"note": schema.StringAttribute{
				Description: "Additional notes about the server.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
		},
	}
}

// ImportState imports the resource state from Terraform state.
func (r *staticserverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	tflog.Debug(ctx, fmt.Sprintf("Importing static server resource ID: %s", req.ID))
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Import ID",
			"Could not parse import ID: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *staticserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan staticserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const insertserver = `mutation InsertServer($name: String, $server_provider_id: bigint, $server_status_id: bigint, $ip: inet, $note: String) {
		insert_staticServer(objects: {name: $name, serverProviderId: $server_provider_id, serverStatusId: $server_status_id, ipAddress: $ip, note: $note}) {
			returning {
				id,
				name,
				serverProviderId,
				serverStatusId,
				ipAddress,
				note
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Creating server: %v", plan))
	request := graphql.NewRequest(insertserver)
	request.Var("name", plan.Name.ValueString())
	request.Var("server_provider_id", plan.ServerProviderID.ValueInt64())
	request.Var("server_status_id", plan.ServerStatusId.ValueInt64())
	request.Var("ip", plan.IpAddress.ValueString())
	request.Var("note", plan.Note.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error creating server",
			"Could not create server, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	servers := respData["insert_staticServer"].(map[string]interface{})["returning"].([]interface{})
	if len(servers) == 1 {
		server := servers[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(server["id"].(float64)))
		plan.Name = types.StringValue(server["name"].(string))
		plan.ServerProviderID = types.Int64Value(int64(server["serverProviderId"].(float64)))
		plan.ServerStatusId = types.Int64Value(int64(server["serverStatusId"].(float64)))
		plan.IpAddress = types.StringValue(server["ipAddress"].(string))
		plan.Note = types.StringValue(server["note"].(string))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error creating server",
			"Could not create server: Server not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *staticserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state staticserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const queryserver = `query StaticServer ($id: bigint){
		staticServer(where: {id: {_eq: $id}}) {
			id,
			name,
			serverProviderId,
			serverStatusId,
			ipAddress,
			note
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Reading server: %v", state.ID))
	request := graphql.NewRequest(queryserver)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Server",
			"Could not read Ghostwriter server ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	servers := respData["staticServer"].([]interface{})
	if len(servers) == 1 {
		server := servers[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(server["id"].(float64)))
		state.Name = types.StringValue(server["name"].(string))
		state.ServerProviderID = types.Int64Value(int64(server["serverProviderId"].(float64)))
		state.ServerStatusId = types.Int64Value(int64(server["serverStatusId"].(float64)))
		state.IpAddress = types.StringValue(server["ipAddress"].(string))
		state.Note = types.StringValue(server["note"].(string))
		state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error creating server",
			"Could not create server: Server not found",
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *staticserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan staticserverResourceModel
	var state staticserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const updateserver = `mutation UpdateServer($id: bigint, $name: String, $server_provider_id: bigint, $server_status_id: bigint, $ip: inet, $note: String) {
		update_staticServer(where: {id: {_eq: $id}}, _set: {name: $name, serverProviderId: $server_provider_id, serverStatusId: $server_status_id, ipAddress: $ip, note: $note}) {
			returning {
				id,
				name,
				serverProviderId,
				serverStatusId,
				ipAddress,
				note
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Updating server: %v", plan))
	request := graphql.NewRequest(updateserver)
	request.Var("id", state.ID.ValueInt64())
	request.Var("name", plan.Name.ValueString())
	request.Var("server_provider_id", plan.ServerProviderID.ValueInt64())
	request.Var("server_status_id", plan.ServerStatusId.ValueInt64())
	request.Var("ip", plan.IpAddress.ValueString())
	request.Var("note", plan.Note.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Server",
			"Could not update server ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	updated_servers := respData["update_staticServer"].(map[string]interface{})["returning"].([]interface{})
	if len(updated_servers) == 1 {
		server := updated_servers[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(server["id"].(float64)))
		plan.Name = types.StringValue(server["name"].(string))
		plan.ServerProviderID = types.Int64Value(int64(server["serverProviderId"].(float64)))
		plan.ServerStatusId = types.Int64Value(int64(server["serverStatusId"].(float64)))
		plan.IpAddress = types.StringValue(server["ipAddress"].(string))
		plan.Note = types.StringValue(server["note"].(string))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Server",
			"Could not update server ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": Server not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *staticserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state staticserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const deleteserver = `mutation DeleteServer ($id: bigint){
		delete_staticServer(where: {id: {_eq: $id}}) {
			returning {
				id
			}
		}
	}`
	request := graphql.NewRequest(deleteserver)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Ghostwriter Server",
			"Could not delete server ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}
}
