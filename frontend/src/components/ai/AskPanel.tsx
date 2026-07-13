import { Send } from "lucide-react";
import { PrimaryButton } from "../common";

export function AskPanel() {
  return (
    <section className="panel p-5">
      <h2 className="font-bold text-ink">AI Asistan'a Sor</h2>
      <p className="mt-1 text-sm text-slate-500">Belgelerinizle ilgili sorular sorun, yapay zeka anında cevaplasın.</p>
      <div className="mt-4 flex gap-3">
        <input className="h-12 min-w-0 flex-1 rounded-lg border border-line px-4 outline-none transition focus:border-violet-brand focus:ring-4 focus:ring-violet-soft" placeholder="Örn: Kira sözleşmemde depozito ne kadar?" />
        <PrimaryButton className="h-12 w-12 px-0" aria-label="Gönder">
          <Send className="h-5 w-5" />
        </PrimaryButton>
      </div>
    </section>
  );
}
