import { createContext, useCallback, useContext, useState, type ReactNode } from "react";
import { AppModal } from "./AppModal";
import { Button } from "./Button";

export type ConfirmOptions = {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: "danger" | "primary";
};

type ConfirmContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
};

const ConfirmContext = createContext<ConfirmContextValue | null>(null);

type PendingConfirm = ConfirmOptions & {
  resolve: (value: boolean) => void;
};

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingConfirm | null>(null);

  const confirm = useCallback((options: ConfirmOptions) => {
    return new Promise<boolean>((resolve) => {
      setPending({ ...options, resolve });
    });
  }, []);

  const close = (result: boolean) => {
    pending?.resolve(result);
    setPending(null);
  };

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}
      {pending ? (
        <AppModal
          title={pending.title}
          onClose={() => close(false)}
          showCloseButton={false}
          footer={
            <>
              <Button variant="secondary" onClick={() => close(false)}>
                {pending.cancelLabel ?? "Vazgeç"}
              </Button>
              <Button variant={pending.variant === "danger" ? "danger" : "primary"} onClick={() => close(true)}>
                {pending.confirmLabel ?? "Onayla"}
              </Button>
            </>
          }
        >
          <p className="text-sm leading-6 text-slate-600">{pending.message}</p>
        </AppModal>
      ) : null}
    </ConfirmContext.Provider>
  );
}

export function useConfirm() {
  const context = useContext(ConfirmContext);
  if (!context) {
    throw new Error("useConfirm must be used within ConfirmProvider");
  }
  return context.confirm;
}
