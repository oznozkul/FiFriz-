package documents

import (
	"encoding/json"
	"errors"
	"strings"

	"fifriz/internal/response"
	"fifriz/internal/stream"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	documents, err := h.service.List(c.UserContext(), c.Query("folderId"), c.Query("tagId"))
	if err != nil {
		if errors.Is(err, ErrInvalidFolderID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "folder id must be a valid UUID"))
		}
		if errors.Is(err, ErrFolderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("folder_not_found", "folder not found"))
		}
		if errors.Is(err, ErrInvalidTagID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_id", "tag id must be a valid UUID"))
		}
		if errors.Is(err, ErrTagNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("tag_not_found", "tag not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("documents_list_failed", "could not list documents"))
	}

	return c.JSON(response.OK(documents))
}

func (h *Handler) ListDeleted(c *fiber.Ctx) error {
	documents, err := h.service.ListDeleted(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("trash_list_failed", "could not list trash documents"))
	}

	return c.JSON(response.OK(documents))
}

func (h *Handler) Search(c *fiber.Ctx) error {
	results, err := h.service.Search(c.UserContext(), c.Query("q"))
	if err != nil {
		if errors.Is(err, ErrInvalidSearchQuery) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_search_query", "q query parameter is required"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_search_failed", "could not search documents"))
	}

	return c.JSON(response.OK(results))
}

func (h *Handler) SemanticSearch(c *fiber.Ctx) error {
	results, err := h.service.SemanticSearch(c.UserContext(), c.Query("q"))
	if err != nil {
		if errors.Is(err, ErrInvalidSearchQuery) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_search_query", "q query parameter is required"))
		}
		if errors.Is(err, ErrEmbeddingProviderDisabled) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("embedding_provider_disabled", "embedding provider is not configured"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("semantic_search_failed", "could not perform semantic search"))
	}

	return c.JSON(response.OK(results))
}

func (h *Handler) HybridSearch(c *fiber.Ctx) error {
	results, err := h.service.HybridSearch(c.UserContext(), c.Query("q"))
	if err != nil {
		if errors.Is(err, ErrInvalidSearchQuery) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_search_query", "q query parameter is required"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("hybrid_search_failed", "could not perform hybrid search"))
	}
	return c.JSON(response.OK(results))
}

func (h *Handler) ListFavorites(c *fiber.Ctx) error {
	documents, err := h.service.ListFavorites(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_favorites_list_failed", "could not list favorite documents"))
	}

	return c.JSON(response.OK(documents))
}

