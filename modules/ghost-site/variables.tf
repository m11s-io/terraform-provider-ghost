# --- Site settings ---

variable "title" {
  description = "Publication title."
  type        = string
}

variable "description" {
  description = "Short publication description."
  type        = string
  default     = ""
}

variable "locale" {
  description = "Site language/locale code (e.g. en, fr, de)."
  type        = string
  default     = "en"
}

variable "timezone" {
  description = "IANA timezone identifier (e.g. Europe/Paris, Etc/UTC)."
  type        = string
  default     = "Etc/UTC"
}

# --- Social accounts ---

variable "twitter" {
  description = "Twitter/X handle (e.g. @ghost)."
  type        = string
  default     = ""
}

variable "facebook" {
  description = "Facebook page name."
  type        = string
  default     = ""
}

variable "threads" {
  description = "Threads handle."
  type        = string
  default     = ""
}

variable "bluesky" {
  description = "Bluesky handle."
  type        = string
  default     = ""
}

variable "mastodon" {
  description = "Mastodon profile URL."
  type        = string
  default     = ""
}

variable "tiktok" {
  description = "TikTok handle."
  type        = string
  default     = ""
}

variable "youtube" {
  description = "YouTube channel URL or handle."
  type        = string
  default     = ""
}

variable "instagram" {
  description = "Instagram handle."
  type        = string
  default     = ""
}

variable "linkedin" {
  description = "LinkedIn profile or page URL."
  type        = string
  default     = ""
}

# --- Integration ---

variable "integration_name" {
  description = "Name of the custom Ghost integration. Shown in Ghost Admin → Integrations."
  type        = string
  default     = "Managed by OpenTofu"
}

variable "integration_description" {
  description = "Optional description for the integration."
  type        = string
  default     = "Managed by OpenTofu"
}

# --- Webhooks ---

variable "webhooks" {
  description = "Map of webhooks to attach to the integration. Key is an arbitrary local name."
  type = map(object({
    name       = string
    event      = string
    target_url = string
    secret     = optional(string, "")
  }))
  default = {}
}
