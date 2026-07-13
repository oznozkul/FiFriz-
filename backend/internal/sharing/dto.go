package sharing

import "time"

type CreateShareLinkRequest struct {
	ShareType         string          `json:"shareType"`
	Permission        SharePermission `json:"permission"`
	Password          *string         `json:"password"`
	ExpirationPreset  string          `json:"expirationPreset"`
	ExpiresAt         *time.Time      `json:"expiresAt"`
	MaxViewCount      *int            `json:"maxViewCount"`
	AllowDownload     *bool           `json:"allowDownload"`
	AllowPrint        *bool           `json:"allowPrint"`
	AllowCopyText     *bool           `json:"allowCopyText"`
	AllowedDomains    []string        `json:"allowedDomains"`
	WatermarkEnabled  *bool           `json:"watermarkEnabled"`
	WatermarkTemplate *string         `json:"watermarkTemplate"`
	InvitedEmail      *string         `json:"invitedEmail"`
}

type BulkCreateShareLinksRequest struct {
	DocumentIDs []string `json:"documentIds"`
	CreateShareLinkRequest
}

type ShareLinkResponse struct {
	ID               string          `json:"id"`
	DocumentID       string          `json:"documentId"`
	ShareType        string          `json:"shareType"`
	Status           string          `json:"status"`
	ShareURL         string          `json:"shareUrl"`
	Token            string          `json:"token"`
	Permission       SharePermission `json:"permission"`
	ExpirationPreset string          `json:"expirationPreset,omitempty"`
	ExpiresAt        *time.Time      `json:"expiresAt"`
	MaxViewCount     *int            `json:"maxViewCount"`
	ViewCount        int             `json:"viewCount"`
	DownloadCount    int             `json:"downloadCount"`
	AllowDownload    bool            `json:"allowDownload"`
	AllowPrint       bool            `json:"allowPrint"`
	AllowCopyText    bool            `json:"allowCopyText"`
	AllowedDomains   []string        `json:"allowedDomains,omitempty"`
	WatermarkEnabled bool            `json:"watermarkEnabled"`
	IsActive         bool            `json:"isActive"`
	HasPassword      bool            `json:"hasPassword"`
	InvitedEmail     *string         `json:"invitedEmail,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
}

type ShareLinkAnalyticsResponse struct {
	ShareLinkID      string     `json:"shareLinkId"`
	DocumentID       string     `json:"documentId"`
	OpenCount        int        `json:"openCount"`
	DownloadCount    int        `json:"downloadCount"`
	LastOpenedAt     *time.Time `json:"lastOpenedAt"`
	LastDownloadedAt *time.Time `json:"lastDownloadedAt"`
	LastIPAddress    *string    `json:"lastIpAddress"`
	LastUserAgent    *string    `json:"lastUserAgent"`
	CreatedByUserID  string     `json:"createdByUserId"`
	CreatedByName    string     `json:"createdByName"`
	Status           string     `json:"status"`
}

type PublicShareResponse struct {
	DocumentName     string `json:"documentName"`
	FileExtension    string `json:"fileExtension"`
	FileSize         int64  `json:"fileSize"`
	PreviewURL       string `json:"previewUrl,omitempty"`
	AllowDownload    bool   `json:"allowDownload"`
	AllowPrint       bool   `json:"allowPrint"`
	AllowCopyText    bool   `json:"allowCopyText"`
	WatermarkEnabled bool   `json:"watermarkEnabled"`
	Status           string `json:"status"`
	PasswordRequired bool   `json:"passwordRequired"`
	DomainRestricted bool   `json:"domainRestricted"`
	CanAccess        bool   `json:"canAccess"`
	Expired          bool   `json:"expired"`
	MaxViewsReached  bool   `json:"maxViewsReached"`
}

type VerifyPasswordRequest struct {
	Password string `json:"password"`
}

type VerifyPasswordResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

type ShareChannel string

const (
	ShareChannelWhatsApp ShareChannel = "whatsapp"
	ShareChannelTelegram ShareChannel = "telegram"
	ShareChannelEmail    ShareChannel = "email"
	ShareChannelCopy     ShareChannel = "copy"
	ShareChannelNative   ShareChannel = "native"
)

type ChannelLogRequest struct {
	Channel ShareChannel `json:"channel"`
}
