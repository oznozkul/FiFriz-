package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"fifriz/internal/applog"
)

func DiscoverIMAPSettings(email, password string) (IMAPSettings, error) {
	email = strings.TrimSpace(email)
	domain := emailDomain(email)
	if domain == "" {
		return IMAPSettings{}, fmt.Errorf("gecerli bir e-posta adresi girin")
	}

	localPart := email
	if at := strings.LastIndex(email, "@"); at > 0 {
		localPart = email[:at]
	}

	usernames := []string{email}
	if localPart != "" && localPart != email {
		usernames = append(usernames, localPart)
	}

	type connectionAttempt struct {
		host           string
		port           int
		mode           string
		insecureSkipTLS bool
	}

	attempts := make([]connectionAttempt, 0, len(candidateIMAPHosts(domain))*4)
	for _, host := range candidateIMAPHosts(domain) {
		for _, insecure := range []bool{false, true} {
			attempts = append(attempts,
				connectionAttempt{host: host, port: 993, mode: "ssl", insecureSkipTLS: insecure},
				connectionAttempt{host: host, port: 143, mode: "starttls", insecureSkipTLS: insecure},
			)
		}
	}

	var lastErr error
	for _, attempt := range attempts {
		for _, username := range usernames {
			settings := defaultIMAPSettings()
			settings.Host = attempt.host
			settings.Port = attempt.port
			settings.SecurityMode = attempt.mode
			settings.UseTLS = attempt.mode != "none"
			settings.Username = username
			settings.InsecureSkipTLS = attempt.insecureSkipTLS
			if net.ParseIP(attempt.host) != nil {
				settings.TLSServerName = "mail." + domain
			}

			client, err := connectIMAP(settings, password)
			if err == nil {
				_ = client.Logout()
				return settings, nil
			}
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("IMAP sunucusuna baglanilamadi")
	}
	return IMAPSettings{}, userFriendlyIMAPError(lastErr)
}

func candidateIMAPHosts(domain string) []string {
	seen := make(map[string]struct{}, 3)
	result := make([]string, 0, 3)
	for _, host := range []string{"mail." + domain, "imap." + domain, domain} {
		host = NormalizeIMAPHost(host)
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		result = append(result, host)
	}
	return result
}

func IMAPSettingsToMap(settings IMAPSettings) map[string]any {
	payload := map[string]any{
		"host":              settings.Host,
		"port":              settings.Port,
		"username":          settings.Username,
		"useTLS":            settings.UseTLS,
		"securityMode":      settings.ConnectionMode(),
		"mailbox":           settings.Mailbox,
		"syncFromDate":      settings.SyncFromDate.Format("2006-01-02"),
		"importAttachments": settings.ImportAttachments,
		"importEmailBody":   settings.ImportEmailBody,
		"allowedExtensions": settings.AllowedExtensions,
		"maxEmailsPerSync":  settings.MaxEmailsPerSync,
		"insecureSkipTLS":   settings.InsecureSkipTLS,
	}
	if settings.TLSServerName != "" {
		payload["tlsServerName"] = settings.TLSServerName
	}
	return payload
}

func EnsureIMAPAccountSettings(ctx context.Context, repository *Repository, credentials *CredentialService, account *IntegrationAccount) error {
	if repository == nil || credentials == nil || account == nil {
		return fmt.Errorf("imap settings preparation is not configured")
	}
	svc := &Service{repository: repository, credentials: credentials}
	_, err := svc.EnsureIMAPSettings(ctx, account)
	return err
}

func (s *Service) EnsureIMAPSettings(ctx context.Context, account *IntegrationAccount) (IMAPSettings, error) {
	if account == nil {
		return IMAPSettings{}, fmt.Errorf("integration account is required")
	}

	if settings, err := ParseIMAPSettings(account.Settings); err == nil {
		return settings, nil
	}

	email := ""
	if account.EmailAddress != nil {
		email = strings.TrimSpace(*account.EmailAddress)
	}
	if email == "" {
		email = extractEmailFromIMAPAccount(account)
	}
	if email == "" {
		return IMAPSettings{}, fmt.Errorf("IMAP e-posta adresi gerekli")
	}

	password, err := s.credentials.GetPassword(account.EncryptedCredentials)
	if err != nil || strings.TrimSpace(password) == "" {
		return IMAPSettings{}, ErrIMAPPasswordRequired
	}

	discovered, err := DiscoverIMAPSettings(email, password)
	if err != nil {
		return IMAPSettings{}, err
	}
	if preserved, ok := preservedIMAPSyncFromDate(account.Settings); ok {
		discovered.SyncFromDate = preserved
	}

	encoded, err := encodeJSONMap(IMAPSettingsToMap(discovered))
	if err != nil {
		return IMAPSettings{}, err
	}

	account.Settings = encoded
	if err := s.repository.Update(ctx, account); err != nil {
		applog.Warn("failed to persist discovered IMAP settings: integration_id=%s error=%v", account.ID, err)
	}

	return discovered, nil
}

func extractEmailFromIMAPAccount(account *IntegrationAccount) string {
	if account == nil {
		return ""
	}

	candidates := []string{account.DisplayName}
	if len(account.Settings) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(account.Settings, &payload); err == nil {
			if username := stringValue(payload["username"]); strings.Contains(username, "@") {
				candidates = append([]string{username}, candidates...)
			}
		}
	}

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if strings.Contains(candidate, "@") && emailDomain(candidate) != "" {
			return candidate
		}
	}
	return ""
}
