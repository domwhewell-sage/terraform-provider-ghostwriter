package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &domainserverResource{}
	_ resource.ResourceWithConfigure   = &domainserverResource{}
	_ resource.ResourceWithImportState = &domainserverResource{}
)

// NewdomainserverResource is a helper function to simplify the provider implementation.
func NewdomainserverResource() resource.Resource {
	return &domainserverResource{}
}

// domainserverResource is the resource implementation.
type domainserverResource struct {
	client *graphql.Client
}

// orderResourceModel maps the resource schema data.
type domainserverResourceModel struct {
	ID                     types.Int64  `tfsdk:"id"`
	DomainCheckoutID       types.Int64  `tfsdk:"domain_checkout_id"`
	ProjectID              types.Int64  `tfsdk:"project_id"`
	StaticServerCheckoutID types.Int64  `tfsdk:"static_server_checkout_id"`
	TransientServerID      types.Int64  `tfsdk:"cloud_server_id"`
	Subdomain              types.String `tfsdk:"subdomain"`
	Endpoint               types.String `tfsdk:"endpoint"`
	ForceDelete            types.Bool   `tfsdk:"force_delete"`
	LastUpdated            types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *domainserverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_server"
}

// Configure adds the provider configured client to the resource.
func (r *domainserverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *domainserverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associate a Domain + Server in Ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the association.",
				Computed:    true,
			},
			"domain_checkout_id": schema.Int64Attribute{
				Description: "The identifier of the domain checkout resource.",
				Required:    true,
			},
			"project_id": schema.Int64Attribute{
				Description: "The unique identifier of the project the domain + server association should be created for.",
				Required:    true,
			},
			"static_server_checkout_id": schema.Int64Attribute{
				Description: "The identifier of the static server resource.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("cloud_server_id")),
				},
			},
			"cloud_server_id": schema.Int64Attribute{
				Description: "The identifier of the cloud server resource.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("static_server_checkout_id")),
				},
			},
			"subdomain": schema.StringAttribute{
				Description: "The subdomain of the domain. Default is '*' for wildcard.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("*"),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"endpoint": schema.StringAttribute{
				Description: "The endpoint of the domain.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"force_delete": schema.BoolAttribute{
				Description: "If false, will not be deleted from the ghostwriter instance when not managed by terraform. If true, the domain will be hard-deleted from the ghostwriter instance. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// ImportState imports the resource state from Terraform state.
func (r *domainserverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	tflog.Debug(ctx, fmt.Sprintf("Importing domain server resource ID: %s", req.ID))
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
func (r *domainserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var insertdomainserver string
	if plan.StaticServerCheckoutID.ValueInt64() != 0 {
		insertdomainserver = `mutation InsertDomainServerConnection ($domain_checkout_id: bigint, $project_id: bigint, $static_server_checkout_id: bigint, $subdomain: String, $endpoint: String){
			insert_domainServerConnection(objects: {domainId: $domain_checkout_id, endpoint: $endpoint, projectId: $project_id, staticServerId: $static_server_checkout_id, subdomain: $subdomain}) {
				returning {
					domainId
					endpoint
					id
					projectId
					staticServerId
					subdomain
					transientServerId
				}
			}
		}`
	} else if plan.TransientServerID.ValueInt64() != 0 {
		insertdomainserver = `mutation InsertDomainServerConnection ($domain_checkout_id: bigint, $project_id: bigint, $cloud_server_id: bigint, $subdomain: String, $endpoint: String){
			insert_domainServerConnection(objects: {domainId: $domain_checkout_id, endpoint: $endpoint, projectId: $project_id, transientServerId: $cloud_server_id, subdomain: $subdomain}) {
				returning {
					domainId
					endpoint
					id
					projectId
					staticServerId
					subdomain
					transientServerId
				}
			}
		}`
	} else {
		tflog.Error(ctx, "Error creating domain server association can only specify either static_server_checkout_id or cloud_server_id")
	}
	tflog.Debug(ctx, fmt.Sprintf("Creating domain server association: %v", plan))
	request := graphql.NewRequest(insertdomainserver)
	request.Var("domain_checkout_id", plan.DomainCheckoutID.ValueInt64())
	request.Var("project_id", plan.ProjectID.ValueInt64())
	if plan.StaticServerCheckoutID.ValueInt64() != 0 {
		request.Var("static_server_checkout_id", plan.StaticServerCheckoutID.ValueInt64())
	} else if plan.TransientServerID.ValueInt64() != 0 {
		request.Var("cloud_server_id", plan.TransientServerID.ValueInt64())
	} else {
		tflog.Error(ctx, "Error creating domain server association can only specify either static_server_checkout_id or cloud_server_id")
	}
	request.Var("subdomain", plan.Subdomain.ValueString())
	request.Var("endpoint", plan.Endpoint.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain server association",
			"Could not create domain server association, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	domainservers := respData["insert_domainServerConnection"].(map[string]interface{})["returning"].([]interface{})
	if len(domainservers) == 1 {
		domainserver := domainservers[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(domainserver["id"].(float64)))
		plan.DomainCheckoutID = types.Int64Value(int64(domainserver["domainId"].(float64)))
		plan.ProjectID = types.Int64Value(int64(domainserver["projectId"].(float64)))
		static_server_id, ok := domainserver["staticServerId"].(float64)
		if !ok || domainserver["staticServerId"] == nil {
			plan.StaticServerCheckoutID = types.Int64Value(0)
		} else {
			plan.StaticServerCheckoutID = types.Int64Value(int64(static_server_id))
		}
		cloud_server_id, ok := domainserver["transientServerId"].(float64)
		if !ok || domainserver["transientServerId"] == nil {
			plan.TransientServerID = types.Int64Value(0)
		} else {
			plan.TransientServerID = types.Int64Value(int64(cloud_server_id))
		}
		plan.Subdomain = types.StringValue(domainserver["subdomain"].(string))
		plan.Endpoint = types.StringValue(domainserver["endpoint"].(string))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error creating domain server association",
			"Could not create domain server association: Domain + Server not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *domainserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state domainserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const querydomainserver = `query QueryDomainServerConnection ($id: bigint) {
		domainServerConnection(where: {id: {_eq: $id}}) {
			domainId
			endpoint
			id
			projectId
			staticServerId
			subdomain
			transientServerId
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Reading domain server association: %v", state.ID))
	request := graphql.NewRequest(querydomainserver)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Domain Server association",
			"Could not read Ghostwriter domain server association ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	domainservers := respData["domainServerConnection"].([]interface{})
	if len(domainservers) == 1 {
		domainserver := domainservers[0].(map[string]interface{})
		state.ID = types.Int64Value(int64(domainserver["id"].(float64)))
		state.DomainCheckoutID = types.Int64Value(int64(domainserver["domainId"].(float64)))
		state.ProjectID = types.Int64Value(int64(domainserver["projectId"].(float64)))
		static_server_id, ok := domainserver["staticServerId"].(float64)
		if !ok || domainserver["staticServerId"] == nil {
			state.StaticServerCheckoutID = types.Int64Value(0)
		} else {
			state.StaticServerCheckoutID = types.Int64Value(int64(static_server_id))
		}
		cloud_server_id, ok := domainserver["transientServerId"].(float64)
		if !ok || domainserver["transientServerId"] == nil {
			state.TransientServerID = types.Int64Value(0)
		} else {
			state.TransientServerID = types.Int64Value(int64(cloud_server_id))
		}
		state.Subdomain = types.StringValue(domainserver["subdomain"].(string))
		state.Endpoint = types.StringValue(domainserver["endpoint"].(string))

		// Set refreshed state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter domain server association",
			"Could not read Ghostwriter domain server association ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": Domain Server not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *domainserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan domainserverResourceModel
	var state domainserverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var updatedomain string
	if plan.StaticServerCheckoutID.ValueInt64() != 0 {
		updatedomain = `mutation UpdateDomainServerConnection ($id: bigint, $domain_checkout_id: bigint, $project_id: bigint, $static_server_checkout_id: bigint, $subdomain: String, $endpoint: String){
			update_domainServerConnection(where: {id: {_eq: $id}}, _set: {domainId: $domain_checkout_id, endpoint: $endpoint, projectId: $project_id, staticServerId: $static_server_checkout_id, subdomain: $subdomain}) {
				returning {
					domainId
					endpoint
					id
					projectId
					staticServerId
					subdomain
					transientServerId
				}
			}
		}`
	} else if plan.TransientServerID.ValueInt64() != 0 {
		updatedomain = `mutation UpdateDomainServerConnection ($id: bigint, $domain_checkout_id: bigint, $project_id: bigint, $cloud_server_id: bigint, $subdomain: String, $endpoint: String){
			update_domainServerConnection(where: {id: {_eq: $id}}, _set: {domainId: $domain_checkout_id, endpoint: $endpoint, projectId: $project_id, subdomain: $subdomain, transientServerId: $cloud_server_id}) {
				returning {
					domainId
					endpoint
					id
					projectId
					staticServerId
					subdomain
					transientServerId
				}
			}
		}`
	} else {
		tflog.Error(ctx, "Error creating domain server association can only specify either static_server_checkout_id or cloud_server_id")
	}
	tflog.Debug(ctx, fmt.Sprintf("Updating domain server association: %v", plan))
	request := graphql.NewRequest(updatedomain)
	request.Var("id", state.ID.ValueInt64())
	request.Var("domain_checkout_id", plan.DomainCheckoutID.ValueInt64())
	request.Var("project_id", plan.ProjectID.ValueInt64())
	if plan.StaticServerCheckoutID.ValueInt64() != 0 {
		request.Var("static_server_checkout_id", plan.StaticServerCheckoutID.ValueInt64())
	} else if plan.TransientServerID.ValueInt64() != 0 {
		request.Var("cloud_server_id", plan.TransientServerID.ValueInt64())
	} else {
		tflog.Error(ctx, "Error creating domain server association can only specify either static_server_checkout_id or cloud_server_id")
	}
	request.Var("subdomain", plan.Subdomain.ValueString())
	request.Var("endpoint", plan.Endpoint.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Domain Server association",
			"Could not update domain server association ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))
	updated_domainservers := respData["update_domainServerConnection"].(map[string]interface{})["returning"].([]interface{})
	if len(updated_domainservers) == 1 {
		domainserver := updated_domainservers[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(domainserver["id"].(float64)))
		plan.DomainCheckoutID = types.Int64Value(int64(domainserver["domainId"].(float64)))
		plan.ProjectID = types.Int64Value(int64(domainserver["projectId"].(float64)))
		static_server_id, ok := domainserver["staticServerId"].(float64)
		if !ok || domainserver["staticServerId"] == nil {
			plan.StaticServerCheckoutID = types.Int64Value(0)
		} else {
			plan.StaticServerCheckoutID = types.Int64Value(int64(static_server_id))
		}
		cloud_server_id, ok := domainserver["transientServerId"].(float64)
		if !ok || domainserver["transientServerId"] == nil {
			plan.TransientServerID = types.Int64Value(0)
		} else {
			plan.TransientServerID = types.Int64Value(int64(cloud_server_id))
		}
		plan.Subdomain = types.StringValue(domainserver["subdomain"].(string))
		plan.Endpoint = types.StringValue(domainserver["endpoint"].(string))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Domain Server association",
			"Could not update domain server association ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": Domain not found",
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state domainserverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ForceDelete.ValueBool() {
		// Generate API request body from plan
		const deletedomainserver = `mutation DeleteDomainServer ($id: bigint){
			delete_domainServerConnection(where: {id: {_eq: $id}}) {
				returning {
					id
				}
			}
		}`
		request := graphql.NewRequest(deletedomainserver)
		request.Var("id", state.ID.ValueInt64())
		var respData map[string]interface{}
		if err := r.client.Run(ctx, request, &respData); err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Ghostwriter Domain Server association",
				"Could not delete domain server association ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
			)
			return
		}
	} else {
		tflog.Info(ctx, "Cowardly refusing to delete domain server association. Association expiration will be managed by ghostwriter. Set force_delete to true to delete domain server connection.")
		return
	}
}
