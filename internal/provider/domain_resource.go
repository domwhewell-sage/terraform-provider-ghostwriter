package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"registrar": schema.StringAttribute{
				Computed: true,
			},
			"creation": schema.StringAttribute{
				Computed: true,
			},
			"expiration": schema.StringAttribute{
				Computed: true,
			},
			"auto_renew": schema.BoolAttribute{
				Computed: true,
			},
			"burned_explanation": schema.StringAttribute{
				Computed: true,
			},
			"note": schema.StringAttribute{
				Computed: true,
			},
			"vt_permalink": schema.StringAttribute{
				Computed: true,
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
	const insertdomain = `mutation InsertDomain ($burned_explanation: String, $autoRenew: Boolean, $name: String, $registrar: String, $creation: String, $expiration: String, $note: String, $vtPermalink: String) {
		insert_domain(objects: {burned_explanation: $burned_explanation, autoRenew: $autoRenew, name: $name, registrar: $registrar, creation: $creation, expiration: $expiration, note: $note, vtPermalink: $vtPermalink}) {
			returning {
				id
			}
		}
	}`
	request := graphql.NewRequest(insertdomain)
	request.Var("burned_explanation", plan.BurnedExplanation)
	request.Var("autoRenew", plan.AutoRenew)
	request.Var("name", plan.Name)
	request.Var("registrar", plan.Registrar)
	request.Var("creation", plan.Creation)
	request.Var("expiration", plan.Expiration)
	request.Var("note", plan.Note)
	request.Var("vtPermalink", plan.VtPermalink)
	var respData map[string]interface{}
	if err := r.client.Run(ctx, request, &respData); err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain",
			"Could not create domain, unexpected error: "+err.Error(),
		)
		return
	}

	domainID := respData["data"].(map[string]interface{})["insert_domain"].(map[string]interface{})["returning"].([]interface{})[0].(map[string]interface{})
	plan.ID = types.Int64Value(domainID["id"].(int64))
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
			"Error creating domain",
			"Could not create domain, unexpected error: "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	domainData := respData["data"].(map[string]interface{})["domain"].([]interface{})
	if len(domainData) > 0 {
		domain := domainData[0].(map[string]interface{})
		state.AutoRenew = types.BoolValue(domain["autoRenew"].(bool))
		state.BurnedExplanation = types.StringValue(domain["burned_explanation"].(string))
		state.Creation = types.StringValue(domain["creation"].(string))
		state.Expiration = types.StringValue(domain["expiration"].(string))
		state.Name = types.StringValue(domain["name"].(string))
		state.Note = types.StringValue(domain["note"].(string))
		state.Registrar = types.StringValue(domain["registrar"].(string))
		state.VtPermalink = types.StringValue(domain["vtPermalink"].(string))
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
