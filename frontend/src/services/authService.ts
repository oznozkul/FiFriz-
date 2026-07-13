import { apiPost, apiRequest } from "./apiClient";
import type { AuthSession, ChangePasswordInput, ForgotPasswordInput, LoginInput, MeResponse, MessageResponse, RegisterInput, ResetPasswordInput, UpdateProfileInput } from "../types/auth";

export function register(input: RegisterInput) {
  return apiRequest<AuthSession>(
    "/api/v1/auth/register",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(input),
    },
    "Kayıt tamamlanamadı.",
  );
}

export function login(input: LoginInput) {
  return apiRequest<AuthSession>(
    "/api/v1/auth/login",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(input),
    },
    "Giriş yapılamadı.",
  );
}

export function me() {
  return apiRequest<MeResponse>("/api/v1/auth/me", undefined, "Oturum bilgisi alınamadı.");
}

export function forgotPassword(input: ForgotPasswordInput) {
  return apiPost<MessageResponse>("/api/v1/auth/forgot-password", input, "Şifre sıfırlama isteği gönderilemedi.");
}

export function resetPassword(input: ResetPasswordInput) {
  return apiPost<MessageResponse>("/api/v1/auth/reset-password", input, "Şifre sıfırlanamadı.");
}

export function updateProfile(input: UpdateProfileInput) {
  return apiRequest<MeResponse>(
    "/api/v1/profile",
    {
      method: "PUT",
      body: JSON.stringify(input),
    },
    "Profil güncellenemedi.",
  );
}

export function changePassword(input: ChangePasswordInput) {
  return apiRequest<MessageResponse>(
    "/api/v1/profile/change-password",
    {
      method: "PATCH",
      body: JSON.stringify(input),
    },
    "Şifre değiştirilemedi.",
  );
}
