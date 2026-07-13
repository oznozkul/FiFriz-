import { CreditCard } from "lucide-react";
import { SettingsSection } from "./shared";

export function BillingSettingsPage() {
  return (
    <SettingsSection title="Faturalandırma" description="Abonelik planınız ve faturalandırma bilgileri.">
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-line bg-slate-50/60 px-6 py-12 text-center">
        <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
          <CreditCard className="h-6 w-6" />
        </div>
        <h2 className="text-base font-bold text-ink">Yakında</h2>
        <p className="mt-2 max-w-md text-sm text-slate-500">Abonelik ve faturalandırma yakında eklenecek.</p>
      </div>
    </SettingsSection>
  );
}
