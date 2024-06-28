package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/machinebox/graphql"
	"golang.org/x/oauth2"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &ghostwriterProvider{}
)

// GhostwriterProviderModel maps provider schema data to a Go type.
type ghostwriterProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	Apikey      types.String `tfsdk:"api_key"`
	TlsInsecure types.Bool   `tfsdk:"tls_insecure"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ghostwriterProvider{
			version: version,
		}
	}
}

// ghostwriterProvider is the provider implementation.
type ghostwriterProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *ghostwriterProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ghostwriter"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *ghostwriterProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A provider to create project resources for ghostwriter.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The graphql endpoint for the ghostwriter API. May also be provided via the GHOSTWRITER_ENDPOINT environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for the ghostwriter API. May also be provided via the GHOSTWRITER_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"tls_insecure": schema.BoolAttribute{
				Description: "Whether to skip TLS verification when connecting to the API endpoint. May also be provided via the GHOSTWRITER_TLS_INSECURE environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a ghostwriter API client for data sources and resources.
func (p *ghostwriterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Ghostwriter API client...")
	// Retrieve provider data from configuration
	var config ghostwriterProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown Ghostwriter API Endpoint",
			"The provider cannot create the Ghostwriter API client as there is an unknown configuration value for the Ghostwriter Graphql endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GHOSTWRITER_ENDPOINT environment variable.",
		)
	}

	if config.Apikey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Ghostwriter API Key",
			"The provider cannot create the Ghostwriter API client as there is an unknown configuration value for the Ghostwriter API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GHOSTWRITER_API_KEY environment variable.",
		)
	}

	var tls_insecure bool
	if config.TlsInsecure.IsUnknown() {
		tls_insecure = false
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	endpoint := os.Getenv("GHOSTWRITER_ENDPOINT")
	api_key := os.Getenv("GHOSTWRITER_API_KEY")
	tls_insecure = os.Getenv("GHOSTWRITER_TLS_INSECURE") == "false"

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.Apikey.IsNull() {
		api_key = config.Apikey.ValueString()
	}

	if !config.TlsInsecure.IsNull() {
		tls_insecure = config.TlsInsecure.ValueBool()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Ghostwriter API Endpoint",
			"The provider cannot create the Ghostwriter API client as there is a missing or empty value for the Ghostwriter Graphql endpoint. "+
				"Set the environment value in the configuration or use the GHOSTWRITER_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if api_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Ghostwriter API Key",
			"The provider cannot create the Ghostwriter API client as there is a missing or empty value for the Ghostwriter API key. "+
				"Set the api key value in the configuration or use the GHOSTWRITER_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "ghostwriter_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "ghostwriter_api_key", api_key)
	ctx = tflog.SetField(ctx, "ghostwriter_tls_insecure", tls_insecure)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "ghostwriter_api_key")

	tflog.Debug(ctx, "Creating Ghostwriter graphql client")

	// Create a new Ghostwriter client using the configuration values
	var httpClient *http.Client
	var httpctx context.Context
	if tls_insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient = &http.Client{Transport: tr}
		httpctx = context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	} else {
		httpctx = context.Background()
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: api_key},
	)
	httpClient = oauth2.NewClient(httpctx, src)
	client := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))

	// Make the Ghostwriter client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Ghostwriter API client configured.", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *ghostwriterProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewactivitytypeDataSource,
		NewserverproviderDataSource,
		NewserverroleDataSource,
		NewprojectDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *ghostwriterProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewdomainResource,
		NewdomainCheckoutResource,
		NewstaticserverCheckoutResource,
		NewstaticserverResource,
		NewcloudserverResource,
		NewoplogResource,
		NewdomainserverResource,
	}
}
