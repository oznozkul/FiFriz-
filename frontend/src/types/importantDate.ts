export type ImportantDateType = "expiration" | "renewal" | "payment_due" | "review" | "custom";

export type ImportantDate = {
  id: string;
  documentId: string;
  title: string;
  date: string;
  reminderDaysBefore?: number | null;
  note?: string | null;
  type?: ImportantDateType | null;
  isCompleted: boolean;
  completedAt?: string | null;
  createdByUserId: string;
  createdByUserName: string;
  assignedToUserId?: string | null;
  assignedToUserName?: string | null;
  documentName?: string;
  source?: "manual" | "ai";
  status?: "suggested" | "active" | "rejected";
  confidence?: number | null;
  sourceText?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type ImportantDateInput = {
  title: string;
  date: string;
  reminderDaysBefore?: number | null;
  note?: string | null;
  type?: ImportantDateType | null;
  assignedToUserId?: string | null;
};

export type ImportantDateListParams = {
  upcomingOnly?: boolean;
  completed?: boolean;
  assignedToMe?: boolean;
  from?: string;
  to?: string;
  type?: ImportantDateType;
};

export const IMPORTANT_DATE_TYPE_OPTIONS: Array<{ value: ImportantDateType; label: string }> = [
  { value: "expiration", label: "Son Geçerlilik" },
  { value: "renewal", label: "Yenileme" },
  { value: "payment_due", label: "Son Ödeme" },
  { value: "review", label: "İnceleme" },
  { value: "custom", label: "Özel" },
];

export function importantDateTypeLabel(type?: ImportantDateType | null) {
  return IMPORTANT_DATE_TYPE_OPTIONS.find((item) => item.value === type)?.label ?? "Özel";
}
