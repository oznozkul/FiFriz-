export type DocumentEntity = {
  id: string;
  documentId: string;
  entityType: string;
  value: string;
  label?: string | null;
  confidence?: number | null;
  sourceText?: string | null;
  pageNumber?: number | null;
  extractedBy: string;
  createdAt: string;
  updatedAt: string;
};

export const ENTITY_TYPE_LABELS: Record<string, string> = {
  person: "Kişi",
  company: "Şirket",
  organization: "Kurum",
  email: "E-posta",
  phone: "Telefon",
  address: "Adres",
  tax_number: "Vergi No",
  identity_number: "Kimlik No",
  passport_number: "Pasaport No",
  license_plate: "Plaka",
  vehicle_vin: "Şasi No",
  invoice_number: "Fatura No",
  contract_number: "Sözleşme No",
  court_case_number: "Dava No",
  date: "Tarih",
  amount: "Tutar",
  currency: "Para Birimi",
  product: "Ürün",
  location: "Konum",
  custom: "Özel",
};

export type EntityGroupKey =
  | "people"
  | "companies"
  | "dates"
  | "amounts"
  | "identifiers"
  | "addresses"
  | "other";

export const ENTITY_GROUP_LABELS: Record<EntityGroupKey, string> = {
  people: "Kişiler",
  companies: "Şirketler",
  dates: "Tarihler",
  amounts: "Tutarlar",
  identifiers: "Kimlik / Vergi / Referans Numaraları",
  addresses: "Adresler",
  other: "Diğer",
};

const ENTITY_GROUP_TYPES: Record<EntityGroupKey, string[]> = {
  people: ["person"],
  companies: ["company", "organization"],
  dates: ["date"],
  amounts: ["amount", "currency"],
  identifiers: [
    "tax_number",
    "identity_number",
    "passport_number",
    "license_plate",
    "vehicle_vin",
    "invoice_number",
    "contract_number",
    "court_case_number",
    "email",
    "phone",
  ],
  addresses: ["address", "location"],
  other: ["product", "custom"],
};

export function entityGroupForType(entityType: string): EntityGroupKey {
  for (const [group, types] of Object.entries(ENTITY_GROUP_TYPES) as [EntityGroupKey, string[]][]) {
    if (types.includes(entityType)) {
      return group;
    }
  }
  return "other";
}

export function groupEntities(entities: DocumentEntity[]) {
  const groups: Record<EntityGroupKey, DocumentEntity[]> = {
    people: [],
    companies: [],
    dates: [],
    amounts: [],
    identifiers: [],
    addresses: [],
    other: [],
  };

  for (const entity of entities) {
    groups[entityGroupForType(entity.entityType)].push(entity);
  }

  return (Object.keys(ENTITY_GROUP_LABELS) as EntityGroupKey[])
    .map((key) => ({ key, label: ENTITY_GROUP_LABELS[key], items: groups[key] }))
    .filter((group) => group.items.length > 0);
}
