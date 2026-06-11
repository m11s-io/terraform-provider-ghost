package ghost

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ClientConfig configures a Ghost Admin API client.
type ClientConfig struct {
	URL    string
	APIKey string // format: "<id>:<secret_hex>"
}

// Client is an HTTP client for the Ghost Admin API.
type Client struct {
	baseURL string
	keyID   string
	secret  []byte
	http    *http.Client
}

// NewClient creates a Ghost Admin API client from a ClientConfig.
// The APIKey must be in the format returned by Ghost: "<id>:<hex_secret>".
func NewClient(cfg ClientConfig) (*Client, error) {
	parts := strings.SplitN(cfg.APIKey, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("ghost: api_key must be in the format <id>:<hex_secret>")
	}
	secret, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("ghost: decoding api_key secret: %w", err)
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.URL, "/"),
		keyID:   parts[0],
		secret:  secret,
		http:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) token() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		Audience:  jwt.ClaimStrings{"/ghost/api/admin/"},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t.Header["kid"] = c.keyID
	return t.SignedString(c.secret)
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	tok, err := c.token()
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/ghost/api/admin"+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Ghost "+tok)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Version", "v5.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	return resp, nil
}

func expectStatus(resp *http.Response, codes ...int) error {
	for _, code := range codes {
		if resp.StatusCode == code {
			return nil
		}
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

// --- Settings ---

// Settings represents a subset of Ghost site settings manageable via Terraform.
type Settings struct {
	Title           string `json:"title,omitempty"`
	Description     string `json:"description,omitempty"`
	Lang            string `json:"lang,omitempty"`
	Timezone        string `json:"timezone,omitempty"`
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	Twitter         string `json:"twitter,omitempty"`
	Facebook        string `json:"facebook,omitempty"`
}

type settingsEnvelope struct {
	Settings []settingKV `json:"settings"`
}

type settingKV struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// GetSettings fetches the current Ghost site settings.
func (c *Client) GetSettings(ctx context.Context) (*Settings, error) {
	resp, err := c.do(ctx, http.MethodGet, "/settings/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var env settingsEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("decoding settings: %w", err)
	}

	s := &Settings{}
	for _, kv := range env.Settings {
		v, _ := kv.Value.(string)
		switch kv.Key {
		case "title":
			s.Title = v
		case "description":
			s.Description = v
		case "lang":
			s.Lang = v
		case "timezone":
			s.Timezone = v
		case "meta_title":
			s.MetaTitle = v
		case "meta_description":
			s.MetaDescription = v
		case "twitter":
			s.Twitter = v
		case "facebook":
			s.Facebook = v
		}
	}
	return s, nil
}

// UpdateSettings writes Ghost site settings via PUT /settings/.
func (c *Client) UpdateSettings(ctx context.Context, s Settings) error {
	kvs := []settingKV{
		{Key: "title", Value: s.Title},
		{Key: "description", Value: s.Description},
		{Key: "lang", Value: s.Lang},
		{Key: "timezone", Value: s.Timezone},
		{Key: "meta_title", Value: s.MetaTitle},
		{Key: "meta_description", Value: s.MetaDescription},
		{Key: "twitter", Value: s.Twitter},
		{Key: "facebook", Value: s.Facebook},
	}
	env := settingsEnvelope{Settings: kvs}
	resp, err := c.do(ctx, http.MethodPut, "/settings/", env)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return expectStatus(resp, http.StatusOK)
}

// --- Webhooks ---

// Webhook represents a Ghost webhook.
type Webhook struct {
	ID             string `json:"id,omitempty"`
	Event          string `json:"event"`
	TargetURL      string `json:"target_url"`
	Name           string `json:"name,omitempty"`
	Secret         string `json:"secret,omitempty"`
	APIVersion     string `json:"api_version,omitempty"`
	IntegrationID  string `json:"integration_id,omitempty"`
}

type webhookEnvelope struct {
	Webhooks []Webhook `json:"webhooks"`
}

// CreateWebhook creates a new Ghost webhook.
func (c *Client) CreateWebhook(ctx context.Context, w Webhook) (*Webhook, error) {
	env := webhookEnvelope{Webhooks: []Webhook{w}}
	resp, err := c.do(ctx, http.MethodPost, "/webhooks/", env)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusCreated); err != nil {
		return nil, err
	}
	var out webhookEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding webhook: %w", err)
	}
	if len(out.Webhooks) == 0 {
		return nil, fmt.Errorf("no webhook in response")
	}
	return &out.Webhooks[0], nil
}

// UpdateWebhook updates an existing Ghost webhook.
func (c *Client) UpdateWebhook(ctx context.Context, id string, w Webhook) (*Webhook, error) {
	env := webhookEnvelope{Webhooks: []Webhook{w}}
	resp, err := c.do(ctx, http.MethodPut, "/webhooks/"+id+"/", env)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var out webhookEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding webhook: %w", err)
	}
	if len(out.Webhooks) == 0 {
		return nil, fmt.Errorf("no webhook in response")
	}
	return &out.Webhooks[0], nil
}

// DeleteWebhook deletes a Ghost webhook by ID.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/webhooks/"+id+"/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return expectStatus(resp, http.StatusNoContent)
}
