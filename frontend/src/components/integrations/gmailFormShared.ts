export const gmailExtensionOptions = ["pdf", "docx", "xlsx", "png", "jpg", "jpeg"];

export type GmailFormState = {
  query: string;
  syncFromDate: string;
  importAttachments: boolean;
  importEmailBody: boolean;
  maxEmailsPerSync: string;
  allowedExtensions: string[];
};

export function defaultSyncFromDate() {
  const date = new Date();
  date.setMonth(date.getMonth() - 1);
  return date.toISOString().slice(0, 10);
}

export function defaultGmailFormState(): GmailFormState {
  return {
    query: "has:attachment",
    syncFromDate: defaultSyncFromDate(),
    importAttachments: true,
    importEmailBody: true,
    maxEmailsPerSync: "500",
    allowedExtensions: [...gmailExtensionOptions],
  };
}

export function gmailFormStateFromSettings(settings: Record<string, unknown> | null | undefined): GmailFormState {
  const extensions = Array.isArray(settings?.allowedExtensions)
    ? settings.allowedExtensions.map(String)
    : [...gmailExtensionOptions];

  return {
    query: String(settings?.query ?? "has:attachment"),
    syncFromDate: String(settings?.syncFromDate ?? defaultSyncFromDate()),
    importAttachments: settings?.importAttachments !== false,
    importEmailBody: settings?.importEmailBody !== false,
    maxEmailsPerSync: String(settings?.maxEmailsPerSync ?? 500),
    allowedExtensions: extensions.length > 0 ? extensions : [...gmailExtensionOptions],
  };
}

export function gmailSettingsFromFormState(state: GmailFormState): Record<string, unknown> {
  return {
    query: state.query.trim() || "has:attachment",
    syncFromDate: state.syncFromDate,
    importAttachments: state.importAttachments,
    importEmailBody: state.importEmailBody,
    allowedExtensions: state.allowedExtensions,
    maxEmailsPerSync: Number(state.maxEmailsPerSync) || 500,
  };
}
