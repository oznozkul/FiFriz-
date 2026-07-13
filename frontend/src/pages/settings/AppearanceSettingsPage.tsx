import { SettingsSection } from "./shared";

export function AppearanceSettingsPage() {
  return (
    <SettingsSection title="Görünüm" description="Arayüz tercihlerinizi özelleştirin.">
      <div className="space-y-4">
        <PreferenceRow title="Açık tema" description="Varsayılan beyaz arayüz teması aktif." enabled disabled />
        <PreferenceRow title="Koyu tema" description="Koyu mod yakında eklenecek." enabled={false} disabled />
        <PreferenceRow title="Kompakt görünüm" description="Daha sıkı liste ve tablo düzeni yakında eklenecek." enabled={false} disabled />
      </div>
    </SettingsSection>
  );
}

function PreferenceRow({
  title,
  description,
  enabled,
  disabled,
}: {
  title: string;
  description: string;
  enabled: boolean;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg border border-line bg-slate-50/50 px-4 py-4">
      <div>
        <p className="text-sm font-bold text-ink">{title}</p>
        <p className="mt-1 text-sm text-slate-500">{description}</p>
      </div>
      <button
        type="button"
        disabled={disabled}
        aria-pressed={enabled}
        className={`relative h-6 w-11 shrink-0 rounded-full transition ${enabled ? "bg-violet-brand" : "bg-slate-300"} ${disabled ? "cursor-not-allowed opacity-70" : ""}`}
      >
        <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition ${enabled ? "left-5" : "left-0.5"}`} />
      </button>
    </div>
  );
}
