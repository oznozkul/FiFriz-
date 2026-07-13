import { Loader2, RefreshCw, Sparkles, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { PrimaryButton } from "../common";
import {
  deleteDocumentEntity,
  extractDocumentEntities,
  getDocumentEntities,
} from "../../services/documentEntityService";
import {
  ENTITY_TYPE_LABELS,
  groupEntities,
  type DocumentEntity,
} from "../../types/documentEntity";

type DocumentEntitiesPanelProps = {
  documentId: string;
  refreshToken?: number;
};

function formatConfidence(confidence?: number | null) {
  if (confidence == null || Number.isNaN(confidence)) {
    return null;
  }
  return `${Math.round(confidence * 100)}%`;
}

function EntityCard({
  entity,
  isDeleting,
  onDelete,
}: {
  entity: DocumentEntity;
  isDeleting: boolean;
  onDelete: (entity: DocumentEntity) => void;
}) {
  const confidence = formatConfidence(entity.confidence);

  return (
    <div className="rounded-xl border border-line bg-slate-50/70 px-3 py-3">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p className="font-semibold text-ink">{entity.value}</p>
          {entity.label ? <p className="mt-0.5 text-xs font-medium text-slate-500">{entity.label}</p> : null}
          <p className="mt-1 text-[11px] font-semibold uppercase tracking-wide text-violet-brand">
            {ENTITY_TYPE_LABELS[entity.entityType] ?? entity.entityType}
          </p>
        </div>
        <button
          type="button"
          className="rounded-lg p-1.5 text-slate-400 transition hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
          aria-label="Varlığı sil"
          disabled={isDeleting}
          onClick={() => onDelete(entity)}
        >
          {isDeleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
        </button>
      </div>
      {confidence ? <p className="mt-2 text-xs text-slate-500">Güven: {confidence}</p> : null}
      {entity.sourceText ? (
        <p className="mt-2 line-clamp-3 text-xs italic text-slate-500">"{entity.sourceText}"</p>
      ) : null}
    </div>
  );
}

export function DocumentEntitiesPanel({ documentId, refreshToken = 0 }: DocumentEntitiesPanelProps) {
  const [entities, setEntities] = useState<DocumentEntity[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isExtracting, setIsExtracting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadEntities = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await getDocumentEntities(documentId);
      setEntities(result);
    } catch (loadError) {
      setEntities([]);
      setError(loadError instanceof Error ? loadError.message : "AI varlıklar yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadEntities();
  }, [documentId, loadEntities, refreshToken]);

  const handleExtract = async () => {
    setIsExtracting(true);
    setError(null);
    try {
      const result = await extractDocumentEntities(documentId);
      setEntities(result);
    } catch (extractError) {
      setError(extractError instanceof Error ? extractError.message : "Varlıklar çıkarılamadı. Lütfen tekrar deneyin.");
    } finally {
      setIsExtracting(false);
    }
  };

  const handleDelete = async (entity: DocumentEntity) => {
    setDeletingId(entity.id);
    setError(null);
    try {
      await deleteDocumentEntity(documentId, entity.id);
      setEntities((current) => current.filter((item) => item.id !== entity.id));
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : "Varlık silinemedi.");
    } finally {
      setDeletingId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI varlıklar yükleniyor...
      </div>
    );
  }

  if (isExtracting) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI varlıklar çıkarılıyor...
      </div>
    );
  }

  const grouped = groupEntities(entities);

  return (
    <div className="space-y-4">
      {error ? <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p> : null}

      {grouped.length === 0 ? (
        <div className="rounded-xl border border-dashed border-line bg-slate-50/80 px-4 py-5 text-center">
          <Sparkles className="mx-auto h-5 w-5 text-violet-brand" />
          <p className="mt-2 text-sm text-slate-600">Bu belgeden henüz AI varlık çıkarılmadı.</p>
        </div>
      ) : (
        grouped.map((group) => (
          <section key={group.key}>
            <h4 className="text-xs font-bold uppercase tracking-wide text-slate-400">{group.label}</h4>
            <div className="mt-2 space-y-2">
              {group.items.map((entity) => (
                <EntityCard
                  key={entity.id}
                  entity={entity}
                  isDeleting={deletingId === entity.id}
                  onDelete={(item) => void handleDelete(item)}
                />
              ))}
            </div>
          </section>
        ))
      )}

      <PrimaryButton type="button" className="w-full" onClick={() => void handleExtract()}>
        <RefreshCw className="h-4 w-4" />
        AI ile Varlıkları Çıkar
      </PrimaryButton>
    </div>
  );
}
