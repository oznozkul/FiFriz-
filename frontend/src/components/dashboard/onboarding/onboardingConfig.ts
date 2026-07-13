import type { FolderLabels } from "../../../hooks/useFolderLabels";

export type OnboardingStepId =
  | "upload-document"
  | "select-category"
  | "add-important-date"
  | "search"
  | "ai-summary"
  | "workspace-created"
  | "invite-member"
  | "create-folder"
  | "start-workflow"
  | "connect-integration";

export type OnboardingStepAction = "upload";

export type OnboardingStepDef = {
  id: OnboardingStepId;
  title: string;
  description: string;
  href?: string;
  action?: OnboardingStepAction;
};

export function getPersonalOnboardingSteps(labels: FolderLabels): OnboardingStepDef[] {
  return [
    {
      id: "upload-document",
      title: "İlk belgeni yükle",
      description: "PDF, Word veya görsel dosyanızı ekleyerek başlayın.",
      action: "upload",
    },
    {
      id: "select-category",
      title: `${labels.singular} seç`,
      description: `Belgelerinizi düzenlemek için bir ${labels.singularLower} oluşturun veya seçin.`,
      href: "/folders",
    },
    {
      id: "add-important-date",
      title: "Önemli tarih ekle",
      description: "Son ödeme, yenileme veya hatırlatma tarihi ekleyin.",
      href: "/important-dates",
    },
    {
      id: "search",
      title: "Arama yap",
      description: "Belgelerinizde hızlı veya akıllı arama deneyin.",
      href: "/documents",
    },
    {
      id: "ai-summary",
      title: "AI ile özet oluştur",
      description: "Bir belgenin özetini yapay zeka ile oluşturun.",
      href: "/summaries",
    },
  ];
}

export function getOrganizationOnboardingSteps(labels: FolderLabels): OnboardingStepDef[] {
  return [
    {
      id: "workspace-created",
      title: "Çalışma alanı oluşturuldu",
      description: "Kurumsal çalışma alanınız hazır.",
    },
    {
      id: "invite-member",
      title: "Üye davet et",
      description: "Ekibinizi çalışma alanına ekleyin.",
      href: "/settings/members",
    },
    {
      id: "create-folder",
      title: `${labels.singular} oluştur`,
      description: `Belgeleri organize etmek için ${labels.pluralLower} yapısı kurun.`,
      href: "/folders",
    },
    {
      id: "upload-document",
      title: "Belge yükle",
      description: "İlk kurumsal belgenizi yükleyin.",
      action: "upload",
    },
    {
      id: "start-workflow",
      title: "Onay süreci başlat",
      description: "Bir belge için onay akışı oluşturun.",
      href: "/documents",
    },
    {
      id: "connect-integration",
      title: "Entegrasyon bağla",
      description: "Gmail, IMAP veya diğer kaynakları bağlayın.",
      href: "/integrations",
    },
  ];
}
