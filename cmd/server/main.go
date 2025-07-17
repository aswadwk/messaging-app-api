// package main
package main

//	@title			Messaging Task API
//	@version		1.0.0
//	@description	API Documentation.
//	@contact.name	Hajar Aswad
//	@contact.email	hajaraswdkom@gmail.com
//	@BasePath		/
//	@schemes		http https
//// @securityDefinitions.apikey ApiKeyInHeader
//// @in              header
//// @name            X-API-KEY
//// @description     API Key
//// @description     Example: api_key_123

import (
	"aswadwk/messaging-task-go/internal/config"
	"aswadwk/messaging-task-go/internal/routes"
	"aswadwk/messaging-task-go/internal/utils"
	"fmt"
	"log"
	"os"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigration() error {
	log.Println("Running migration...")

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Cfg.DBUserName,
		config.Cfg.DBPassword,
		config.Cfg.DBHost,
		config.Cfg.DBPort,
		config.Cfg.DBName,
	)

	if config.Cfg.Debug {
		fmt.Println("Connecting to database with URL:", dbUrl)
	}

	m, err := migrate.New(
		"file://db/migrations",
		dbUrl,
	)

	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("✅ No migration needed. Database is up to date.")
			return nil
		}
		return fmt.Errorf("❌ Migration failed: %w", err)
	}

	log.Println("✅ Migration applied successfully.")
	return nil
}

func main() {
	config.LoadConfig()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigration(); err != nil {
			log.Fatal("Migration failed:", err)
		}

		return
	}

	routes.Init()

	// Create Fiber app with increased header limit
	app := fiber.New(fiber.Config{
		ServerHeader:          config.Cfg.AppName,
		ReadBufferSize:        16 * 1024, // 16KB (default: 4096)
		WriteBufferSize:       16 * 1024, // 16KB (default: 4096)
		DisableStartupMessage: false,     // Keep or set to true to disable startup message
		ErrorHandler:          utils.HandleError,
	})

	// Static Handler
	app.Static("/static", "./static")
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendFile("./static/favicon.ico")
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return utils.Output(c, "OK")
	})

	app.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
		data, err := os.ReadFile("./internal/docs/v3/openapi.json")
		if err != nil {
			return utils.Output(c, "Failed to load OpenAPI spec", false, 500)
		}
		c.Set("Content-Type", "application/json")
		return c.Send(data)
	})

	app.Get("/internal/docs/swagger.json", func(c *fiber.Ctx) error {
		data, err := os.ReadFile("./internal/docs/swagger.json")
		if err != nil {
			return utils.Output(c, "Failed to load Swagger spec", false, 500)
		}
		c.Set("Content-Type", "application/json")
		return c.Send(data)
	})

	app.Get("/swagger", func(c *fiber.Ctx) error {
		data, _ := os.ReadFile("./internal/docs/index.html")
		html := string(data)
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "./internal/docs/v3/openapi.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Ximply Api Documentation",
			},
			ShowSidebar: true,
			DarkMode:    true,
		})

		if err != nil {
			return err
		}

		c.Set("Content-Type", "text/html")
		return c.SendString(htmlContent)
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		// AllowCredentials: false,
		AllowHeaders: "*",
	}))

	app.Use(logger.New())

	if config.Cfg.AppEnv != "local" {
		app.Use(recover.New())
	}

	// app.Use(middleware.AuthMiddleware())
	routes.SetupRoutes(app)

	// Start Server
	port := config.Cfg.AppPort
	err := app.Listen(":" + port)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
	fmt.Printf("Server is running on port %s...\n", port)
}
