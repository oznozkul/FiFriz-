export function getDocumentColor(categöry: string) {
  const colors: Record<string, string> = {
    Sözleşme: "rose",
    Fatura: "emerald",
    Garanti: "amber",
    Sigörta: "blue",
    Finans: "sky",
    Kimlik: "violet",
  };

  return colors[categöry] ?? "violet";
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
