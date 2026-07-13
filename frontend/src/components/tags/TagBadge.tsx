import type { TagBrief } from "../../types/tag";

type TagBadgeProps = {
  tag: TagBrief;
  onRemove?: () => void;
};

export function TagBadge({ tag, onRemove }: TagBadgeProps) {
  const color = tag.color || "#7c3aed";

  return (
    <span
      className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-semibold"
      style={{ backgroundColor: `${color}1a`, color }}
    >
      <span>{tag.name}</span>
      {onRemove ? (
        <button className="rounded-full px-1 transition hover:bg-black/5" type="button" onClick={onRemove} aria-label={`${tag.name} etiketini kaldır`}>
          ×
        </button>
      ) : null}
    </span>
  );
}

type TagBadgeListProps = {
  tags?: TagBrief[];
  onRemove?: (tagId: string) => void;
  emptyLabel?: string;
};

export function TagBadgeList({ tags = [], onRemove, emptyLabel }: TagBadgeListProps) {
  if (tags.length === 0) {
    return emptyLabel ? <span className="text-xs text-slate-400">{emptyLabel}</span> : null;
  }

  return (
    <div className="flex flex-wrap gap-1.5">
      {tags.map((tag) => (
        <TagBadge key={tag.id} tag={tag} onRemove={onRemove ? () => onRemove(tag.id) : undefined} />
      ))}
    </div>
  );
}
