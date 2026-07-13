import { outlookExtensionOptions, type OutlookFormState } from "./outlookFormShared";

type OutlookSettingsFieldsProps = {
  state: OutlookFormState;
  onChange: (patch: Partial<OutlookFormState>) => void;
};

export function OutlookSettingsFields({ state, onChange }: OutlookSettingsFieldsProps) {
  const toggleExtension = (extension: string) => {
    onChange({
      allowedExtensions: state.allowedExtensions.includes(extension)
        ? state.allowedExtensions.filter((item) => item !== extension)
        : [...state.allowedExtensions, extension],
    });
  };

  return (
    <div className="space-y-4 rounded-lg border border-line bg-slate-50 p-4">
      <p className="text-sm font-semibold text-slate-700">Outlook ayarları</p>
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block space-y-2 text-sm sm:col-span-2">
          <span className="font-semibold text-slate-600">Klasör</span>
          <input
            className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
            value={state.folder}
            onChange={(e) => onChange({ folder: e.target.value })}
            placeholder="Inbox"
          />
        </label>
        <label className="block space-y-2 text-sm">
          <span className="font-semibold text-slate-600">Hangi tarihten itibaren</span>
          <input
            type="date"
            className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
            value={state.syncFromDate}
            max={new Date().toISOString().slice(0, 10)}
            onChange={(e) => onChange({ syncFromDate: e.target.value })}
          />
        </label>
        <label className="block space-y-2 text-sm">
          <span className="font-semibold text-slate-600">Senkron başına max e-posta</span>
          <input
            type="number"
            min={1}
            max={2000}
            className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
            value={state.maxEmailsPerSync}
            onChange={(e) => onChange({ maxEmailsPerSync: e.target.value })}
          />
        </label>
      </div>

      <div className="flex flex-wrap gap-4 text-sm font-semibold text-slate-600">
        <label className="flex items-center gap-2">
          <input type="checkbox" checked={state.importAttachments} onChange={(e) => onChange({ importAttachments: e.target.checked })} />
          Ekleri içe aktar
        </label>
        <label className="flex items-center gap-2">
          <input type="checkbox" checked={state.importEmailBody} onChange={(e) => onChange({ importEmailBody: e.target.checked })} />
          Mail gövdesini içe aktar
        </label>
      </div>

      <div>
        <p className="text-sm font-semibold text-slate-600">İzin verilen uzantılar</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {outlookExtensionOptions.map((extension) => (
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
