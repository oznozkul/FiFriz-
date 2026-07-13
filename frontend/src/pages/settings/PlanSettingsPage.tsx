import { Sparkles } from "lucide-react";
import { SettingsSection } from "./shared";

export function PlanSettingsPage() {
  return (
    <SettingsSection title="Plan" description="Abonelik planınız ve özellik limitleri.">
      <div className="rounded-xl border border-dashed border-line bg-slate-50/60 px-6 py-12 text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
          <Sparkles className="h-6 w-6" />
        </div>
        <h2 className="text-base font-bold text-ink">Yakında</h2>
        <p className="mx-auto mt-2 max-w-lg text-sm text-slate-500">
          Abonelik planları ve özellik karşılaştırması yakında eklenecek. Şu an tüm temel özellikler kullanılabilir.
        </p>
      </div>
    </SettingsSection>
  );
}
