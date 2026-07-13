type SkeletonProps = {
  className?: string;
};

export function Skeleton({ className = "" }: SkeletonProps) {
  return <div className={`animate-pulse rounded bg-slate-100 ${className}`} />;
}

export function StatCardSkeleton() {
  return (
    <div className="panel p-5">
      <div className="mb-5 flex items-center justify-between">
        <Skeleton className="h-10 w-10 rounded-lg" />
        <Skeleton className="h-3 w-16" />
      </div>
      <Skeleton className="mb-3 h-4 w-24" />
      <Skeleton className="h-8 w-20" />
    </div>
  );
}

export function TableRowSkeleton() {
  return (
    <div className="grid grid-cols-[36px_minmax(260px,1fr)_160px_110px_160px_48px] items-center px-5 py-4">
      <Skeleton className="h-4 w-4 rounded" />
      <div className="flex items-center gap-3">
        <Skeleton className="h-8 w-8 rounded-lg" />
        <Skeleton className="h-4 w-48 max-w-full" />
      </div>
      <Skeleton className="h-6 w-24" />
      <Skeleton className="h-4 w-14" />
      <Skeleton className="h-4 w-24" />
      <Skeleton className="h-8 w-8 rounded-lg" />
    </div>
  );
}

export function ChatMessageSkeleton() {
  return (
    <div className="max-w-[760px] rounded-lg border border-line bg-white p-4">
      <Skeleton className="mb-3 h-4 w-3/4" />
      <Skeleton className="h-4 w-1/2" />
    </div>
  );
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="divide-y divide-line">
      {Array.from({ length: rows }).map((_, index) => (
        <TableRowSkeleton key={index} />
      ))}
    </div>
  );
}

export function CardSkeleton({ count = 4 }: { count?: number }) {
  return (
    <>
      {Array.from({ length: count }).map((_, index) => (
        <StatCardSkeleton key={index} />
      ))}
    </>
  );
}
