package handlers

import (
	"aswadwk/messaging-task-go/dto"
	"aswadwk/messaging-task-go/internal/services"
	"context"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TenantHandler struct {
	Manager *services.TenantManager
}

// NewTenantHandler constructor
func NewTenantHandler(manager *services.TenantManager) *TenantHandler {
	return &TenantHandler{
		Manager: manager,
	}
}

// POST /tenants
// @FileName		tenant_handler.go
// @Description	Create a new tenant
// @Tags			Tenant
// @Accept			json
// @Produce		json
// @Param			body	body		dto.CreateConsumerDto	true	"Request body"	Example
// @Success		201	{object}	fiber.Map	"Tenant created"
// @Failure		400	{object}	fiber.Map	"Invalid request"
// @Failure		500	{object}	fiber.Map	"Internal server error"
// @Router			/tenants [post]
func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	createDto := dto.CreateConsumerDto{}

	if err := c.BodyParser(&createDto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
	}

	tenantID, err := uuid.Parse(createDto.TenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid tenant_id")
	}
	if createDto.Workers <= 0 {
		createDto.Workers = 3 // default
	}

	// Create partition for tenant
	if err := h.Manager.CreatePartition(tenantID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if err := h.Manager.StartTenantConsumer(context.Background(), tenantID, createDto.Workers); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	log.Printf("[API] Tenant created: %s with %d workers", createDto.TenantID, createDto.Workers)
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Tenant created",
	})
}

// @FileName		tenant_handler.go
// @Description	Tenant handler
// @Tags			Tenant
// @Accept			json
// @Produce		json
// @Param			id	path		string																			true					"Tenant ID"			Example	("tenant-123")
// @Router			/tenants/{id} [delete]
// DeleteTenant
// @Success		200	{object}	fiber.Map	"Tenant stopped"
// @Failure		400	{object}	fiber.Map	"Invalid tenant_id"
// @Failure		500	{object}	fiber.Map	"Internal server error"
func (h *TenantHandler) DeleteTenant(c *fiber.Ctx) error {
	tenantIDStr := c.Params("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid tenant_id")
	}

	if err := h.Manager.StopTenantConsumer(tenantID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	log.Printf("[API] Tenant deleted: %s", tenantID)
	return c.JSON(fiber.Map{
		"message": "Tenant stopped",
	})
}

// @FileName		tenant_handler.go
// @Description	Tenant handler
// @Tags			Tenant
// @Accept			json
// @Accept			multipart/form-data
// @Accept			plain
// @Produce		json
// @Param			tenantId	path		string																			true					"Tenant ID"			Example("tenant-123")
// @Param			body	body		object{workers=int}														false					"Request body"		Example({"workers": 5})	default({"workers": 3})
// @Router			/tenants/{id}/config/concurrency [put]
// UpdateTenantConcurrency updates the concurrency for a tenant
func (h *TenantHandler) UpdateConcurrency(c *fiber.Ctx) error {
	tenantIDStr := c.Params("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid tenant_id")
	}

	type request struct {
		Workers int `json:"workers"`
	}
	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
	}

	if req.Workers <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Workers must be > 0")
	}

	// Stop old consumer
	if err := h.Manager.StopTenantConsumer(tenantID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Start new consumer with new worker count
	if err := h.Manager.StartTenantConsumer(context.Background(), tenantID, req.Workers); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	log.Printf("[API] Tenant %s concurrency updated to %d", tenantID, req.Workers)
	return c.JSON(fiber.Map{
		"message": "Concurrency updated",
	})
}
