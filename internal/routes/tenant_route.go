package routes

import (
	"github.com/gofiber/fiber/v2"
)

func TenantRoute(router fiber.Router) {
	tenants := router.Group("/tenants")

	tenants.Post("/", tenantHandler.CreateTenant)
	tenants.Delete("/:id", tenantHandler.DeleteTenant)
	// PUT /tenants/{id}/config/concurrency
	tenants.Put("/:id/config/concurrency", tenantHandler.UpdateConcurrency)
}
