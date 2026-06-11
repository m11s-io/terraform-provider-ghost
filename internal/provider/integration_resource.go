package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/m11s-io/terraform-provider-ghost/internal/ghost"
)

var _ resource.Resource = &IntegrationResource{}

type IntegrationResource struct {
	client *ghost.Client
}

type IntegrationModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ContentAPIKey  types.String `tfsdk:"content_api_key"`
	AdminAPIKey    types.String `tfsdk:"admin_api_key"`
}

func NewIntegrationResource() resource.Resource {
	return &IntegrationResource{}
}

func (r *IntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Ghost custom integration.\n\n" +
			"Creating an integration automatically generates one Content API key and one Admin API key. " +
			"Both are exposed as sensitive computed attributes so they can be passed to other resources " +
			"(e.g. storing them in Vault via `vault_generic_secret`).\n\n" +
			"Webhooks can be attached to an integration by setting `integration_id` on `ghost_webhook` resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Ghost-assigned integration ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Integration name shown in Ghost Admin → Integrations.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description shown in the Ghost Admin.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content_api_key": schema.StringAttribute{
				MarkdownDescription: "Content API key (plain hex string). Read-only — set by Ghost on creation.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					// Ghost never changes keys unless explicitly refreshed; keep state value.
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_api_key": schema.StringAttribute{
				MarkdownDescription: "Admin API key in `<id>:<hex_secret>` format. Read-only — set by Ghost on creation.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ghost.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *ghost.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *IntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integration, err := r.client.CreateIntegration(ctx, ghost.Integration{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating integration", err.Error())
		return
	}

	integrationToModel(integration, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integration, err := r.client.GetIntegration(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration", err.Error())
		return
	}
	if integration == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	integrationToModel(integration, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integration, err := r.client.UpdateIntegration(ctx, data.ID.ValueString(), ghost.Integration{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating integration", err.Error())
		return
	}

	// Ghost always returns api_keys (with secrets) in the update response.
	integrationToModel(integration, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteIntegration(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting integration", err.Error())
	}
}

func integrationToModel(in *ghost.Integration, m *IntegrationModel) {
	m.ID = types.StringValue(in.ID)
	m.Name = types.StringValue(in.Name)
	m.Description = types.StringValue(in.Description)

	for _, key := range in.APIKeys {
		switch key.Type {
		case "content":
			m.ContentAPIKey = types.StringValue(key.Secret)
		case "admin":
			m.AdminAPIKey = types.StringValue(key.Secret)
		}
	}
}
