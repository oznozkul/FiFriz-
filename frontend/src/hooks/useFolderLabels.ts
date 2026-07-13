import { useMemo } from "react";
import { useAuth } from "../services/AuthContext";

export function useFolderLabels() {
  const { activeTenant } = useAuth();
  const isPersonal = activeTenant?.type === "personal";

  return useMemo(() => {
    const singular = isPersonal ? "Kategori" : "Klasör";
    const plural = isPersonal ? "Kategoriler" : "Klasörler";
    const pluralLower = isPersonal ? "kategoriler" : "klasörler";
    const singularLower = isPersonal ? "kategori" : "klasör";

    return {
      isPersonal,
      navLabel: plural,
      pageTitle: isPersonal ? "Kategorilerim" : "Klasörler",
      pageSubtitle: isPersonal
        ? "Belgelerinizi kategorilerle düzenleyin."
        : "Belgelerinizi klasör yapısıyla organize edin.",
      newLabel: isPersonal ? "Yeni Kategori" : "Yeni Klasör",
      createTitle: isPersonal ? "Yeni Kategori" : "Yeni Klasör",
      renameTitle: isPersonal ? "Kategoriyi Düzenle" : "Klasörü Düzenle",
      nameLabel: isPersonal ? "Kategori adı" : "Klasör adı",
      selectLabel: isPersonal ? "Kategori" : "Klasör",
      noneLabel: isPersonal ? "Kategorisiz" : "Klasörsüz",
      rootLabel: isPersonal ? "Ana Kategori" : "Ana Klasör",
      locationLabel: "Konum",
      emptyTreeTitle: isPersonal ? "Henüz kategori yok." : "Henüz klasör yok.",
      emptyTreeDescription: isPersonal ? "İlk kategorini oluştur." : "İlk klasörünü oluştur.",
      loadingLabel: isPersonal ? "Kategoriler yükleniyor..." : "Klasörler yükleniyor...",
      refreshToast: isPersonal ? "Kategoriler yenilendi." : "Klasörler yenilendi.",
      createdToast: isPersonal ? "Kategori oluşturuldu." : "Klasör oluşturuldu.",
      updatedToast: isPersonal ? "Kategori güncellendi." : "Klasör güncellendi.",
      deletedToast: isPersonal ? "Kategori silindi." : "Klasör silindi.",
      deleteConfirm: (name: string) =>
        isPersonal ? `${name} kategorisi silinsin mi?` : `${name} klasörü silinsin mi?`,
      loadFailed: isPersonal ? "Kategoriler yüklenemedi." : "Klasörler yüklenemedi.",
      contentLoadFailed: isPersonal ? "Kategori içeriği yüklenemedi." : "Klasör içeriği yüklenemedi.",
      createFailed: isPersonal ? "Kategori oluşturulamadı." : "Klasör oluşturulamadı.",
      updateFailed: isPersonal ? "Kategori güncellenemedi." : "Klasör güncellenemedi.",
      deleteFailed: isPersonal ? "Kategori silinemedi." : "Klasör silinemedi.",
      cannotDeleteSystem: isPersonal ? "Sistem kategorileri silinemez." : "Sistem klasörleri silinemez.",
      newChildLabel: isPersonal ? "Yeni alt kategori" : "Yeni alt klasör",
      newSiblingLabel: isPersonal ? "Yeni kategori" : "Yeni klasör",
      addRootAria: isPersonal ? "Üst kategori ekle" : "Üst klasör ekle",
      addChildAria: isPersonal ? "Alt kategori ekle" : "Alt klasör ekle",
      renameAria: isPersonal ? "Kategoriyi yeniden adlandır" : "Klasörü yeniden adlandır",
      deleteAria: isPersonal ? "Kategoriyi sil" : "Klasörü sil",
      expandAria: isPersonal ? "Kategoriyi genişlet" : "Klasörü genişlet",
      collapseAria: isPersonal ? "Kategoriyi daralt" : "Klasörü daralt",
      emptyCurrentTitle: isPersonal ? "Bu kategori boş" : "Bu klasör boş",
      emptyCurrentDescription: isPersonal
        ? "Alt kategori ekleyebilir veya belge yüklerken bu kategoriyi seçebilirsiniz."
        : "Alt klasör ekleyebilir veya belge yüklerken bu klasörü seçebilirsiniz.",
      emptyRootTitle: isPersonal ? "Henüz kategori yok" : "Henüz klasör yok",
      emptyRootDescription: isPersonal
        ? "İlk kategorinizi oluşturun veya belge yükleyin."
        : "İlk klasörünüzü oluşturun veya belge yükleyin.",
      itemTypeLabel: isPersonal ? "Kategori" : "Klasör",
      treeTitle: isPersonal ? "Kategoriler" : "Klasör ağacı",
      breadcrumbEmptyLabel: isPersonal ? "Kategori seçilmedi" : "Klasör seçilmedi",
      breadcrumbAriaLabel: isPersonal ? "Kategori yolu" : "Klasör yolu",
      searchPlaceholder: isPersonal ? "Kategori veya belge ara..." : "Klasör veya belge ara...",
      filterAll: "Tümü",
      filterFoldersOnly: isPersonal ? "Yalnızca kategoriler" : "Yalnızca klasörler",
      filterDocumentsOnly: "Yalnızca belgeler",
      metadataTitle: isPersonal ? "Kategori özeti" : "Klasör özeti",
      createPlaceholder: isPersonal ? "Örn. Faturalar" : "Örn. Sözleşmeler",
      createParentHint: (parentName: string) =>
        isPersonal ? `${parentName} altına eklenecek.` : `${parentName} altına oluşturulacak.`,
      createRootHint: isPersonal ? "Üst seviyede yeni kategori." : "Üst seviyede yeni klasör.",
      dragDropHint: isPersonal ? "Taşımak için bırakın" : "Taşımak için buraya bırakın",
      loadFoldersErrorTitle: isPersonal ? "Kategoriler yüklenemedi" : "Klasörler yüklenemedi",
      moveSuccessToast: isPersonal ? "Kategori taşındı." : "Klasör taşındı.",
      moveFailedToast: isPersonal ? "Kategori taşınamadı." : "Klasör taşınamadı.",
      lockedDeleteError: isPersonal ? "Kilitli kategori silinemez." : "Kilitli klasör silinemez.",
      deleteConfirmTitle: isPersonal ? "Kategoriyi sil" : "Klasörü sil",
      pinAddedToast: isPersonal ? "Kategori sabitlendi." : "Klasör sabitlendi.",
      pinRemovedToast: "Sabitleme kaldırıldı.",
      colorUpdatedToast: isPersonal ? "Kategori rengi güncellendi." : "Klasör rengi güncellendi.",
      iconUpdatedToast: isPersonal ? "Kategori ikonu güncellendi." : "Klasör ikonu güncellendi.",
      lockSuccessToast: isPersonal ? "Kategori kilitlendi." : "Klasör kilitlendi.",
      unlockSuccessToast: isPersonal ? "Kategori kilidi açıldı." : "Klasör kilidi açıldı.",
      columnName: "Ad",
      columnDate: "Tarih",
      columnType: "Tür",
      columnSize: "Boyut",
      footerSummary: (folderCount: number, docCount: number) =>
        `${folderCount} ${singularLower}, ${docCount} belge`,
      selectLoadFailed: isPersonal ? "Kategori ve etiket listesi yüklenemedi." : "Klasör ve etiket listesi yüklenemedi.",
      uploadSubtitle: isPersonal
        ? "Dosyanızı seçin ve kategori bilgisini ekleyin."
        : "Dosyanızı seçin ve klasör bilgisini ekleyin.",
      singular,
      plural,
      singularLower,
      pluralLower,
    };
  }, [isPersonal]);
}

export type FolderLabels = ReturnType<typeof useFolderLabels>;

export function getFolderNavLabel(tenantType?: string | null) {
  return tenantType === "personal" ? "Kategoriler" : "Klasörler";
}
