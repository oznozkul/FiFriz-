import type { IntegrationStatus } from "../../services/integrationService";
import { integrationStatusLabel } from "../../services/integrationService";

const statusClasses: Record<string, string> = {
  connected: "bg-emerald-50 text-emerald-700 border-emerald-200",
  disconnected: "bg-slate-100 text-slate-600 border-slate-200",
  sync_pending: "bg-amber-50 text-amber-700 border-amber-200",
  syncing: "bg-sky-50 text-sky-700 border-sky-200",
  sync_failed: "bg-red-50 text-red-700 border-red-200",
  sync_completed: "bg-emerald-50 text-emerald-700 border-emerald-200",
};

type IntegrationStatusBadgeProps = {
  status: IntegrationStatus | string;
  label?: string;
};

export function IntegrationStatusBadge({ status, label }: IntegrationStatusBadgeProps) {
  const classes = statusClasses[status] ?? statusClasses.disconnected;
  const text = label ?? integrationStatusLabel(status as IntegrationStatus);
  return <span className={`inline-flex rounded-full border px-2.5 py-0.5 text-xs font-semibold ${classes}`}>{text}</span>;
}
