import { FormEvent, useState } from "react";
import type { ReactNode } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { PrimaryButton, Spinner, BrandMark } from "../components/common";
import { useAuth } from "../services/AuthContext";
import { toast } from "../services/toast";

export function LoginPage() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/dashboard";

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setIsSubmitting(true);

    try {
      await login({ email, password });
      toast.success("Giriş başarılı.");
      navigate(from, { replace: true });
    } catch (loginError) {
      toast.error(loginError instanceof Error ? loginError.message : "Giriş yapılamadı.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AuthShell title="Giriş yap" subtitle="Belgelerine ve tenant alanına devam et.">
      <form className="space-y-4" onSubmit={handleSubmit}>
        <AuthInput label="E-posta" value={email} type="email" autoComplete="email" placeholder="örnek@fifriz.com" onChange={setEmail} />
        <AuthInput label="Şifre" value={password} type="password" autoComplete="current-password" placeholder="Şifrenizi girin" onChange={setPassword} />
        <div className="text-right">
          <Link className="text-sm font-bold text-violet-brand hover:text-violet-hover" to="/forgot-password">
            Şifremi unuttum
          </Link>
        </div>
        <PrimaryButton type="submit" className="h-11 w-full justify-center" disabled={isSubmitting}>
          {isSubmitting ? <Spinner size="sm" label="Giriş yapılıyor..." className="text-white" /> : "Giriş yap"}
        </PrimaryButton>
        <p className="text-center text-sm text-slate-500">
          Hesabın yok mu?{" "}
          <Link className="font-bold text-violet-brand hover:text-violet-hover" to="/register">
            Kayıt ol
          </Link>
        </p>
      </form>
    </AuthShell>
  );
}

export function RegisterPage() {
  const { register, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [accountType, setAccountType] = useState<"personal" | "organization">("personal");
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [organizationName, setOrganizationName] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setIsSubmitting(true);

    try {
      await register({
        accountType,
        fullName,
        email,
        password,
        organizationName: accountType === "organization" ? organizationName : undefined,
      });
      toast.success("Hesabınız oluşturuldu.");
      navigate("/dashboard", { replace: true });
    } catch (registerError) {
      toast.error(registerError instanceof Error ? registerError.message : "Kayıt tamamlanamadı.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AuthShell title="Kayıt ol" subtitle="Bireysel veya kurumsal tenant alanını oluştur.">
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div className="grid grid-cols-2 gap-2 rounded-lg border border-line bg-slate-50 p-1">
          <button className={`h-10 rounded-md text-sm font-bold transition ${accountType === "personal" ? "bg-white text-violet-brand shadow-sm" : "text-slate-500 hover:text-slate-700"}`} type="button" onClick={() => setAccountType("personal")}>
            Bireysel
          </button>
          <button className={`h-10 rounded-md text-sm font-bold transition ${accountType === "organization" ? "bg-white text-violet-brand shadow-sm" : "text-slate-500 hover:text-slate-700"}`} type="button" onClick={() => setAccountType("organization")}>
            Kurumsal
          </button>
        </div>
        <AuthInput label="Ad Soyad" value={fullName} autoComplete="name" placeholder="Ali Ozan Ozkul" onChange={setFullName} />
        <AuthInput label="E-posta" value={email} type="email" autoComplete="email" placeholder="örnek@fifriz.com" onChange={setEmail} />
        <AuthInput label="Şifre" value={password} type="password" autoComplete="new-password" placeholder="En az 6 karakter" onChange={setPassword} />
        {accountType === "organization" && <AuthInput label="Organizasyon adı" value={organizationName} autoComplete="organization" placeholder="Şirket veya organizasyon adı" onChange={setOrganizationName} />}
        <PrimaryButton type="submit" className="h-11 w-full justify-center" disabled={isSubmitting}>
          {isSubmitting ? <Spinner size="sm" label="Kayıt oluşturuluyor..." className="text-white" /> : "Kayıt ol"}
        </PrimaryButton>
        <p className="text-center text-sm text-slate-500">
          Zaten hesabın var mı?{" "}
          <Link className="font-bold text-violet-brand hover:text-violet-hover" to="/login">
            Giriş yap
          </Link>
        </p>
      </form>
    </AuthShell>
  );
}

export function AuthShell({ title, subtitle, children }: { title: string; subtitle: string; children: ReactNode }) {
  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4 py-10">
      <section className="w-full max-w-md rounded-xl border border-line bg-white p-6 shadow-panel">
        <div className="mb-6 text-center">
          <BrandMark size="md" className="mx-auto mb-4 shadow-lg shadow-slate-200" />
          <h1 className="text-2xl font-black text-ink">{title}</h1>
          <p className="mt-2 text-sm text-slate-500">{subtitle}</p>
        </div>
        {children}
      </section>
    </main>
  );
}

export function AuthInput({
  label,
  value,
  type = "text",
  autoComplete,
  placeholder,
  onChange,
}: {
  label: string;
  value: string;
  type?: string;
  autoComplete?: string;
  placeholder?: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className="block space-y-2">
      <span className="text-sm font-bold text-slate-600">{label}</span>
      <input className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm outline-none transition placeholder:text-slate-400 focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/60" value={value} type={type} autoComplete={autoComplete} placeholder={placeholder} onChange={(event) => onChange(event.target.value)} required />
    </label>
  );
}
