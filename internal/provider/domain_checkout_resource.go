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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &domainCheckoutResource{}
	_ resource.ResourceWithConfigure   = &domainCheckoutResource{}
	_ resource.ResourceWithImportState = &domainCheckoutResource{}
)

// NewdomainCheckoutResource is a helper function to simplify the provider implementation.
func NewdomainCheckoutResource() resource.Resource {
	return &domainCheckoutResource{}
}

// domainCheckoutResource is the resource implementation.
type domainCheckoutResource struct {
	client *graphql.Client
}

// orderResourceModel maps the resource schema data.
type domainCheckoutResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ActivityTypeId types.Int64  `tfsdk:"activity_type_id"`
	DomainId       types.Int64  `tfsdk:"domain_id"`
	ProjectId      types.Int64  `tfsdk:"project_id"`
	Note           types.String `tfsdk:"note"`
	StartDate      types.String `tfsdk:"start_date"`
	EndDate        types.String `tfsdk:"end_date"`
	ForceDelete    types.Bool   `tfsdk:"force_delete"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *domainCheckoutResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_checkout"
}

// Configure adds the provider configured client to the resource.
func (r *domainCheckoutResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *domainCheckoutResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Checkout an existing domain in ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the domain checkout.",
				Computed:    true,
			},
			"project_id": schema.Int64Attribute{
				Description: "The unique identifier of the project the domain should be checked out to.",
				Required:    true,
			},
			"domain_id": schema.Int64Attribute{
				Description: "The unique identifier of the domain to be checked out.",
				Required:    true,
			},
			"start_date": schema.StringAttribute{
				Description: "The start date of the project. Format: YYYY-MM-DD.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
						"Date must be in the format YYYY-MM-DD. e.g. 2022-01-01",
					),
				},
			},
			"end_date": schema.StringAttribute{
				Description: "The end date of the project. Format: YYYY-MM-DD.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
						"Date must be in the format YYYY-MM-DD. e.g. 2022-01-01",
					),
				},
			},
			"activity_type_id": schema.Int64Attribute{
				Description: "The unique identifier of the activity type being performed.",
				Required:    true,
			},
			"note": schema.StringAttribute{
				Description: "Project-related notes, such as how the domain will be used/how it worked out.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"force_delete": schema.BoolAttribute{
				Description: "If false, the domain checkout not be deleted but the domain will be released and the record will remain. If true, the domain checkout record will be hard-deleted from the ghostwriter instance. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// ImportState imports the resource state from Terraform state.
func (r *domainCheckoutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	tflog.Debug(ctx, fmt.Sprintf("Importing domain resource ID: %s", req.ID))
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
func (r *domainCheckoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainCheckoutResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const checkoutdomain = `mutation checkoutDomain ($activity_type_id: Int!, $domain_id: Int!, $project_id: Int!, $note: String, $start_date: date!, $end_date: date!) {
		checkoutDomain(activityTypeId: $activity_type_id, domainId: $domain_id, projectId: $project_id, note: $note, startDate: $start_date, endDate: $end_date) {
			result
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Checking out domain: %v", plan))
	request := graphql.NewRequest(checkoutdomain)
	request.Var("activity_type_id", plan.ActivityTypeId.ValueInt64())
	request.Var("domain_id", plan.DomainId.ValueInt64())
	request.Var("project_id", plan.ProjectId.ValueInt64())
	request.Var("note", plan.Note.ValueString())
	request.Var("start_date", plan.StartDate.ValueString())
	request.Var("end_date", plan.EndDate.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error checking out domain",
			"Could not checkout domain, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))

	// Run another query to get the ID of the domain checkout
	const querydomaincheckout = `query QueryDomainCheckout ($id: bigint){
		domainCheckout(where: {domain: {id: {_eq: $id}}}, order_by: {id: desc}) {
			id
			domainId
			endDate
			note
			projectId
			startDate
			activityType {
			  id
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Query domain checkout ID: %v", plan.DomainId.ValueInt64()))
	getid_request := graphql.NewRequest(querydomaincheckout)
	getid_request.Var("id", plan.DomainId.ValueInt64())
	var getidResp map[string]interface{}
	if err := r.client.Run(ctx, getid_request, &getidResp); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Domain Checkouts",
			"Could not read Ghostwriter domain checkout ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Domain checkout response: %v", respData))

	domain_checkouts := getidResp["domainCheckout"].([]interface{})
	if len(domain_checkouts) >= 1 {
		latest_checkout := domain_checkouts[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(latest_checkout["id"].(float64)))
		plan.ActivityTypeId = types.Int64Value(int64(latest_checkout["activityType"].(map[string]interface{})["id"].(float64)))
		plan.DomainId = types.Int64Value(int64(latest_checkout["domainId"].(float64)))
		plan.ProjectId = types.Int64Value(int64(latest_checkout["projectId"].(float64)))
		plan.Note = types.StringValue(latest_checkout["note"].(string))
		plan.StartDate = types.StringValue(latest_checkout["startDate"].(string))
		plan.EndDate = types.StringValue(latest_checkout["endDate"].(string))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Domain Checkouts",
			"Could not obtain any checkouts for the domain ID "+strconv.FormatInt(plan.DomainId.ValueInt64(), 10),
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *domainCheckoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state domainCheckoutResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const querydomaincheckout = `query QueryDomainCheckout ($id: bigint){
		domainCheckout(where: {id: {_eq: $id}}) {
			id
			domainId
			endDate
			note
			projectId
			startDate
			activityType {
			  id
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Query domain checkout ID: %v", state.ID.ValueInt64()))
	request := graphql.NewRequest(querydomaincheckout)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Domain Checkouts",
			"Could not read Ghostwriter domain checkout ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Domain checkout response: %v", respData))

	// Overwrite items with refreshed state
	domain_checkouts := respData["domainCheckout"].([]interface{})
	if len(domain_checkouts) >= 1 {
		latest_checkout := domain_checkouts[0].(map[string]interface{})
		state.ActivityTypeId = types.Int64Value(int64(latest_checkout["activityType"].(map[string]interface{})["id"].(float64)))
		state.DomainId = types.Int64Value(int64(latest_checkout["domainId"].(float64)))
		state.ProjectId = types.Int64Value(int64(latest_checkout["projectId"].(float64)))
		state.Note = types.StringValue(latest_checkout["note"].(string))
		state.StartDate = types.StringValue(latest_checkout["startDate"].(string))
		state.EndDate = types.StringValue(latest_checkout["endDate"].(string))

		// Set refreshed state
		diags = resp.State.Set(ctx, &state)
	} else {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Domain Checkouts",
			"Could not obtain any checkouts for the domain ID "+strconv.FormatInt(state.DomainId.ValueInt64(), 10),
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *domainCheckoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan domainCheckoutResourceModel
	var state domainCheckoutResourceModel
	diags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const updatedomaincheckout = `mutation UpdateDomainCheckout ($id: bigint, $activity_type_id: bigint, $domain_id: bigint, $project_id: bigint, $note: String, $start_date: date!, $end_date: date!) {
		update_domainCheckout(where: {id: {_eq: $id}}, _set: {activityTypeId: $activity_type_id, domainId: $domain_id, endDate: $end_date, note: $note, projectId: $project_id, startDate: $start_date}) {
			returning {
				id
				domainId
				endDate
				note
				projectId
				startDate
				activityType {
				  id
				}
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Updating domain checkout: %v", plan))
	request := graphql.NewRequest(updatedomaincheckout)
	request.Var("id", state.ID.ValueInt64())
	request.Var("activity_type_id", plan.ActivityTypeId.ValueInt64())
	request.Var("domain_id", plan.DomainId.ValueInt64())
	request.Var("project_id", plan.ProjectId.ValueInt64())
	request.Var("note", plan.Note.ValueString())
	request.Var("start_date", plan.StartDate.ValueString())
	request.Var("end_date", plan.EndDate.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Domain Checkout",
			"Could not update domain checkout ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response from Ghostwriter: %v", respData))

	updated_domain_checkouts := respData["update_domainCheckout"].(map[string]interface{})["returning"].([]interface{})
	if len(updated_domain_checkouts) == 1 {
		updated_domain_checkout := updated_domain_checkouts[0].(map[string]interface{})
		plan.ID = types.Int64Value(int64(updated_domain_checkout["id"].(float64)))
		plan.ActivityTypeId = types.Int64Value(int64(updated_domain_checkout["activityType"].(map[string]interface{})["id"].(float64)))
		plan.DomainId = types.Int64Value(int64(updated_domain_checkout["domainId"].(float64)))
		plan.ProjectId = types.Int64Value(int64(updated_domain_checkout["projectId"].(float64)))
		plan.Note = types.StringValue(updated_domain_checkout["note"].(string))
		plan.StartDate = types.StringValue(updated_domain_checkout["startDate"].(string))
		plan.EndDate = types.StringValue(updated_domain_checkout["endDate"].(string))
		plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

		// Set state to fully populated data
		diags = resp.State.Set(ctx, plan)
	} else {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Domain Checkout",
			"Could not update domain checkout ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10),
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainCheckoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state domainCheckoutResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ForceDelete.ValueBool() {
		// Generate API request body from plan
		const deletedomaincheckout = `mutation DeleteDomainCheckout ($id: bigint) {
			delete_domainCheckout(where: {id: {_eq: $id}}) {
				returning {
					id
				}
			}
		}`
		tflog.Debug(ctx, fmt.Sprintf("Deleting domain checkout: %v", state))
		request := graphql.NewRequest(deletedomaincheckout)
		request.Var("id", state.ID.ValueInt64())
		var respData map[string]interface{}
		if err := r.client.Run(ctx, request, &respData); err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Ghostwriter Domain Checkout",
				"Could not delete domain checkout ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
			)
		}
	} else {
		tflog.Info(ctx, "Cowardly refusing to delete domain checkout. Releasing domain to the ghostwriter pool and the domain checkout record will remain. Set force_delete to true to delete domain checkout record.")
	}
	// Generate API request body from plan
	const releasedomain = `mutation UpdateDomain ($id: bigint) {
		update_domain(where: {id: {_eq: $id}}, _set: {domainStatusId: 1}) {
			returning {
				id
			}
		}
	}`
	tflog.Debug(ctx, fmt.Sprintf("Releasing domain to the pool: %v", state))
	request := graphql.NewRequest(releasedomain)
	request.Var("id", state.DomainId.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Ghostwriter Domain Checkout",
			"Could not delete domain checkout ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}
}
