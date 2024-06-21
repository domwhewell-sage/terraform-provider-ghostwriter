package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/machinebox/graphql"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &domainResource{}
	_ resource.ResourceWithConfigure = &domainResource{}
)

// NewdomainResource is a helper function to simplify the provider implementation.
func NewdomainResource() resource.Resource {
	return &domainResource{}
}

// domainResource is the resource implementation.
type domainResource struct {
	client *graphql.Client
}

// orderResourceModel maps the resource schema data.
type domainResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Registrar         types.String `tfsdk:"registrar"`
	Creation          types.String `tfsdk:"creation"`
	Expiration        types.String `tfsdk:"expiration"`
	AutoRenew         types.Bool   `tfsdk:"auto_renew"`
	BurnedExplanation types.String `tfsdk:"burned_explanation"`
	Note              types.String `tfsdk:"note"`
	VtPermalink       types.String `tfsdk:"vt_permalink"`
	LastUpdated       types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Configure adds the provider configured client to the resource.
func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Register a domain in Ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Placeholder identifier attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the domain.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The domain name. e.g. example.com",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}$`),
						"Domain name must be a valid domain name. e.g. example.com",
					),
				},
			},
			"registrar": schema.StringAttribute{
				Description: "The domain registrar. e.g. GoDaddy, Namecheap, etc.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 255),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9-]+$`),
						"Registrar must be a valid domain name. e.g. GoDaddy",
					),
				},
			},
			"creation": schema.StringAttribute{
				Description: "The domain creation date. Format: YYYY-MM-DD.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
						"Date must be in the format YYYY-MM-DD. e.g. 2022-01-01",
					),
				},
			},
			"expiration": schema.StringAttribute{
				Description: "The domain expiration date. Format: YYYY-MM-DD.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
						"Date must be in the format YYYY-MM-DD. e.g. 2022-01-01",
					),
				},
			},
			"auto_renew": schema.BoolAttribute{
				Description: "Whether the domain is set to auto-renew.",
				Optional:    true,
			},
			"burned_explanation": schema.StringAttribute{
				Description: "Explanation of why the domain was burned.",
				Optional:    true,
				Default:     nil,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"note": schema.StringAttribute{
				Description: "Additional notes about the domain.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
			"vt_permalink": schema.StringAttribute{
				Description: "The VirusTotal permalink for the domain.",
				Optional:    true,
				Default:     nil,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const insertdomain = `mutation InsertDomain ($burned_explanation: String, $autoRenew: Boolean, $name: String, $registrar: String, $creation: date, $expiration: date, $note: String, $vtPermalink: String) {
		insert_domain(objects: {burned_explanation: $burned_explanation, autoRenew: $autoRenew, name: $name, registrar: $registrar, creation: $creation, expiration: $expiration, note: $note, vtPermalink: $vtPermalink}) {
			returning {
				id
			}
		}
	}`
	request := graphql.NewRequest(insertdomain)
	request.Var("burned_explanation", plan.BurnedExplanation.ValueString())
	request.Var("autoRenew", plan.AutoRenew.ValueBool())
	request.Var("name", plan.Name.ValueString())
	request.Var("registrar", plan.Registrar.ValueString())
	request.Var("creation", plan.Creation.ValueString())
	request.Var("expiration", plan.Expiration.ValueString())
	request.Var("note", plan.Note.ValueString())
	request.Var("vtPermalink", plan.VtPermalink.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain",
			"Could not create domain, unexpected error: "+err.Error(),
		)
		return
	}

	domainID := respData["insert_domain"].(map[string]interface{})["returning"].([]interface{})[0].(map[string]interface{})
	plan.ID = types.Int64Value(int64(domainID["id"].(float64)))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const querydomain = `query QueryDomain ($id: bigint){
		domain(where: {id: {_eq: $id}}) {
			burned_explanation,
			autoRenew,
			name,
			registrar,
			creation,
			expiration,
			note,
			vtPermalink
		}
	}`
	request := graphql.NewRequest(querydomain)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Ghostwriter Domain",
			"Could not read Ghostwriter domain ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	domain := respData["domain"].([]interface{})[0].(map[string]interface{})
	state.AutoRenew = types.BoolValue(domain["autoRenew"].(bool))
	state.BurnedExplanation = types.StringValue(domain["burned_explanation"].(string))
	state.Creation = types.StringValue(domain["creation"].(string))
	state.Expiration = types.StringValue(domain["expiration"].(string))
	state.Name = types.StringValue(domain["name"].(string))
	state.Note = types.StringValue(domain["note"].(string))
	state.Registrar = types.StringValue(domain["registrar"].(string))
	state.VtPermalink = types.StringValue(domain["vtPermalink"].(string))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan domainResourceModel
	var state domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const updatedomain = `mutation UpdateDomain ($id: bigint, $burned_explanation: String, $autoRenew: Boolean, $name: String, $registrar: String, $creation: date, $expiration: date, $note: String, $vtPermalink: String) {
		update_domain(where: {id: {_eq: $id}}, _set: {burned_explanation: $burned_explanation, autoRenew: $autoRenew, name: $name, registrar: $registrar, creation: $creation, expiration: $expiration, note: $note, vtPermalink: $vtPermalink}) {
			returning {
				id
			}
		}
	}`
	request := graphql.NewRequest(updatedomain)
	request.Var("id", state.ID.ValueInt64())
	request.Var("burned_explanation", plan.BurnedExplanation.ValueString())
	request.Var("autoRenew", plan.AutoRenew.ValueBool())
	request.Var("name", plan.Name.ValueString())
	request.Var("registrar", plan.Registrar.ValueString())
	request.Var("creation", plan.Creation.ValueString())
	request.Var("expiration", plan.Expiration.ValueString())
	request.Var("note", plan.Note.ValueString())
	request.Var("vtPermalink", plan.VtPermalink.ValueString())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Ghostwriter Domain",
			"Could not update domain ID "+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	domainID := respData["update_domain"].(map[string]interface{})["returning"].([]interface{})[0].(map[string]interface{})
	plan.ID = types.Int64Value(int64(domainID["id"].(float64)))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	const deletedomain = `mutation DeleteDomain ($id: bigint){
		delete_domain(where: {id: {_eq: $id}}) {
			returning {
				id
			}
		}
	}`
	request := graphql.NewRequest(deletedomain)
	request.Var("id", state.ID.ValueInt64())
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Ghostwriter Domain",
			"Could not delete domain ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}
}
