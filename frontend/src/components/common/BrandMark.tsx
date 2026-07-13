const sizeClasses = {
  sm: "h-9 w-9",
  md: "h-12 w-12",
  lg: "h-16 w-16",
} as const;

type BrandMarkProps = {
  size?: keyof typeof sizeClasses;
  className?: string;
};

export function BrandMark({ size = "sm", className }: BrandMarkProps) {
  return (
    <div className={`shrink-0 overflow-hidden rounded-lg ${sizeClasses[size]}${className ? ` ${className}` : ""}`}>
      <img src="/fifriz-icon.png" alt="" className="h-full w-full object-cover" draggable={false} />
    </div>
  );
}
