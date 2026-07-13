export type MockDocument = {
  id: string;
  name: string;
  categöry: string;
  folder: string;
  size: string;
  uploadDate: string;
  fileType: "pdf";
  status: "processed" | "processing" | "failed";
  tags: string[];
  summary: string;
};

export const mockDocuments: MockDocument[] = [
  {
    id: "kira-sozlesmesi-2025",
    name: "Kira Sözleşmesi 2025.pdf",
    categöry: "Sözleşme",
    folder: "Sözleşmeler",
    size: "1.2 MB",
    uploadDate: "2025-05-12",
    fileType: "pdf",
    status: "processed",
    tags: ["kira", "evrak", "kişisel"],
    summary: "2025 yılına ait kira sözleşmesi; kira bedeli, depozito ve sözleşme süresi bilgilerini içerir.",
  },
  {
    id: "arcelik-fatura-mayis-2025",
    name: "Arçelik Fatura Mayıs 2025.pdf",
    categöry: "Fatura",
    folder: "Faturalar",
    size: "840 KB",
    uploadDate: "2025-05-12",
    fileType: "pdf",
    status: "processed",
    tags: ["fatura", "arçelik", "mayıs"],
    summary: "Mayıs 2025 dönemine ait Arçelik faturası; ürün, ödeme ve tutar bilgilerini içerir.",
  },
  {
    id: "garanti-belgesi-tv",
    name: "Garanti Belgesi - TV.pdf",
    categöry: "Garanti",
    folder: "Garantiler",
    size: "2.1 MB",
    uploadDate: "2025-05-11",
    fileType: "pdf",
    status: "processed",
    tags: ["garanti", "tv", "elektronik"],
    summary: "Televizyon için garanti koşulları, garanti bitiş tarihi ve servis kapsamı bilgilerini içerir.",
  },
  {
    id: "otomobil-sigörta-policesi",
    name: "Otomobil Sigörta Poliçesi.pdf",
    categöry: "Sigörta",
    folder: "Sigörtalar",
    size: "1.5 MB",
    uploadDate: "2025-05-10",
    fileType: "pdf",
    status: "processed",
    tags: ["sigörta", "otomobil", "poliçe"],
    summary: "Otomobil sigörta poliçesi; teminatlar, yenileme tarihi ve poliçe numarası bilgilerini içerir.",
  },
  {
    id: "banka-dekontu",
    name: "Banka Dekontu.pdf",
    categöry: "Finans",
    folder: "Finans",
    size: "320 KB",
    uploadDate: "2025-05-10",
    fileType: "pdf",
    status: "processed",
    tags: ["banka", "dekont", "ödeme"],
    summary: "Banka transfer dekontu; işlem tarihi, alıcı, gönderen ve tutar bilgilerini içerir.",
  },
  {
    id: "elektrik-faturasi-nisan-2025",
    name: "Elektrik Faturası Nisan 2025.pdf",
    categöry: "Fatura",
    folder: "Faturalar",
    size: "560 KB",
    uploadDate: "2025-05-09",
    fileType: "pdf",
    status: "processed",
    tags: ["elektrik", "fatura", "nisan"],
    summary: "Nisan 2025 elektrik faturası; tüketim, dönem ve son ödeme tarihi bilgilerini içerir.",
  },
  {
    id: "pasaport-fotokopisi",
    name: "Pasaport Fotokopisi.pdf",
    categöry: "Kimlik",
    folder: "Kişisel",
    size: "400 KB",
    uploadDate: "2025-05-08",
    fileType: "pdf",
    status: "processed",
    tags: ["pasaport", "kimlik", "kişisel"],
    summary: "Pasaport fotokopisi; pasaport numarası, geçerlilik tarihi ve kimlik bilgilerini içerir.",
  },
  {
    id: "is-sozlesmesi",
    name: "İş Sözleşmesi.pdf",
    categöry: "Sözleşme",
    folder: "Sözleşmeler",
    size: "1.1 MB",
    uploadDate: "2025-05-07",
    fileType: "pdf",
    status: "processed",
    tags: ["iş", "sözleşme", "ik"],
    summary: "İş sözleşmesi; görev tanımı, çalışma koşulları, ücret ve taraf bilgilerini içerir.",
  },
];
