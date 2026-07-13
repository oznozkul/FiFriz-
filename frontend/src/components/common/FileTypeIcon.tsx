import type { IconType } from "react-icons";
import { FaFile, FaFileExcel, FaFileImage, FaFilePdf, FaFileWord } from "react-icons/fa6";

export type FileTypeIconSize = "sm" | "md" | "lg" | "xl";

const iconSizeClass: Record<FileTypeIconSize, string> = {
  sm: "h-4 w-4",
  md: "h-5 w-5",
  lg: "h-6 w-6",
  xl: "h-8 w-8",
};

export function normalizeFileExtension(extension: string) {
  return extension.trim().toLowerCase().replace(/^\./, "");
}

export function extensionFromFileName(fileName: string) {
  const index = fileName.lastIndexOf(".");
  if (index === -1 || index === fileName.length - 1) return "";
  return normalizeFileExtension(fileName.slice(index + 1));
}

export function getFileTypeIconComponent(extension: string): IconType {
  const ext = normalizeFileExtension(extension);

  if (ext === "pdf") return FaFilePdf;
  if (["doc", "docx"].includes(ext)) return FaFileWord;
  if (["xls", "xlsx", "csv"].includes(ext)) return FaFileExcel;
  if (["png", "jpg", "jpeg", "gif", "webp", "bmp", "svg", "tif", "tiff", "heic"].includes(ext)) return FaFileImage;

  return FaFile;
}

export function getFileTypeLabel(extension: string) {
  const ext = normalizeFileExtension(extension);
  return ext ? ext.toUpperCase() : "FILE";
}

type FileTypeIconProps = {
  extension: string;
  size?: FileTypeIconSize;
  className?: string;
  title?: string;
};

export function FileTypeIcon({ extension, size = "md", className = "", title }: FileTypeIconProps) {
  const Icon = getFileTypeIconComponent(extension);
  const label = title ?? getFileTypeLabel(extension);

  return <Icon className={`shrink-0 ${iconSizeClass[size]} ${className}`} aria-label={label} title={label} />;
}
