export const googleDriveExtensionOptions = ["pdf", "docx", "xlsx", "png", "jpg", "jpeg"];

export type GoogleDriveFormState = {
  folderId: string;
  syncFromDate: string;
  includeSubfolders: boolean;
  maxFilesPerSync: string;
  allowedExtensions: string[];
};

export function defaultGoogleDriveSyncFromDate() {
  return "2025-01-01";
}

export function defaultGoogleDriveFormState(): GoogleDriveFormState {
  return {
    folderId: "",
    syncFromDate: defaultGoogleDriveSyncFromDate(),
    includeSubfolders: true,
    maxFilesPerSync: "50",
    allowedExtensions: [...googleDriveExtensionOptions],
  };
}

export function googleDriveFormStateFromSettings(settings: Record<string, unknown> | null | undefined): GoogleDriveFormState {
  const extensions = Array.isArray(settings?.allowedExtensions)
    ? settings.allowedExtensions.map(String)
    : [...googleDriveExtensionOptions];
  const folderId = settings?.folderId;

  return {
    folderId: folderId == null ? "" : String(folderId),
    syncFromDate: String(settings?.syncFromDate ?? defaultGoogleDriveSyncFromDate()),
    includeSubfolders: settings?.includeSubfolders !== false,
    maxFilesPerSync: String(settings?.maxFilesPerSync ?? 50),
    allowedExtensions: extensions.length > 0 ? extensions : [...googleDriveExtensionOptions],
  };
}

export function googleDriveSettingsFromFormState(state: GoogleDriveFormState): Record<string, unknown> {
  const folderId = state.folderId.trim();
  return {
    folderId: folderId === "" ? null : folderId,
    syncFromDate: state.syncFromDate,
    includeSubfolders: state.includeSubfolders,
    allowedExtensions: state.allowedExtensions,
    maxFilesPerSync: Number(state.maxFilesPerSync) || 50,
  };
}
