import type { ApiDocument } from "../../types/document";
import { Badge, type BadgeProps } from "./Badge";

export type StatusBadgeKind =
  | ApiDocument["status"]
  | "pending"
  | "approved"
  | "rejected"
  | "active"
  | "expired"
  | "none"
  | string;

type StatusConfig = {
  label: string;
  tone: NonNullable<BadgeProps["tone"]>;
};

const STATUS_CONFIG: Record<string, StatusConfig> = {
  uploading: { label: "Yükleniyor", tone: "slate" },
  processing: { label: "İşleniyor", tone: "amber" },
  ready: { label: "Hazır", tone: "green" },
  failed: { label: "Başarısız", tone: "rose" },
  archived: { label: "Arşivlendi", tone: "slate" },
  deleted: { label: "Silindi", tone: "slate" },
  locked: { label: "Kilitli", tone: "violet" },
  expired: { label: "Süresi doldu", tone: "slate" },
  retention_locked: { label: "Saklama kilidi", tone: "violet" },
  uploaded: { label: "Yükleniyor", tone: "slate" },
  processed: { label: "Hazır", tone: "green" },
  pending: { label: "Bekliyor", tone: "amber" },
  approved: { label: "Onaylandı", tone: "green" },
  rejected: { label: "Reddedildi", tone: "rose" },
  active: { label: "Aktif", tone: "green" },
  hidden: { label: "Gizli", tone: "slate" },
  none: { label: "Yok", tone: "slate" },
};

export function resolveStatusBadge(status: StatusBadgeKind): StatusConfig {
  return STATUS_CONFIG[status] ?? { label: status, tone: "slate" };
}

export function statusLabel(status: ApiDocument["status"]) {
  return resolveStatusBadge(status).label;
}

export function statusBadgeTone(status: ApiDocument["status"]): NonNullable<BadgeProps["tone"]> {
  return resolveStatusBadge(status).tone;
}

type StatusBadgeProps = {
  status: StatusBadgeKind;
  label?: string;
  className?: string;
};

export function StatusBadge({ status, label, className }: StatusBadgeProps) {
  const config = resolveStatusBadge(status);
  return (
    <Badge tone={config.tone} className={className}>
      {label ?? config.label}
    </Badge>
  );
}
