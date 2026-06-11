terraform {
  required_providers {
    ghost = {
      source  = "m11s-io/ghost"
      version = "~> 0.1"
    }
  }
}

provider "ghost" {
  url     = "https://blog.example.com"
  api_key = var.ghost_api_key
}
