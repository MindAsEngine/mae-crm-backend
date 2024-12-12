package domain

type Integration struct {
    ID          int    `json:"id"`
    AudienceID  int    `json:"audience_id"`
    Platform    string `json:"platform"` // Например: Google Ads, Facebook
    Status      string `json:"status"`
    CreatedAt   string `json:"created_at"`
}
