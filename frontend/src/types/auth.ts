export type AccountType = "personal" | "organization";
export type TenantType = "personal" | "organization";
export type TenantRole = "owner" | "admin" | "member";

export type AuthUser = {
  id: string;
  fullName: string;
  email: string;
  createdAt: string;
};

export type AuthTenant = {
  id: string;
  name: string;
  type: TenantType;
  role: TenantRole;
};

export type AuthSession = {
  accessToken: string;
  user: AuthUser;
  activeTenant: AuthTenant;
  tenants: AuthTenant[];
};

export type MeResponse = {
  user: AuthUser;
  activeTenant: AuthTenant;
  tenants: AuthTenant[];
};

export type RegisterInput = {
  accountType: AccountType;
  fullName: string;
  email: string;
  password: string;
  organizationName?: string;
};

export type LoginInput = {
  email: string;
  password: string;
};

export type MessageResponse = {
  message: string;
};

export type ForgotPasswordInput = {
  email: string;
};

export type ResetPasswordInput = {
  token: string;
  newPassword: string;
};

export type UpdateProfileInput = {
  fullName: string;
};

export type ChangePasswordInput = {
  currentPassword: string;
  newPassword: string;
};

export type TenantProfile = {
  tenantId: string;
  tenantName: string;
  tenantType: TenantType;
  userRole: TenantRole;
  memberCount: number;
  createdAt: string;
};
