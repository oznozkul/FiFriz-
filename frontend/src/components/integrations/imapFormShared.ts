export function defaultSyncFromDate() {
  const date = new Date();
  date.setMonth(date.getMonth() - 1);
  return date.toISOString().slice(0, 10);
}

export function syncFromDateFromSettings(settings: Record<string, unknown> | null | undefined) {
  return String(settings?.syncFromDate ?? defaultSyncFromDate());
}

export function maxEmailsPerSyncFromSettings(settings: Record<string, unknown> | null | undefined) {
  return String(settings?.maxEmailsPerSync ?? 500);
}
