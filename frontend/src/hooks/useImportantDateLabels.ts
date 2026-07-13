import { useMemo } from "react";
import { useAuth } from "../services/AuthContext";

export function useImportantDateLabels() {
  const { activeTenant } = useAuth();
  const isPersonal = activeTenant?.type === "personal";

  return useMemo(
    () => ({
      isPersonal,
      navLabel: isPersonal ? "Önemli Tarihler" : "Önemli Tarihler",
      pageTitle: isPersonal ? "Yaklaşan Tarihler" : "Takip Tarihleri",
      pageSubtitle: isPersonal
        ? "Belgelerinizdeki önemli tarihleri ve hatırlatıcıları yönetin."
        : "Kurum belgelerindeki takip tarihlerini ve sorumlulukları yönetin.",
      panelTitle: isPersonal ? "Önemli Tarihler" : "Takip Tarihleri",
      panelSubtitle: isPersonal ? "Hatırlatıcı ve son tarihler" : "Belge süreci takibi",
      singular: isPersonal ? "Hatırlatıcı" : "Takip Tarihi",
      addLabel: isPersonal ? "Hatırlatıcı Ekle" : "Takip Tarihi Ekle",
      editLabel: isPersonal ? "Hatırlatıcıyı Düzenle" : "Takip Tarihini Düzenle",
      assignedLabel: isPersonal ? "Atanan" : "Sorumlu",
      assigneeLabel: isPersonal ? "Atanan Kişi" : "Atanan Kişi",
      reminderLabel: isPersonal ? "Hatırlatma" : "Hatırlatma",
      upcomingTitle: isPersonal ? "Yaklaşan Tarihler" : "Yaklaşan Takip Tarihleri",
      overdueTitle: isPersonal ? "Gecikenler" : "Geciken Takip Tarihleri",
      assignedTitle: isPersonal ? "Bana Atananlar" : "Bana Atananlar",
      completedTitle: "Tamamlananlar",
      goToDocument: "Belgeye Git",
      emptyTitle: isPersonal ? "Henüz önemli tarih yok." : "Henüz takip tarihi yok.",
      emptyDescription: isPersonal ? "Belgeye hatırlatıcı ekleyerek başlayın." : "Belgeye takip tarihi ekleyerek başlayın.",
      createdToast: isPersonal ? "Hatırlatıcı eklendi." : "Takip tarihi eklendi.",
      updatedToast: isPersonal ? "Hatırlatıcı güncellendi." : "Takip tarihi güncellendi.",
      deletedToast: isPersonal ? "Hatırlatıcı silindi." : "Takip tarihi silindi.",
      completedToast: isPersonal ? "Hatırlatıcı tamamlandı." : "Takip tarihi tamamlandı.",
      loadFailed: isPersonal ? "Önemli tarihler yüklenemedi." : "Takip tarihleri yüklenemedi.",
      createFailed: isPersonal ? "Hatırlatıcı eklenemedi." : "Takip tarihi eklenemedi.",
      updateFailed: isPersonal ? "Hatırlatıcı güncellenemedi." : "Takip tarihi güncellenemedi.",
      deleteFailed: isPersonal ? "Hatırlatıcı silinemedi." : "Takip tarihi silinemedi.",
      completeFailed: isPersonal ? "Hatırlatıcı tamamlanamadı." : "Takip tarihi tamamlanamadı.",
      deleteConfirm: (title: string) =>
        isPersonal ? `${title} hatırlatıcısı silinsin mi?` : `${title} takip tarihi silinsin mi?`,
      dashboardUpcoming: isPersonal ? "Yaklaşan Tarihler" : "Yaklaşan Takip Tarihleri",
      dashboardOverdue: isPersonal ? "Gecikenler" : "Gecikenler",
      dashboardAssigned: isPersonal ? "Bana Atananlar" : "Bana Atananlar",
    }),
    [isPersonal],
  );
}

export type ImportantDateLabels = ReturnType<typeof useImportantDateLabels>;

export function getImportantDatesNavLabel() {
  return "Önemli Tarihler";
}
