import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { clearAccessToken, getAccessToken, setAccessToken } from "./apiClient";
import { login as loginRequest, me, register as registerRequest } from "./authService";
import type { AuthTenant, AuthUser, LoginInput, RegisterInput } from "../types/auth";

type AuthContextValue = {
  user: AuthUser | null;
  activeTenant: AuthTenant | null;
  tenants: AuthTenant[];
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (input: LoginInput) => Promise<void>;
  register: (input: RegisterInput) => Promise<void>;
  logout: () => void;
  refreshSession: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [activeTenant, setActiveTenant] = useState<AuthTenant | null>(null);
  const [tenants, setTenants] = useState<AuthTenant[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const resetSession = useCallback(() => {
    clearAccessToken();
    setUser(null);
    setActiveTenant(null);
    setTenants([]);
  }, []);

  useEffect(() => {
    let cancelled = false;

    const loadSession = async () => {
      if (!getAccessToken()) {
        setIsLoading(false);
        return;
      }

      try {
        const session = await me();
        if (cancelled) return;
        setUser(session.user);
        setActiveTenant(session.activeTenant);
        setTenants(session.tenants);
      } catch {
        if (!cancelled) resetSession();
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };

    void loadSession();

    const handleLogout = () => {
      resetSession();
    };
    window.addEventListener("auth:logout", handleLogout);

    return () => {
      cancelled = true;
      window.removeEventListener("auth:logout", handleLogout);
    };
  }, [resetSession]);

  const login = useCallback(async (input: LoginInput) => {
    const session = await loginRequest(input);
    setAccessToken(session.accessToken);
    setUser(session.user);
    setActiveTenant(session.activeTenant);
    setTenants(session.tenants);
  }, []);

  const register = useCallback(async (input: RegisterInput) => {
    const session = await registerRequest(input);
    setAccessToken(session.accessToken);
    setUser(session.user);
    setActiveTenant(session.activeTenant);
    setTenants(session.tenants);
  }, []);

  const logout = useCallback(() => {
    resetSession();
  }, [resetSession]);

  const refreshSession = useCallback(async () => {
    const session = await me();
    setUser(session.user);
    setActiveTenant(session.activeTenant);
    setTenants(session.tenants);
  }, []);

  const value = useMemo(
    () => ({
      user,
      activeTenant,
      tenants,
      isLoading,
      isAuthenticated: Boolean(user && activeTenant && getAccessToken()),
      login,
      register,
      logout,
      refreshSession,
    }),
    [activeTenant, isLoading, login, logout, refreshSession, register, tenants, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return value;
}
