import type { ButtonHTMLAttributes, ReactNode } from "react";

export type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  variant?: ButtonVariant;
};

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-violet-brand text-white shadow-lg shadow-violet-200/80 hover:bg-violet-hover focus:ring-violet-ring/80",
  secondary:
    "border border-line bg-white text-slate-600 hover:border-violet-ring hover:bg-violet-soft/40 hover:text-violet-brand focus:ring-violet-ring/60",
  danger: "bg-red-600 text-white shadow-lg shadow-red-100/80 hover:bg-red-700 focus:ring-red-300/80",
  ghost: "text-slate-600 hover:bg-slate-50 hover:text-ink focus:ring-violet-ring/50",
};

export function Button({ children, className = "", variant = "primary", type = "button", ...props }: ButtonProps) {
  return (
    <button
      className={`inline-flex h-10 items-center justify-center gap-2 rounded-lg px-4 text-sm font-semibold transition active:scale-[0.98] focus:outline-none focus:ring-4 disabled:cursor-not-allowed disabled:opacity-60 ${variantClasses[variant]} ${className}`}
      type={type}
      {...props}
    >
      {children}
    </button>
  );
}
