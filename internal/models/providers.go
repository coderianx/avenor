package models

type AddNewProviderReq struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

// ProviderRes is the provider representation returned to clients. It never
// contains the API key, only whether one is stored and its last four
// characters.
type ProviderRes struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	APIKeySet   bool   `json:"api_key_set"`
	APIKeyLast4 string `json:"api_key_last4"`
}

// Provider is the provider representation returned by list endpoints. It
// intentionally has no API key fields.
type Provider struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}
