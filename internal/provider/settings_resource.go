package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/m11s-io/terraform-provider-ghost/internal/ghost"
)

var _ resource.Resource = &SettingsResource{}

type SettingsResource struct {
	client *ghost.Client
}

type SettingsModel struct {
	Title           types.String `tfsdk:"title"`
	Description     types.String `tfsdk:"description"`
	Locale          types.String `tfsdk:"locale"`
	Timezone        types.String `tfsdk:"timezone"`
	MetaTitle       types.String `tfsdk:"meta_title"`
	MetaDescription types.String `tfsdk:"meta_description"`
	// Social accounts
	Twitter   types.String `tfsdk:"twitter"`
	Facebook  types.String `tfsdk:"facebook"`
	Threads   types.String `tfsdk:"threads"`
	Bluesky   types.String `tfsdk:"bluesky"`
	Mastodon  types.String `tfsdk:"mastodon"`
	Tiktok    types.String `tfsdk:"tiktok"`
	Youtube   types.String `tfsdk:"youtube"`
	Instagram types.String `tfsdk:"instagram"`
	Linkedin  types.String `tfsdk:"linkedin"`
}

func NewSettingsResource() resource.Resource {
	return &SettingsResource{}
}

func (r *SettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings"
}

func (r *SettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Ghost site settings (title, description, locale, timezone, SEO metadata, and social handles) " +
			"via `PUT /ghost/api/admin/settings/`.\n\n" +
			"There is exactly one settings resource per Ghost instance. Use `terraform import ghost_settings.main _` to import.",
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Publication title.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Short publication description shown in meta tags and the Ghost Admin.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "Site language/locale code (e.g. `en`, `de`, `fr`). Defaults to `en`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("en"),
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "IANA timezone identifier (e.g. `Europe/London`). Defaults to `Etc/UTC`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Etc/UTC"),
			},
			"meta_title": schema.StringAttribute{
				MarkdownDescription: "SEO meta title override for the homepage.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"meta_description": schema.StringAttribute{
				MarkdownDescription: "SEO meta description override for the homepage.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"twitter": schema.StringAttribute{
				MarkdownDescription: "Twitter/X handle (e.g. `@ghost`).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"facebook": schema.StringAttribute{
				MarkdownDescription: "Facebook page name (e.g. `ghost`).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"threads": schema.StringAttribute{
				MarkdownDescription: "Threads handle.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"bluesky": schema.StringAttribute{
				MarkdownDescription: "Bluesky handle.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"mastodon": schema.StringAttribute{
				MarkdownDescription: "Mastodon profile URL.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tiktok": schema.StringAttribute{
				MarkdownDescription: "TikTok handle.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"youtube": schema.StringAttribute{
				MarkdownDescription: "YouTube channel URL or handle.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"instagram": schema.StringAttribute{
				MarkdownDescription: "Instagram handle.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"linkedin": schema.StringAttribute{
				MarkdownDescription: "LinkedIn profile or page URL.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
		},
	}
}

func (r *SettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateSettings(ctx, modelToSettings(data)); err != nil {
		resp.Diagnostics.AddError("Error applying settings", err.Error())
		return
	}
	r.readInto(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readInto(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateSettings(ctx, modelToSettings(data)); err != nil {
		resp.Diagnostics.AddError("Error updating settings", err.Error())
		return
	}
	r.readInto(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is a no-op: site settings cannot be deleted, only overwritten.
func (r *SettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *SettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data SettingsModel
	r.readInto(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SettingsResource) readInto(ctx context.Context, data *SettingsModel, diags *diag.Diagnostics) {
	s, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Error reading settings", err.Error())
		return
	}
	data.Title = types.StringValue(s.Title)
	data.Description = types.StringValue(s.Description)
	data.Locale = types.StringValue(s.Locale)
	data.Timezone = types.StringValue(s.Timezone)
	data.MetaTitle = types.StringValue(s.MetaTitle)
	data.MetaDescription = types.StringValue(s.MetaDescription)
	data.Twitter = types.StringValue(s.Twitter)
	data.Facebook = types.StringValue(s.Facebook)
	data.Threads = types.StringValue(s.Threads)
	data.Bluesky = types.StringValue(s.Bluesky)
	data.Mastodon = types.StringValue(s.Mastodon)
	data.Tiktok = types.StringValue(s.Tiktok)
	data.Youtube = types.StringValue(s.Youtube)
	data.Instagram = types.StringValue(s.Instagram)
	data.Linkedin = types.StringValue(s.Linkedin)
}

func modelToSettings(m SettingsModel) ghost.Settings {
	return ghost.Settings{
		Title:           m.Title.ValueString(),
		Description:     m.Description.ValueString(),
		Locale:          m.Locale.ValueString(),
		Timezone:        m.Timezone.ValueString(),
		MetaTitle:       m.MetaTitle.ValueString(),
		MetaDescription: m.MetaDescription.ValueString(),
		Twitter:         m.Twitter.ValueString(),
		Facebook:        m.Facebook.ValueString(),
		Threads:         m.Threads.ValueString(),
		Bluesky:         m.Bluesky.ValueString(),
		Mastodon:        m.Mastodon.ValueString(),
		Tiktok:          m.Tiktok.ValueString(),
		Youtube:         m.Youtube.ValueString(),
		Instagram:       m.Instagram.ValueString(),
		Linkedin:        m.Linkedin.ValueString(),
	}
}
