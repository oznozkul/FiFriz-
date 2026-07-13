import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { PrimaryButton, Spinner } from "../components/common";
import { forgotPassword } from "../services/authService";
import { toast } from "../services/toast";
import { AuthShell, AuthInput } from "./LoginPage";

export function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setIsSubmitting(true);

    try {
      await forgotPassword({ email });
      setIsSubmitted(true);
      toast.success("Eğer bu email kayıtlıysa şifre sıfırlama bağlantısı gönderildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "İstek gönderilemedi.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AuthShell title="Şifremi unuttum" subtitle="Kayıtlı e-posta adresinize sıfırlama bağlantısı gönderilir.">
      {isSubmitted ? (
        <div className="space-y-4 text-center">
          <p className="text-sm text-slate-600">Eğer bu email kayıtlıysa şifre sıfırlama bağlantısı gönderildi.</p>
          <Link className="inline-block text-sm font-bold text-violet-brand hover:text-violet-hover" to="/login">
            Giriş sayfasına dön
          </Link>
        </div>
      ) : (
        <form className="space-y-4" onSubmit={handleSubmit}>
          <AuthInput label="E-posta" value={email} type="email" autoComplete="email" placeholder="örnek@fifriz.com" onChange={setEmail} />
          <PrimaryButton type="submit" className="h-11 w-full justify-center" disabled={isSubmitting}>
            {isSubmitting ? <Spinner size="sm" label="Gönderiliyor..." className="text-white" /> : "Sıfırlama bağlantısı gönder"}
          </PrimaryButton>
          <p className="text-center text-sm text-slate-500">
            <Link className="font-bold text-violet-brand hover:text-violet-hover" to="/login">
              Giriş sayfasına dön
            </Link>
          </p>
        </form>
      )}
    </AuthShell>
  );
}
