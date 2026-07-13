import { FormEvent, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { PrimaryButton, Spinner } from "../components/common";
import { resetPassword } from "../services/authService";
import { toast } from "../services/toast";
import { AuthShell, AuthInput } from "./LoginPage";

const minPasswordLength = 8;

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = useMemo(() => searchParams.get("token")?.trim() ?? "", [searchParams]);
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!token) {
    return (
      <AuthShell title="Geçersiz bağlantı" subtitle="Şifre sıfırlama bağlantısı eksik veya hatalı.">
        <p className="text-center text-sm text-slate-600">
          <Link className="font-bold text-violet-brand hover:text-violet-hover" to="/forgot-password">
            Yeni sıfırlama bağlantısı iste
          </Link>
        </p>
      </AuthShell>
    );
  }

  const passwordsMatch = newPassword === confirmPassword;
  const canSubmit = newPassword.length >= minPasswordLength && passwordsMatch;

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!canSubmit) {
      toast.error("Şifreler eşleşmeli ve en az 8 karakter olmalıdır.");
      return;
    }

    setIsSubmitting(true);
    try {
      await resetPassword({ token, newPassword });
      toast.success("Şifreniz güncellendi. Giriş yapabilirsiniz.");
      navigate("/login", { replace: true });
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Şifre sıfırlanamadı.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AuthShell title="Yeni şifre belirle" subtitle="Hesabınız için yeni bir şifre oluşturun.">
      <form className="space-y-4" onSubmit={handleSubmit}>
        <AuthInput label="Yeni şifre" value={newPassword} type="password" autoComplete="new-password" placeholder="En az 8 karakter" onChange={setNewPassword} />
        <AuthInput label="Şifre tekrar" value={confirmPassword} type="password" autoComplete="new-password" placeholder="Şifrenizi tekrar girin" onChange={setConfirmPassword} />
        {confirmPassword && !passwordsMatch ? <p className="text-sm text-rose-600">Şifreler eşleşmiyor.</p> : null}
        <PrimaryButton type="submit" className="h-11 w-full justify-center" disabled={isSubmitting || !canSubmit}>
          {isSubmitting ? <Spinner size="sm" label="Kaydediliyor..." className="text-white" /> : "Şifreyi sıfırla"}
        </PrimaryButton>
        <p className="text-center text-sm text-slate-500">
          <Link className="font-bold text-violet-brand hover:text-violet-hover" to="/login">
            Giriş sayfasına dön
          </Link>
        </p>
      </form>
    </AuthShell>
  );
}
