import type { TagBrief } from "./tag";

export type DocumentStatus =
  | "uploading"
  | "processing"
  | "ready"
  | "failed"
  | "archived"
  | "deleted"
  | "locked"
  | "expired"
  | "retention_locked"
  | "uploaded"
  | "processed";

export type ApiDocument = {
  id: string;
  tenant_id: string;
  folder_id: string | null;
  folder_name: string;
  uploaded_by_user_id: string;
  original_file_name: string;
  stored_file_name: string;
  file_path: string;
  fileUrl: string;
  file_extension: string;
  mime_type: string;
  file_size: number;
  compressedSize?: number | null;
  contentEncoding?: string;
  storageProvider?: string;
  storageStatus?: string;
  sha256Checksum?: string;
  status: DocumentStatus;
  lifecycleStatus?: DocumentStatus;
  hasExtractedText: boolean;
  isFavorite: boolean;
  favoritedAt: string | null;
  isLocked?: boolean;
  lockedAt?: string | null;
  tags?: TagBrief[];
  approvalStatus?: "none" | "pending" | "approved" | "rejected";
  sourceType?: string;
  sourceMetadata?: Record<string, unknown> | null;
  aiDocumentType?: string | null;
  aiDocumentTypeConfidence?: number | null;
  aiClassifiedAt?: string | null;
  created_at: string;
  updated_at: string;
};

export type DocumentUploadResult = {
  document: ApiDocument;
  warnings?: string[];
  duplicateOf?: {
    id: string;
    originalFileName: string;
  };
};

export type DocumentHealth = {
  previewGenerated: boolean;
  ocrGenerated: boolean;
  embeddingGenerated: boolean;
  summaryGenerated: boolean;
  metadataGenerated: boolean;
  aiCompleted: boolean;
  versionCount: number;
  storageStatus: string;
  processingStatus: string;
};

export type ApiResponse<T> = {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
};
