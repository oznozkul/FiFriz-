export type MetadataParty = {
  name: string;
  role?: string;
};

export type MetadataDate = {
  label: string;
  value: string;
};

export type MetadataAmount = {
  label: string;
  value: number;
  currency?: string;
};

export type MetadataIdentifier = {
  label: string;
  value: string;
};

export type ExtractedDocumentMetadata = {
  documentType?: string;
  title?: string;
  parties: MetadataParty[];
  dates: MetadataDate[];
  amounts: MetadataAmount[];
  identifiers: MetadataIdentifier[];
  addresses: string[];
  keywords: string[];
  summary?: string;
  confidence?: number | null;
};

export type DocumentAIMetadata = {
  documentId: string;
  documentType?: string | null;
  metadata: ExtractedDocumentMetadata;
  confidence?: number | null;
  model?: string | null;
  extractedBy: string;
  createdAt: string;
  updatedAt: string;
  rawJson?: unknown;
};

export const DOCUMENT_TYPE_LABELS: Record<string, string> = {
  contract: "Sözleşme",
  invoice: "Fatura",
  receipt: "Makbuz",
  identity: "Kimlik Belgesi",
  policy: "Poliçe",
  agreement: "Anlaşma",
  letter: "Yazı",
  report: "Rapor",
  certificate: "Sertifika",
  other: "Diğer",
};
