package tags

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	items, err := h.service.List(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("tags_list_failed", "could not list tags"))
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) Search(c *fiber.Ctx) error {
	items, err := h.service.SearchTags(c.UserContext(), c.Query("q"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("tag_search_failed", "could not search tags"))
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	tag, err := h.service.GetByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeTagError(c, err, "tag_get_failed", "could not get tag")
	}
	return c.JSON(response.OK(tag))
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	stats, err := h.service.GetTagStats(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeTagError(c, err, "tag_stats_failed", "could not get tag stats")
	}
	return c.JSON(response.OK(stats))
}

func (h *Handler) ListAliases(c *fiber.Ctx) error {
	items, err := h.service.ListTagAliases(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeTagError(c, err, "tag_aliases_failed", "could not list tag aliases")
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreateAlias(c *fiber.Ctx) error {
	var input CreateTagAliasRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}
	alias, err := h.service.CreateTagAlias(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeTagError(c, err, "tag_alias_create_failed", "could not create tag alias")
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(alias))
}

func (h *Handler) DeleteAlias(c *fiber.Ctx) error {
	if err := h.service.DeleteTagAlias(c.UserContext(), c.Params("id"), c.Params("aliasId")); err != nil {
		return writeTagError(c, err, "tag_alias_delete_failed", "could not delete tag alias")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Merge(c *fiber.Ctx) error {
	var input MergeTagRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}
	tag, err := h.service.MergeTag(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeTagError(c, err, "tag_merge_failed", "could not merge tags")
	}
	return c.JSON(response.OK(tag))
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var input CreateTagRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	tag, err := h.service.Create(c.UserContext(), input)
	if err != nil {
		return writeTagError(c, err, "tag_create_failed", "could not create tag")
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(tag))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var input UpdateTagRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	tag, err := h.service.Update(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeTagError(c, err, "tag_update_failed", "could not update tag")
	}
	return c.JSON(response.OK(tag))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.UserContext(), c.Params("id")); err != nil {
		return writeTagError(c, err, "tag_delete_failed", "could not delete tag")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func writeTagError(c *fiber.Ctx, err error, fallbackCode string, fallbackMessage string) error {
	if errors.Is(err, ErrInvalidTagID) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_id", "tag id must be a valid UUID"))
	}
	if errors.Is(err, ErrInvalidTagData) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_data", "tag name is required"))
	}
	if errors.Is(err, ErrTagNameTooLong) {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("tag_name_too_long", "tag name is too long"))
	}
	if errors.Is(err, ErrTagNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("tag_not_found", "tag not found"))
	}
	if errors.Is(err, ErrTagNameConflict) || errors.Is(err, ErrTagAliasConflict) {
		return c.Status(fiber.StatusConflict).JSON(response.Fail("tag_name_conflict", "a tag with this name or alias already exists"))
	}
	if errors.Is(err, ErrForbidden) {
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to perform this tag action"))
	}
	if errors.Is(err, ErrSystemTagProtected) {
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("system_tag_protected", "system tags cannot be modified"))
	}
	return c.Status(fiber.StatusInternalServerError).JSON(response.Fail(fallbackCode, fallbackMessage))
}
