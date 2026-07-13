import { RefreshCw } from "lucide-react";
import { Button, type ButtonProps } from "./Button";

type RetryButtonProps = Omit<ButtonProps, "children"> & {
  label?: string;
};

export function RetryButton({ label = "Tekrar dene", className = "", ...props }: RetryButtonProps) {
  return (
    <Button variant="secondary" className={className} {...props}>
      <RefreshCw className="h-4 w-4" />
      {label}
    </Button>
  );
}
