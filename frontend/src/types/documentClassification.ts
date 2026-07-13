export type DocumentClassification = {
  documentId: string;
  documentType: string;
  confidence?: number | null;
  reason?: string | null;
  classifiedAt?: string | null;
};

export const AI_DOCUMENT_TYPE_LABELS: Record<string, string> = {
  invoice: "Fatura",
  contract: "Sözleşme",
  court_decision: "Mahkeme Kararı",
  vehicle_registration: "Araç Ruhsatı",
  identity_document: "Kimlik Belgesi",
  passport: "Pasaport",
  insurance_policy: "Sigorta Poliçesi",
  warranty_document: "Garanti Belgesi",
  bank_receipt: "Banka Dekontu",
  academic_article: "Akademik Makale",
  hr_document: "İK Belgesi",
  medical_document: "Tıbbi Belge",
  tax_document: "Vergi Belgesi",
  certificate: "Sertifika",
  other: "Diğer",
};

export function aiDocumentTypeLabel(value?: string | null) {
  if (!value) return null;
  return AI_DOCUMENT_TYPE_LABELS[value] ?? value;
}
