package routes

import (
	"github.com/gofiber/fiber/v2"
)

func MessageRoute(router fiber.Router) {
	messages := router.Group("/messages")

	messages.Post("/", messageHandler.PublishMessage)
	messages.Get("/", messageHandler.GetMessages)
}
