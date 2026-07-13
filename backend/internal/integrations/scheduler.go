package integrations

import (
	"context"
	"time"

	"fifriz/internal/applog"
	"fifriz/internal/jobs"
)

// RunSyncScheduler periodically enqueues sync jobs for active mail integrations.
func (s *Service) RunSyncScheduler(ctx context.Context) {
	if s == nil || s.syncInterval <= 0 {
		applog.Info("integration sync scheduler disabled")
		return
	}

	applog.Info("integration sync scheduler started: interval=%s", s.syncInterval)
	s.enqueueDueScheduledSyncs(ctx)

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.enqueueDueScheduledSyncs(ctx)
		}
	}
}

func (s *Service) enqueueDueScheduledSyncs(ctx context.Context) {
	if s.repository == nil {
		return
	}

	dueBefore := time.Now().UTC().Add(-s.syncInterval)
	providers := s.registry.SyncableProviderNames()
	accounts, err := s.repository.ListDueForScheduledSync(ctx, dueBefore, providers)
	if err != nil {
		applog.Warn("integration scheduled sync listing failed: error=%v", err)
		return
	}

	for index := range accounts {
		if err := s.enqueueSyncForAccount(ctx, &accounts[index], SyncTypeScheduled, false); err != nil {
			applog.Warn("integration scheduled sync enqueue failed: integration_id=%s error=%v", accounts[index].ID, err)
		}
	}
}

func (s *Service) enqueueSyncForAccount(ctx context.Context, account *IntegrationAccount, syncType string, auditManual bool) error {
	if account == nil || !account.IsActive {
		return nil
	}
	plugin, ok := s.registry.Get(account.Provider)
	if !ok || !plugin.Capabilities().SupportsScheduling {
		return nil
	}
	if account.Status == StatusSyncing || account.Status == StatusSyncPending {
		return nil
	}

	account.Status = StatusSyncPending
	if err := s.repository.Update(ctx, account); err != nil {
		return err
	}

	if err := s.enqueueIntegrationSyncJob(ctx, account, syncType); err != nil {
		return err
	}

	return nil
}

func (s *Service) enqueueIntegrationSyncJob(ctx context.Context, account *IntegrationAccount, syncType string) error {
	if err := jobs.EnqueueSyncIntegration(ctx, s.jobQueue, account.TenantID, account.ID, account.UserID, syncType); err != nil {
		applog.Warn("failed to enqueue sync_integration: integration_id=%s error=%v", account.ID, err)
		return err
	}
	return nil
}
