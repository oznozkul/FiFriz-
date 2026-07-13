import { apiRequest } from "./apiClient";

export type TenantMember = {
  userId: string;
  fullName: string;
  email: string;
  role: "owner" | "admin" | "member";
};

export function listTenantMembers() {
  return apiRequest<TenantMember[]>("/api/v1/tenant/members");
}
