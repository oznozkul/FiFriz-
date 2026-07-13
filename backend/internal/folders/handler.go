package folders

import (
	"errors"

	"fifriz/internal/documents"
	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service         *Service
	documentService *documents.Service
}

func NewHandler(service *Service, documentService *documents.Service) *Handler {
	return &Handler{service: service, documentService: documentService}
}

func (h *Handler) List(c *fiber.Ctx) error {
	folders, err := h.service.List(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("folders_list_failed", "could not list folders"))
	}
	return c.JSON(response.OK(folders))
}

func (h *Handler) Search(c *fiber.Ctx) error {
	results, err := h.service.SearchFolders(c.UserContext(), c.Query("q"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("folder_search_failed", "could not search folders"))
	}
	return c.JSON(response.OK(results))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	folder, err := h.service.GetByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_get_failed", "could not get folder")
	}
	return c.JSON(response.OK(folder))
}

func (h *Handler) GetMetadata(c *fiber.Ctx) error {
	metadata, err := h.service.GetFolderMetadata(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_metadata_failed", "could not get folder metadata")
	}
	return c.JSON(response.OK(metadata))
}

func (h *Handler) GetBreadcrumb(c *fiber.Ctx) error {
	items, err := h.service.GetBreadcrumb(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_breadcrumb_failed", "could not get folder breadcrumb")
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) GetPermissions(c *fiber.Ctx) error {
	items, err := h.service.GetFolderPermissions(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_permissions_failed", "could not get folder permissions")
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) UpdatePermissions(c *fiber.Ctx) error {
	var input UpdateFolderPermissionsRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}
	result, err := h.service.UpdateFolderPermissions(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeFolderError(c, err, "folder_permissions_update_failed", "could not update folder permissions")
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var input CreateFolderRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	folder, err := h.service.Create(c.UserContext(), input)
	if err != nil {
		return writeFolderError(c, err, "folder_create_failed", "could not create folder")
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(folder))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var input UpdateFolderRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	folder, err := h.service.Update(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeFolderError(c, err, "folder_update_failed", "could not update folder")
	}
	return c.JSON(response.OK(folder))
}

func (h *Handler) Move(c *fiber.Ctx) error {
	var input MoveFolderRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}
	folder, err := h.service.MoveFolder(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeFolderError(c, err, "folder_move_failed", "could not move folder")
	}
	return c.JSON(response.OK(folder))
}

func (h *Handler) Lock(c *fiber.Ctx) error {
	folder, err := h.service.LockFolder(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_lock_failed", "could not lock folder")
	}
	return c.JSON(response.OK(folder))
}

func (h *Handler) Unlock(c *fiber.Ctx) error {
	folder, err := h.service.UnlockFolder(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_unlock_failed", "could not unlock folder")
	}
	return c.JSON(response.OK(folder))
}

func (h *Handler) Restore(c *fiber.Ctx) error {
	folder, err := h.service.RestoreFolder(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeFolderError(c, err, "folder_restore_failed", "could not restore folder")
	}
	return c.JSON(response.OK(folder))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.UserContext(), c.Params("id")); err != nil {
		return writeFolderError(c, err, "folder_delete_failed", "could not delete folder")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListDocuments(c *fiber.Ctx) error {
	documentItems, err := h.documentService.ListByFolder(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, documents.ErrInvalidDocumentID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "folder id must be a valid UUID"))
		}
		if errors.Is(err, documents.ErrInvalidFolderID) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "folder id must be a valid UUID"))
		}
		if errors.Is(err, documents.ErrFolderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("folder_not_found", "folder not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("folder_documents_list_failed", "could not list folder documents"))
	}
	h.service.TrackOpened(c.UserContext(), c.Params("id"))
	return c.JSON(response.OK(documentItems))
}

func writeFolderError(c *fiber.Ctx, err error, fallbackCode string, fallbackMessage string) error {
	if errors.Is(err, ErrInvalidFolderID) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "folder id must be a valid UUID"))
	}
	if errors.Is(err, ErrInvalidFolderData) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_data", "folder name is required"))
	}
	if errors.Is(err, ErrFolderNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("folder_not_found", "folder not found"))
	}
	if errors.Is(err, ErrFolderNameConflict) {
		return c.Status(fiber.StatusConflict).JSON(response.Fail("folder_name_conflict", "a folder with this name already exists under the selected parent"))
	}
	if errors.Is(err, ErrForbidden) {
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to perform this folder action"))
	}
	if errors.Is(err, ErrSystemFolderProtected) {
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("system_folder_protected", "system folders cannot be modified"))
	}
	if errors.Is(err, ErrFolderLocked) {
		return c.Status(fiber.StatusConflict).JSON(response.Fail("folder_locked", "folder is locked"))
	}
	if errors.Is(err, ErrFolderMoveCycle) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("folder_move_cycle", "folder cannot be moved into its own descendant"))
	}
	if errors.Is(err, ErrFolderMoveIntoSelf) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("folder_move_into_self", "folder cannot be moved into itself"))
	}
	return c.Status(fiber.StatusInternalServerError).JSON(response.Fail(fallbackCode, fallbackMessage))
}
