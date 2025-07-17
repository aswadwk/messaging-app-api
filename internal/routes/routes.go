package routes

import (
	"aswadwk/messaging-task-go/internal/config"
	"aswadwk/messaging-task-go/internal/handlers"
	"aswadwk/messaging-task-go/internal/repositories"
	"aswadwk/messaging-task-go/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	validate *validator.Validate

	// Repositories
	messageRepository repositories.MessageRepository

	// Services
	rabbitService    *services.RabbitMQ
	tenantService    *services.TenantManager
	publisherService *services.PublisherService

	// Handlers
	tenantHandler  *handlers.TenantHandler
	messageHandler *handlers.MessageHandler
)

func Init() {
	db = config.DBConnect()

	validate = validator.New(validator.WithRequiredStructEnabled())

	// Repository
	messageRepository = repositories.NewMessageRepository(db)

	// Services
	rabbitService = services.NewRabbitMQ(config.Cfg.RabbitMQURL)
	tenantService = services.NewTenantManager(rabbitService, messageRepository)
	publisherService = services.NewPublisherService(rabbitService)

	// Handlers
	tenantHandler = handlers.NewTenantHandler(tenantService)
	messageHandler = handlers.NewMessageHandler(publisherService, tenantService)
}

func SetupRoutes(app *fiber.App) {
	// api := app.Group("/v1")
	TenantRoute(app)
	MessageRoute(app)
}
