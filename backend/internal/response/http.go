package response

import "github.com/gofiber/fiber/v2"

func WriteOK(c *fiber.Ctx, data any) error {
	return c.JSON(OK(data))
}

func WriteCreated(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(OK(data))
}

func WriteNoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func WriteError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(Fail(code, message))
}

func WriteBadRequest(c *fiber.Ctx, code, message string) error {
	return WriteError(c, fiber.StatusBadRequest, code, message)
}

func WriteUnauthorized(c *fiber.Ctx, code, message string) error {
	return WriteError(c, fiber.StatusUnauthorized, code, message)
}

func WriteForbidden(c *fiber.Ctx, code, message string) error {
	return WriteError(c, fiber.StatusForbidden, code, message)
}

func WriteNotFound(c *fiber.Ctx, code, message string) error {
	return WriteError(c, fiber.StatusNotFound, code, message)
}

func WriteConflict(c *fiber.Ctx, code, message string) error {
	return WriteError(c, fiber.StatusConflict, code, message)
}

func WriteInternalError(c *fiber.Ctx, code, message string) error {
	return WriteError(c, fiber.StatusInternalServerError, code, message)
}
