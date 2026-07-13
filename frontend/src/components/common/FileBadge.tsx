import type { FileBadgeType } from "../../types";
import { FileTypeIcon, type FileTypeIconSize } from "./FileTypeIcon";

type FileBadgeProps = {
  extension?: string;
  type?: FileBadgeType;
  color?: string;
  size?: FileTypeIconSize;
};

function legacyTypeToExtension(type: FileBadgeType) {
  if (type === "doc") return "docx";
  if (type === "sheet") return "xlsx";
  return "pdf";
}

export function FileBadge({ extension, type = "pdf", size = "md" }: FileBadgeProps) {
  const resolvedExtension = extension ?? legacyTypeToExtension(type);
  return <FileTypeIcon extension={resolvedExtension} size={size} />;
}
