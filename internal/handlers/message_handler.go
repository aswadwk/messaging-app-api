package handlers

import (
	"aswadwk/messaging-task-go/internal/services"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type MessageHandler struct {
	Publisher     *services.PublisherService
	TenantManager *services.TenantManager
}

// NewMessageHandler constructor
func NewMessageHandler(
	publisher *services.PublisherService,
	tenantManager *services.TenantManager,
) *MessageHandler {
	return &MessageHandler{
		Publisher:     publisher,
		TenantManager: tenantManager,
	}
}

// PublishMessage publishes a message to the specified tenant
// @FileName		message_handler.go
// @Description	Publish a message to a tenant
// @Tags			Message
// @Accept			json
// @Produce		json
// @Param			body	body		dto.NewMessageDto	true	"Request body"	Example
// @Success		202	{object}	fiber.Map	"Message published"
// @Failure		400	{object}	fiber.Map	"Invalid request"
// @Failure		500	{object}	fiber.Map	"Internal server error"
// @Router			/messages [post]
func (h *MessageHandler) PublishMessage(ctx *fiber.Ctx) error {
	type payload struct {
		TenantID string         `json:"tenant_id"`
		Payload  map[string]any `json:"payload"`
	}
	var p payload
	if err := ctx.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
	}
	// Create a message with tenant ID and payload
	if p.TenantID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tenant_id is required")
	}
	if p.Payload == nil {
		return fiber.NewError(fiber.StatusBadRequest, "payload cannot be empty")
	}

	msg := services.Message{
		TenantID: p.TenantID,
		Payload:  p.Payload,
	}

	queueName := fmt.Sprintf("tenant_%s_queue", p.TenantID)

	err := h.Publisher.Publish(queueName, msg)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Message published successfully",
		"tenant":  p.TenantID,
		"payload": p.Payload,
	})
}

// GetMessages retrieves messages for a tenant with pagination
// @FileName		message_handler.go
// @Description	Get messages for a tenant with pagination
// @Tags			Message
// @Accept			json
// @Produce		json
// @Param			cursor	query		int	false	"Page cursor"	Example(1)
// @Success		200	{object}	dto.MessageResponseDto	"Messages retrieved"
// @Failure		400	{object}	fiber.Map	"Invalid request"
// @Failure		500	{object}	fiber.Map	"Internal server error"
// @Router			/messages [get]
func (h *MessageHandler) GetMessages(ctx *fiber.Ctx) error {
	cursor := ctx.Query("cursor", "1")
	if cursor == "" {
		return fiber.NewError(fiber.StatusBadRequest, "cursor is required")
	}

	cursorInt, err := strconv.Atoi(cursor)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cursor must be an integer")
	}

	messages, err := h.TenantManager.GetMessages(cursorInt)
	if err != nil {
		return err
	}

	return ctx.JSON(messages)
}
