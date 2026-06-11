package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/m11s-io/terraform-provider-ghost/internal/ghost"
)

var _ resource.Resource = &WebhookResource{}

type WebhookResource struct {
	client *ghost.Client
}

type WebhookModel struct {
	ID            types.String `tfsdk:"id"`
	Event         types.String `tfsdk:"event"`
	TargetURL     types.String `tfsdk:"target_url"`
	Name          types.String `tfsdk:"name"`
	Secret        types.String `tfsdk:"secret"`
	APIVersion    types.String `tfsdk:"api_version"`
	IntegrationID types.String `tfsdk:"integration_id"`
}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Ghost webhook. Webhooks fire HTTP POST requests to a target URL when specific " +
			"Ghost events occur (e.g. `post.published`, `member.added`).\n\n" +
			"See the [Ghost webhook docs](https://ghost.org/docs/webhooks/) for the full list of supported events.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Ghost-assigned webhook ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event": schema.StringAttribute{
				MarkdownDescription: "Ghost event that triggers this webhook (e.g. `post.published`, `member.added`). Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_url": schema.StringAttribute{
				MarkdownDescription: "URL that receives the HTTP POST payload.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable webhook name shown in the Ghost Admin.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Optional HMAC secret used to sign webhook payloads (`X-Ghost-Signature` header).",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Default:             stringdefault.StaticString(""),
			},
			"api_version": schema.StringAttribute{
				MarkdownDescription: "Ghost Admin API version for the webhook payload format (e.g. `v5.0`). Defaults to `v5.0`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("v5.0"),
			},
			"integration_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Ghost integration that owns this webhook. Required when using an integration API key.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wh, err := r.client.CreateWebhook(ctx, modelToWebhook(data))
	if err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	webhookToModel(wh, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op: Ghost has no GET endpoint for individual webhooks.
// State written on Create/Update is kept as-is.
func (r *WebhookResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wh, err := r.client.UpdateWebhook(ctx, data.ID.ValueString(), modelToWebhook(data))
	if err != nil {
		resp.Diagnostics.AddError("Error updating webhook", err.Error())
		return
	}

	webhookToModel(wh, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteWebhook(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting webhook", err.Error())
	}
}

func modelToWebhook(m WebhookModel) ghost.Webhook {
	return ghost.Webhook{
		Event:         m.Event.ValueString(),
		TargetURL:     m.TargetURL.ValueString(),
		Name:          m.Name.ValueString(),
		Secret:        m.Secret.ValueString(),
		APIVersion:    m.APIVersion.ValueString(),
		IntegrationID: m.IntegrationID.ValueString(),
	}
}

func webhookToModel(w *ghost.Webhook, m *WebhookModel) {
	m.ID = types.StringValue(w.ID)
	m.Event = types.StringValue(w.Event)
	m.TargetURL = types.StringValue(w.TargetURL)
	m.Name = types.StringValue(w.Name)
	m.APIVersion = types.StringValue(w.APIVersion)
	m.IntegrationID = types.StringValue(w.IntegrationID)
	// Don't overwrite secret from response — Ghost doesn't return it after creation.
}
