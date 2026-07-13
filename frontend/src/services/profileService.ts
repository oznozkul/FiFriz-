import { apiGet } from "./apiClient";
import type { TenantProfile } from "../types/auth";

export function getProfileTenant() {
  return apiGet<TenantProfile>("/api/v1/profile/tenant", "Tenant bilgisi alınamadı.");
}
