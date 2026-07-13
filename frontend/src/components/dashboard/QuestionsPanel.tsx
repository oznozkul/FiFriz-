const recentQuestions = [
  ["Geçen yılki kira sözleşmemi bul", "2 saat önce"],
  ["Arçelik televizyonumun garanti süresi ne zaman bitiyor?", "5 saat önce"],
  ["Mayıs ayındaki faturalarımı göster", "1 gün önce"],
  ["Otomobil sigörtam ne zaman yenileniyor?", "2 gün önce"],
  ["2024 yılına ait banka dekontları", "3 gün önce"],
];

export function QuestionsPanel() {
  return (
    <section className="panel overflow-hidden">
      <div className="px-5 py-4">
        <h2 className="font-bold text-ink">Son Sorular</h2>
      </div>
      <div className="divide-y divide-line">
        {recentQuestions.map(([question, time]) => (
          <button key={question} className="flex w-full items-center gap-4 px-5 py-4 text-left transition hover:bg-violet-soft/35 focus:outline-none focus:ring-4 focus:ring-inset focus:ring-violet-ring/70" type="button">
            <p className="min-w-0 flex-1 truncate font-medium text-ink">{question}</p>
            <span className="text-xs text-slate-500">{time}</span>
          </button>
        ))}
      </div>
      <button className="px-5 py-4 text-sm font-semibold text-violet-brand transition hover:text-violet-hover" type="button">
        Tümünü Gör →
      </button>
    </section>
  );
}
