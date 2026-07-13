export const outlookExtensionOptions = ["pdf", "docx", "xlsx", "png", "jpg", "jpeg"];

export type OutlookFormState = {
  folder: string;
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

export function defaultOutlookFormState(): OutlookFormState {
  return {
    folder: "Inbox",
    syncFromDate: defaultSyncFromDate(),
    importAttachments: true,
    importEmailBody: true,
    maxEmailsPerSync: "500",
    allowedExtensions: [...outlookExtensionOptions],
  };
}

export function outlookFormStateFromSettings(settings: Record<string, unknown> | null | undefined): OutlookFormState {
  const extensions = Array.isArray(settings?.allowedExtensions)
    ? settings.allowedExtensions.map(String)
    : [...outlookExtensionOptions];

  return {
    folder: String(settings?.folder ?? "Inbox"),
    syncFromDate: String(settings?.syncFromDate ?? defaultSyncFromDate()),
    importAttachments: settings?.importAttachments !== false,
    importEmailBody: settings?.importEmailBody !== false,
    maxEmailsPerSync: String(settings?.maxEmailsPerSync ?? 500),
    allowedExtensions: extensions.length > 0 ? extensions : [...outlookExtensionOptions],
  };
}

export function outlookSettingsFromFormState(state: OutlookFormState): Record<string, unknown> {
  return {
    folder: state.folder.trim() || "Inbox",
    syncFromDate: state.syncFromDate,
    importAttachments: state.importAttachments,
    importEmailBody: state.importEmailBody,
    allowedExtensions: state.allowedExtensions,
    maxEmailsPerSync: Number(state.maxEmailsPerSync) || 500,
  };
}
