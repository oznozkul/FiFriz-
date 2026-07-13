import type { ButtonHTMLAttributes, ReactNode } from "react";
import { Button } from "./Button";

type PrimaryButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
};

export function PrimaryButton({ children, className = "", type = "button", ...props }: PrimaryButtonProps) {
  return (
    <Button className={className} type={type} variant="primary" {...props}>
      {children}
    </Button>
  );
}
