package handlers

import (
	"aswadwk/messaging-task-go/internal/config"
	"aswadwk/messaging-task-go/internal/repositories"
	"aswadwk/messaging-task-go/internal/services"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMessageTestApp() (*fiber.App, *MessageHandler) {
	// Change to project root directory to ensure .env file is found
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..")
	os.Chdir(projectRoot)

	// Load existing config
	config.LoadConfig()

	// Initialize dependencies using existing config
	db := config.DBConnect()
	messageRepo := repositories.NewMessageRepository(db)
	rabbitService := services.NewRabbitMQ(config.Cfg.RabbitMQURL)
	tenantManager := services.NewTenantManager(rabbitService, messageRepo)
	publisherService := services.NewPublisherService(rabbitService)
	messageHandler := NewMessageHandler(publisherService, tenantManager)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return ctx.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	messages := app.Group("/messages")
	messages.Post("/", messageHandler.PublishMessage)
	messages.Get("/", messageHandler.GetMessages)

	return app, messageHandler
}

// Test PublishMessage - Success case
func TestPublishMessageSuccess(t *testing.T) {
	app, _ := setupMessageTestApp()

	payload := map[string]interface{}{
		"tenant_id": "test-tenant-123",
		"payload": map[string]interface{}{
			"message": "Hello World",
			"type":    "notification",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "Message published successfully", response["message"])
	assert.Equal(t, "test-tenant-123", response["tenant"])
	assert.NotNil(t, response["payload"])
}

// Test PublishMessage - Invalid JSON
func TestPublishMessageInvalidJSON(t *testing.T) {
	app, _ := setupMessageTestApp()

	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "Invalid JSON", response["error"])
}

// Test PublishMessage - Missing tenant_id
func TestPublishMessageMissingTenantID(t *testing.T) {
	app, _ := setupMessageTestApp()

	payload := map[string]interface{}{
		"payload": map[string]interface{}{
			"message": "Hello World",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "tenant_id is required", response["error"])
}

// Test PublishMessage - Empty tenant_id
func TestPublishMessageEmptyTenantID(t *testing.T) {
	app, _ := setupMessageTestApp()

	payload := map[string]interface{}{
		"tenant_id": "",
		"payload": map[string]interface{}{
			"message": "Hello World",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "tenant_id is required", response["error"])
}

// Test PublishMessage - Nil payload
func TestPublishMessageNilPayload(t *testing.T) {
	app, _ := setupMessageTestApp()

	payload := map[string]interface{}{
		"tenant_id": "test-tenant-123",
		"payload":   nil,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "payload cannot be empty", response["error"])
}

// Test PublishMessage - Missing payload
func TestPublishMessageMissingPayload(t *testing.T) {
	app, _ := setupMessageTestApp()

	payload := map[string]interface{}{
		"tenant_id": "test-tenant-123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "payload cannot be empty", response["error"])
}

// Test GetMessages - Success with default cursor
func TestGetMessagesSuccess(t *testing.T) {
	app, _ := setupMessageTestApp()

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// Response should be a valid JSON (messages array or object)
	assert.NotNil(t, response)
}

// Test GetMessages - Success with cursor parameter
func TestGetMessagesWithCursor(t *testing.T) {
	app, _ := setupMessageTestApp()

	req := httptest.NewRequest(http.MethodGet, "/messages?cursor=2", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// Response should be a valid JSON (messages array or object)
	assert.NotNil(t, response)
}

// Test GetMessages - Invalid cursor (non-integer)
func TestGetMessagesInvalidCursor(t *testing.T) {
	app, _ := setupMessageTestApp()

	req := httptest.NewRequest(http.MethodGet, "/messages?cursor=invalid", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "cursor must be an integer", response["error"])
}

// Test GetMessages - Empty cursor (should use default)
func TestGetMessagesEmptyCursor(t *testing.T) {
	app, _ := setupMessageTestApp()

	req := httptest.NewRequest(http.MethodGet, "/messages?cursor=", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// Response should be a valid JSON (messages array or object)
	assert.NotNil(t, response)
}
