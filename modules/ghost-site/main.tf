terraform {
  required_providers {
    ghost = {
      source  = "registry.terraform.io/m11s-io/ghost"
      version = "~> 0.1"
    }
  }
}

resource "ghost_settings" "this" {
  title       = var.title
  description = var.description
  locale      = var.locale
  timezone    = var.timezone

  twitter   = var.twitter
  facebook  = var.facebook
  threads   = var.threads
  bluesky   = var.bluesky
  mastodon  = var.mastodon
  tiktok    = var.tiktok
  youtube   = var.youtube
  instagram = var.instagram
  linkedin  = var.linkedin
}

resource "ghost_integration" "this" {
  name        = var.integration_name
  description = var.integration_description
}

resource "ghost_webhook" "this" {
  for_each = var.webhooks

  name           = each.value.name
  event          = each.value.event
  target_url     = each.value.target_url
  secret         = try(each.value.secret, "")
  integration_id = ghost_integration.this.id
}
