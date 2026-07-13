package integrations

const (
	DefaultMaxEmailsPerSync  = 500
	AbsoluteMaxEmailsPerSync = 2000
	MailSyncPageSize         = 100
)

// EffectiveMaxEmailsPerSync returns the per-sync mail processing limit.
// Zero or negative values use the absolute cap so date-range sync is not stuck at a tiny batch.
func EffectiveMaxEmailsPerSync(configured int) int {
	if configured <= 0 {
		return AbsoluteMaxEmailsPerSync
	}
	// Upgrade integrations that still carry the old default of 50.
	if configured == 50 {
		return DefaultMaxEmailsPerSync
	}
	if configured > AbsoluteMaxEmailsPerSync {
		return AbsoluteMaxEmailsPerSync
	}
	return configured
}
