package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/m11s-io/terraform-provider-ghost/internal/ghost"
)

var _ provider.Provider = &GhostProvider{}
var _ provider.ProviderWithFunctions = &GhostProvider{}

type GhostProvider struct {
	version string
}

type GhostProviderModel struct {
	URL    types.String `tfsdk:"url"`
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GhostProvider{version: version}
	}
}

func (p *GhostProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ghost"
	resp.Version = p.version
}

func (p *GhostProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for managing [Ghost](https://ghost.org) self-hosted publishing platform configuration " +
			"via the Ghost Admin API.\n\n" +
			"Authenticate using a Ghost Admin API key (`<id>:<hex_secret>`) obtained from **Settings → Integrations** " +
			"in your Ghost Admin panel.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Base URL of your Ghost instance (e.g. `https://blog.example.com`). " +
					"Can also be set via the `GHOST_URL` environment variable.",
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Ghost Admin API key in `<id>:<hex_secret>` format. " +
					"Can also be set via the `GHOST_API_KEY` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *GhostProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GhostProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := firstNonEmpty(data.URL.ValueString(), os.Getenv("GHOST_URL"))
	apiKey := firstNonEmpty(data.APIKey.ValueString(), os.Getenv("GHOST_API_KEY"))

	if url == "" {
		resp.Diagnostics.AddError("Missing Ghost URL",
			"Set the `url` provider attribute or the GHOST_URL environment variable.")
		return
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing Ghost API key",
			"Set the `api_key` provider attribute or the GHOST_API_KEY environment variable.")
		return
	}

	client, err := ghost.NewClient(ghost.ClientConfig{URL: url, APIKey: apiKey})
	if err != nil {
		resp.Diagnostics.AddError("Invalid Ghost API key", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *GhostProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSettingsResource,
		NewIntegrationResource,
		NewWebhookResource,
	}
}

func (p *GhostProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *GhostProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
