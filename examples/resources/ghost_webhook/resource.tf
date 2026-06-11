resource "ghost_webhook" "deploy" {
  event      = "site.changed"
  target_url = "https://ci.example.com/hooks/ghost-deploy"
  name       = "Trigger deploy on site change"
}
