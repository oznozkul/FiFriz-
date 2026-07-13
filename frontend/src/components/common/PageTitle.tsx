type PageTitleProps = {
  title: string;
  subtitle?: string;
};

export function PageTitle({ title, subtitle }: PageTitleProps) {
  return (
    <div className="min-w-0">
      <h1 className="text-2xl font-bold tracking-tight text-ink md:text-3xl">{title}</h1>
      {subtitle ? <p className="mt-1.5 max-w-3xl text-sm leading-relaxed text-slate-500">{subtitle}</p> : null}
    </div>
  );
}
