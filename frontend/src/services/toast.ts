export type ToastType = "success" | "error" | "info" | "loading";

export type ToastItem = {
  id: string;
  type: ToastType;
  message: string;
  duration: number;
};

type ToastListener = (items: ToastItem[]) => void;

const listeners = new Set<ToastListener>();
let items: ToastItem[] = [];

function emit() {
  const snapshot = [...items];
  listeners.forEach((listener) => listener(snapshot));
}

function addToast(type: ToastType, message: string, duration: number) {
  const id = crypto.randomUUID();
  items = [...items, { id, type, message, duration }];
  emit();

  if (duration > 0) {
    window.setTimeout(() => dismiss(id), duration);
  }

  return id;
}

export function dismiss(id: string) {
  items = items.filter((item) => item.id !== id);
  emit();
}

export function subscribe(listener: ToastListener) {
  listeners.add(listener);
  listener([...items]);
  return () => {
    listeners.delete(listener);
  };
}

export const toast = {
  success(message: string, duration = 4000) {
    return addToast("success", message, duration);
  },
  error(message: string, duration = 6000) {
    return addToast("error", message, duration);
  },
  info(message: string, duration = 4500) {
    return addToast("info", message, duration);
  },
  loading(message: string) {
    return addToast("loading", message, 0);
  },
  dismiss,
  async promise<T>(
    promise: Promise<T>,
    messages: {
      loading: string;
      success: string;
      error: string;
    },
  ): Promise<T> {
    const loadingId = toast.loading(messages.loading);
    try {
      const result = await promise;
      dismiss(loadingId);
      toast.success(messages.success);
      return result;
    } catch (error) {
      dismiss(loadingId);
      toast.error(error instanceof Error ? error.message : messages.error);
      throw error;
    }
  },
};
