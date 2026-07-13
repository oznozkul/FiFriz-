import { googleDriveExtensionOptions, type GoogleDriveFormState } from "./googleDriveFormShared";

type GoogleDriveSettingsFieldsProps = {
  state: GoogleDriveFormState;
  onChange: (patch: Partial<GoogleDriveFormState>) => void;
  providerLabel?: string;
  folderPlaceholder?: string;
  folderHelpText?: string;
};

export function GoogleDriveSettingsFields({
  state,
  onChange,
  providerLabel = "Google Drive",
  folderPlaceholder,
  folderHelpText,
}: GoogleDriveSettingsFieldsProps) {
  const toggleExtension = (extension: string) => {
    onChange({
      allowedExtensions: state.allowedExtensions.includes(extension)
        ? state.allowedExtensions.filter((item) => item !== extension)
        : [...state.allowedExtensions, extension],
    });
  };

  return (
    <div className="space-y-4 rounded-lg border border-line bg-slate-50 p-4">
      <p className="text-sm font-semibold text-slate-700">{providerLabel} ayarları</p>
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block space-y-2 text-sm sm:col-span-2">
          <span className="font-semibold text-slate-600">Klasör ID</span>
          <input
            className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
            value={state.folderId}
            onChange={(e) => onChange({ folderId: e.target.value })}
            placeholder={folderPlaceholder ?? "Boş bırakılırsa Drive kök klasörü"}
          />
          <p className="text-xs text-slate-500">
            {folderHelpText ?? "Belirli bir klasörden başlamak için Google Drive klasör ID'sini girin."}
          </p>
        </label>
        <label className="block space-y-2 text-sm">
          <span className="font-semibold text-slate-600">Senkron başlangıç tarihi</span>
          <input
            type="date"
            className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
            value={state.syncFromDate}
            onChange={(e) => onChange({ syncFromDate: e.target.value })}
          />
        </label>
        <label className="block space-y-2 text-sm">
          <span className="font-semibold text-slate-600">Senkron başına max dosya</span>
          <input
            type="number"
            min={1}
            max={500}
            className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
            value={state.maxFilesPerSync}
            onChange={(e) => onChange({ maxFilesPerSync: e.target.value })}
          />
        </label>
      </div>

      <label className="flex items-center gap-2 text-sm font-semibold text-slate-600">
        <input type="checkbox" checked={state.includeSubfolders} onChange={(e) => onChange({ includeSubfolders: e.target.checked })} />
        Alt klasörleri dahil et
      </label>

      <div>
        <p className="text-sm font-semibold text-slate-600">İzin verilen uzantılar</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {googleDriveExtensionOptions.map((extension) => (
            <label
              key={extension}
              className={`inline-flex cursor-pointer items-center gap-2 rounded-full border px-3 py-1 text-xs font-semibold ${state.allowedExtensions.includes(extension) ? "border-violet-brand bg-violet-soft text-violet-brand" : "border-line bg-white text-slate-500"}`}
            >
              <input className="sr-only" type="checkbox" checked={state.allowedExtensions.includes(extension)} onChange={() => toggleExtension(extension)} />
              {extension.toUpperCase()}
            </label>
          ))}
        </div>
      </div>
    </div>
  );
}
