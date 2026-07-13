import { Spinner } from "./Spinner";
import { CardSkeleton, TableSkeleton } from "./Skeleton";

type PageLoaderProps = {
  label?: string;
  variant?: "spinner" | "table" | "cards";
  rows?: number;
  cards?: number;
  className?: string;
};

export function PageLoader({
  label = "Yükleniyor...",
  variant = "spinner",
  rows = 5,
  cards = 4,
  className = "",
}: PageLoaderProps) {
  if (variant === "table") {
    return (
      <section className={`panel overflow-hidden rounded-2xl ${className}`}>
        <TableSkeleton rows={rows} />
      </section>
    );
  }

  if (variant === "cards") {
    return (
      <div className={`grid gap-4 sm:grid-cols-2 xl:grid-cols-4 ${className}`}>
        <CardSkeleton count={cards} />
      </div>
    );
  }

  return (
    <section className={`panel flex items-center justify-center gap-3 rounded-2xl p-8 ${className}`}>
      <Spinner size="md" label={label} />
    </section>
  );
}