func (h *Handler) Favorite(c *fiber.Ctx) error {
	document, err := h.service.Favorite(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to favorite this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_favorite_failed", "could not favorite document"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) Unfavorite(c *fiber.Ctx) error {
	document, err := h.service.Unfavorite(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to unfavorite this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_unfavorite_failed", "could not unfavorite document"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	document, err := h.service.GetByID(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you do not have permission for this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_get_failed", "could not get document"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) GetWorkspace(c *fiber.Ctx) error {
	document, err := h.service.GetWorkspace(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you do not have permission for this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_workspace_get_failed", "could not get document workspace"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) GetFile(c *fiber.Ctx) error {
	accessMode := "preview"
	if c.Query("download") == "1" {
		accessMode = "download"
	}

	documentFile, err := h.service.GetFile(c.UserContext(), c.Params("id"), accessMode)
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you do not have permission for this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_file_get_failed", "could not get document file"))
	}

	if err := stream.SendFile(c, documentFile.Reader, documentFile.Size, documentFile.MimeType, documentFile.OriginalFileName, "inline"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_file_get_failed", "could not get document file"))
	}
	return nil
}

func (h *Handler) GetText(c *fiber.Ctx) error {
	documentText, err := h.service.GetText(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrDocumentTextNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_text_not_found", "document text not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_text_get_failed", "could not get document text"))
	}

	return c.JSON(response.OK(documentText))
}

func (h *Handler) GetChunks(c *fiber.Ctx) error {
	chunks, err := h.service.GetChunks(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_chunks_get_failed", "could not get document chunks"))
	}

	return c.JSON(response.OK(chunks))
}

func (h *Handler) GetChunkStats(c *fiber.Ctx) error {
	stats, err := h.service.GetChunkStats(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_chunk_stats_failed", "could not get document chunk stats"))
	}

	return c.JSON(response.OK(stats))
}

func (h *Handler) ReprocessText(c *fiber.Ctx) error {
	document, err := h.service.ReprocessText(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrTextExtractionUnsupported) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("text_extraction_unsupported", "text extraction is only supported for PDF, DOCX, XLSX, TXT and image documents"))
		}
		if errors.Is(err, ErrTesseractNotAvailable) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("tesseract_not_available", "Tesseract OCR is not installed. Install Tesseract with tur and eng language packs, or set TESSERACT_CMD."))
		}
		if errors.Is(err, ErrTextExtractionFailed) {
			return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("text_extraction_failed", "could not extract text from document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_text_reprocess_failed", "could not reprocess document text"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) Upload(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("missing_upload_file", "form field 'file' is required"))
	}

	document, err := h.service.PrepareUpload(fileHeader)
	if err != nil {
		if errors.Is(err, ErrMissingUploadFile) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("missing_upload_file", "form field 'file' is required"))
		}
		if errors.Is(err, ErrEmptyUploadFile) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("empty_upload_file", "file must not be empty"))
		}
		if errors.Is(err, ErrUnsupportedUploadFile) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("unsupported_upload_file", "supported formats are PDF, PNG, JPG, JPEG, DOCX and XLSX"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_upload_failed", "could not upload document"))
	}

	uploadReader, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("upload_file_open_failed", "could not read uploaded file"))
	}
	defer uploadReader.Close()

	documentResponse, err := h.service.SaveUploaded(c.UserContext(), document, uploadReader, fileHeader.Size, c.FormValue("folderId"), parseTagIDsFromForm(c.FormValue("tagIds")))
	if err != nil {
		if errors.Is(err, ErrInvalidFolderID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "folder id must be a valid UUID"))
		}
		if errors.Is(err, ErrFolderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("folder_not_found", "folder not found"))
		}
		if errors.Is(err, ErrInvalidTagID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_id", "tag id must be a valid UUID"))
		}
		if errors.Is(err, ErrTagNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("tag_not_found", "tag not found"))
		}
		if errors.Is(err, ErrFileTooLarge) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("file_too_large", "file exceeds maximum allowed size"))
		}
		if errors.Is(err, ErrMIMETypeNotAllowed) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("mime_type_not_allowed", "mime type is not allowed"))
		}
		if errors.Is(err, ErrInvalidMIMEContent) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_mime_content", "file content does not match extension"))
		}
		if errors.Is(err, ErrUnsupportedUploadFile) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("unsupported_upload_file", "supported formats are PDF, PNG, JPG, JPEG, DOCX and XLSX"))
		}
		if errors.Is(err, ErrUploadMetadataSaveFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("upload_metadata_save_failed", "could not save document metadata"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_upload_failed", "could not upload document"))
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(documentResponse))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var input UpdateDocumentRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	document, err := h.service.Update(c.UserContext(), c.Params("id"), input)
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrInvalidUpdateData) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_data", "original_file_name is required and status must be valid"))
		}
		if errors.Is(err, ErrInvalidFolderID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "folder id must be a valid UUID"))
		}
		if errors.Is(err, ErrFolderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("folder_not_found", "folder not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to manage this document"))
		}
		if errors.Is(err, ErrDocumentLocked) {
			return c.Status(fiber.StatusConflict).JSON(response.Fail("document_locked", "document is locked"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_update_failed", "could not update document"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to manage this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_delete_failed", "could not delete document"))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Restore(c *fiber.Ctx) error {
	document, err := h.service.Restore(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to manage this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_restore_failed", "could not restore document"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) PermanentlyDelete(c *fiber.Ctx) error {
	if err := h.service.PermanentlyDelete(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to manage this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_permanent_delete_failed", "could not permanently delete document"))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListByTag(c *fiber.Ctx) error {
	documents, err := h.service.ListByTag(c.UserContext(), c.Params("tagId"))
	if err != nil {
		if errors.Is(err, ErrInvalidTagID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_id", "tag id must be a valid UUID"))
		}
		if errors.Is(err, ErrTagNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("tag_not_found", "tag not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("documents_by_tag_failed", "could not list documents by tag"))
	}
	return c.JSON(response.OK(documents))
}

func (h *Handler) SetDocumentTags(c *fiber.Ctx) error {
	var input struct {
		TagIDs []string `json:"tagIds"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	document, err := h.service.SetDocumentTags(c.UserContext(), c.Params("id"), input.TagIDs)
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrInvalidTagID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_id", "tag id must be a valid UUID"))
		}
		if errors.Is(err, ErrTagNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("tag_not_found", "tag not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to update document tags"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_tags_update_failed", "could not update document tags"))
	}

	return c.JSON(response.OK(document))
}

func (h *Handler) GetSummary(c *fiber.Ctx) error {
	result, err := h.service.GetSummary(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrDocumentSummaryNotFound) {
			return c.JSON(fiber.Map{
				"success": true,
				"data":    nil,
			})
		}
		return h.handleSummaryError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) GenerateSummary(c *fiber.Ctx) error {
	result, err := h.service.GenerateSummary(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleSummaryError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) GetAIMetadata(c *fiber.Ctx) error {
	result, err := h.service.GetAIMetadata(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrDocumentAIMetadataNotFound) {
			return c.JSON(fiber.Map{
				"success": true,
				"data":    nil,
			})
		}
		return h.handleAIMetadataError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) GenerateAIMetadata(c *fiber.Ctx) error {
	result, err := h.service.GenerateAIMetadata(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleAIMetadataError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) handleAIMetadataError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
	case errors.Is(err, ErrMetadataProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("metadata_provider_disabled", "metadata service is not configured"))
	case errors.Is(err, ErrSummaryProviderUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("metadata_provider_unavailable", "AI metadata servisi şu anda kullanılamıyor. Lütfen daha sonra tekrar deneyin."))
	case errors.Is(err, ErrMetadataNoContent):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("metadata_no_content", "document has no extractable content for metadata"))
	case errors.Is(err, ErrMetadataParseFailed):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Fail("metadata_parse_failed", "AI metadata yanıtı işlenemedi"))
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to access this document metadata"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("metadata_failed", "could not process document metadata"))
	}
}

func (h *Handler) GetDocumentClassification(c *fiber.Ctx) error {
	result, err := h.service.GetDocumentClassification(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleClassificationError(c, err)
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

func (h *Handler) GenerateDocumentClassification(c *fiber.Ctx) error {
	result, err := h.service.GenerateDocumentClassification(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleClassificationError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) handleClassificationError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
	case errors.Is(err, ErrClassificationProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("classification_provider_disabled", "classification service is not configured"))
	case errors.Is(err, ErrSummaryProviderUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("classification_provider_unavailable", "AI sınıflandırma servisi şu anda kullanılamıyor. Lütfen daha sonra tekrar deneyin."))
	case errors.Is(err, ErrClassificationNoContent):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("classification_no_content", "document has no extractable content for classification"))
	case errors.Is(err, ErrClassificationParseFailed):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Fail("classification_parse_failed", "AI sınıflandırma yanıtı işlenemedi"))
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to access this document classification"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("classification_failed", "could not process document classification"))
	}
}

func (h *Handler) GetAIProgress(c *fiber.Ctx) error {
	progress, err := h.service.GetAIProgress(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to access this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("ai_progress_failed", "could not load ai progress"))
	}
	return c.JSON(response.OK(progress))
}

func (h *Handler) GetSuggestedTags(c *fiber.Ctx) error {
	items, err := h.service.GetSuggestedTags(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleSuggestedTagError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) AcceptSuggestedTag(c *fiber.Ctx) error {
	result, err := h.service.AcceptSuggestedTag(c.UserContext(), c.Params("id"), c.Params("suggestedTagId"))
	if err != nil {
		return h.handleSuggestedTagError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) RejectSuggestedTag(c *fiber.Ctx) error {
	result, err := h.service.RejectSuggestedTag(c.UserContext(), c.Params("id"), c.Params("suggestedTagId"))
	if err != nil {
		return h.handleSuggestedTagError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) AcceptAllSuggestedTags(c *fiber.Ctx) error {
	result, err := h.service.AcceptAllSuggestedTags(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleSuggestedTagError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) handleSuggestedTagError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
	case errors.Is(err, ErrSuggestedTagNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("suggested_tag_not_found", "suggested tag not found"))
	case errors.Is(err, ErrSuggestedTagInvalidStatus):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("suggested_tag_invalid_status", "suggested tag is not pending"))
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to manage suggested tags for this document"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("suggested_tags_failed", "could not process suggested tags"))
	}
}

func (h *Handler) handleSummaryError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
	case errors.Is(err, ErrDocumentSummaryNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("summary_not_found", "document summary not found"))
	case errors.Is(err, ErrSummaryProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("summary_provider_disabled", "summary service is not configured"))
	case errors.Is(err, ErrSummaryProviderUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("summary_provider_unavailable", "AI özet servisi şu anda kullanılamıyor. Ollama yanıt vermedi; lütfen birkaç dakika sonra tekrar deneyin."))
	case errors.Is(err, ErrSummaryNoContent):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("summary_no_content", "document has no extractable content for summary"))
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to access this document"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("summary_failed", "could not process document summary"))
	}
}

func parseTagIDsFromForm(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var tagIDs []string
	if err := json.Unmarshal([]byte(raw), &tagIDs); err == nil {
		return tagIDs
	}

	parts := strings.Split(raw, ",")
	tagIDs = make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			tagIDs = append(tagIDs, part)
		}
	}
	return tagIDs
}

func (h *Handler) LockDocument(c *fiber.Ctx) error {
	document, err := h.service.LockDocument(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeDocumentMutationError(c, err, "document_lock_failed", "could not lock document")
	}
	return c.JSON(response.OK(document))
}

func (h *Handler) UnlockDocument(c *fiber.Ctx) error {
	document, err := h.service.UnlockDocument(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeDocumentMutationError(c, err, "document_unlock_failed", "could not unlock document")
	}
	return c.JSON(response.OK(document))
}

func (h *Handler) GetDocumentHealth(c *fiber.Ctx) error {
	health, err := h.service.GetDocumentHealth(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
		}
		if errors.Is(err, ErrDocumentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
		}
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you do not have permission for this document"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_health_failed", "could not load document health"))
	}
	return c.JSON(response.OK(health))
}

func writeDocumentMutationError(c *fiber.Ctx, err error, code, message string) error {
	if errors.Is(err, ErrInvalidDocumentID) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	}
	if errors.Is(err, ErrDocumentNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("document_not_found", "document not found"))
	}
	if errors.Is(err, ErrForbidden) {
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to manage this document"))
	}
	if errors.Is(err, ErrDocumentLocked) {
		return c.Status(fiber.StatusConflict).JSON(response.Fail("document_locked", "document is locked"))
	}
	return c.Status(fiber.StatusInternalServerError).JSON(response.Fail(code, message))
}
