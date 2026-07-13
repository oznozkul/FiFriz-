import type { ButtonHTMLAttributes, ReactNode } from "react";

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
};

export function IconButton({ children, className = "", type = "button", ...props }: IconButtonProps) {
  return (
    <button
      className={`inline-flex h-9 w-9 items-center justify-center rounded-lg border border-line bg-white text-slate-500 transition hover:border-violet-ring hover:bg-violet-soft hover:text-violet-brand active:scale-[0.98] focus:outline-none focus:ring-4 focus:ring-violet-ring/70 disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
      type={type}
      {...props}
    >
      {children}
    </button>
  );
}
