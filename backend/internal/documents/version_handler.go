package documents

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListVersions(c *fiber.Ctx) error {
	items, err := h.service.ListVersions(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeVersionError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) UploadVersion(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("missing_upload_file", "form field 'file' is required"))
	}

	preparedDocument, err := h.service.PrepareUpload(fileHeader)
	if err != nil {
		return writeUploadPrepareError(c, err)
	}

	uploadReader, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("upload_file_open_failed", "could not read uploaded file"))
	}
	defer uploadReader.Close()

	result, err := h.service.UploadVersion(c.UserContext(), c.Params("id"), preparedDocument, c.FormValue("changeNote"), uploadReader, fileHeader.Size)
	if err != nil {
		return writeVersionError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) RestoreVersion(c *fiber.Ctx) error {
	document, err := h.service.RestoreVersion(c.UserContext(), c.Params("id"), c.Params("versionId"))
	if err != nil {
		return writeVersionError(c, err)
	}
	return c.JSON(response.OK(document))
}

func writeVersionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrInvalidVersionID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_version_id", "version id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrDocumentVersionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	case errors.Is(err, ErrMissingUploadFile):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("missing_upload_file", "form field 'file' is required"))
	case errors.Is(err, ErrEmptyUploadFile):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("empty_upload_file", "file must not be empty"))
	case errors.Is(err, ErrUnsupportedUploadFile):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("unsupported_upload_file", "supported formats are PDF, PNG, JPG, JPEG, DOCX and XLSX"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_version_failed", "could not process document version"))
	}
}

func writeUploadPrepareError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrMissingUploadFile):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("missing_upload_file", "form field 'file' is required"))
	case errors.Is(err, ErrEmptyUploadFile):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("empty_upload_file", "file must not be empty"))
	case errors.Is(err, ErrUnsupportedUploadFile):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("unsupported_upload_file", "supported formats are PDF, PNG, JPG, JPEG, DOCX and XLSX"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_upload_failed", "could not upload document"))
	}
}
