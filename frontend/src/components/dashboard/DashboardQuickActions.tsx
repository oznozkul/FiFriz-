import {
  ClipboardCheck,
  FileText,
  FolderPlus,
  Plug,
  Receipt,
  ShieldCheck,
  UploadCloud,
} from "lucide-react";
import { useNavigate } from "react-router-dom";
import type { FolderLabels } from "../../hooks/useFolderLabels";

type DashboardQuickActionsProps = {
  isPersonal: boolean;
  folderLabels: FolderLabels;
  onUpload: () => void;
  onUploadWithTags: (tagNames: string[]) => void;
};

type QuickAction = {
  label: string;
  icon: typeof UploadCloud;
  onClick: () => void;
};

export function DashboardQuickActions({ isPersonal, folderLabels, onUpload, onUploadWithTags }: DashboardQuickActionsProps) {
  const navigate = useNavigate();

  const personalActions: QuickAction[] = [
    { label: "Belge yükle", icon: UploadCloud, onClick: onUpload },
    { label: "Fatura ekle", icon: Receipt, onClick: () => onUploadWithTags(["Fatura", "fatura"]) },
    { label: "Garanti belgesi ekle", icon: ShieldCheck, onClick: () => onUploadWithTags(["Garanti", "garanti"]) },
    { label: "Kira sözleşmesi ekle", icon: FileText, onClick: () => onUploadWithTags(["Kira", "kira", "Kira Sözleşmesi"]) },
  ];

  const organizationActions: QuickAction[] = [
    { label: "Belge yükle", icon: UploadCloud, onClick: onUpload },
    { label: folderLabels.isPersonal ? "Kategori oluştur" : "Klasör oluştur", icon: FolderPlus, onClick: () => navigate("/folders") },
    { label: "Onay süreci başlat", icon: ClipboardCheck, onClick: () => navigate("/documents") },
    { label: "Entegrasyon ekle", icon: Plug, onClick: () => navigate("/integrations") },
  ];

  const actions = isPersonal ? personalActions : organizationActions;

  return (
    <section className="panel rounded-2xl p-5">
      <h2 className="text-sm font-bold text-ink">Hızlı aksiyonlar</h2>
      <p className="mt-1 text-sm text-slate-500">Sık kullandığınız işlemlere tek tıkla ulaşın.</p>
      <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {actions.map((action) => {
          const Icon = action.icon;
          return (
            <button
              key={action.label}
              type="button"
              className="flex items-center gap-3 rounded-xl border border-line bg-white px-4 py-3 text-left text-sm font-semibold text-slate-700 transition hover:border-violet-ring hover:bg-violet-soft/40 hover:text-ink active:scale-[0.99] focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
              onClick={action.onClick}
            >
              <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
                <Icon className="h-4 w-4" />
              </span>
              <span>{action.label}</span>
            </button>
          );
        })}
      </div>
    </section>
  );
}
