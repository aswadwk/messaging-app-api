package handlers

import (
	"aswadwk/messaging-task-go/dto"
	"aswadwk/messaging-task-go/internal/config"
	"aswadwk/messaging-task-go/internal/repositories"
	"aswadwk/messaging-task-go/internal/services"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp() (*fiber.App, *TenantHandler) {
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
	tenantHandler := NewTenantHandler(tenantManager)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	tenants := app.Group("/tenants")
	tenants.Post("/", tenantHandler.CreateTenant)
	tenants.Delete("/:id", tenantHandler.DeleteTenant)
	tenants.Put("/:id/config/concurrency", tenantHandler.UpdateConcurrency)

	return app, tenantHandler
}

// Test Create Tenant - Positive Cases
func TestCreateTenantSuccess(t *testing.T) {
	app, _ := setupTestApp()

	tenantID := uuid.New()
	createDto := dto.CreateConsumerDto{
		TenantID: tenantID.String(),
		Workers:  5,
	}

	body, err := json.Marshal(createDto)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant created", response["message"])
}

func TestCreateTenantDefaultWorkers(t *testing.T) {
	app, _ := setupTestApp()

	tenantID := uuid.New()
	createDto := dto.CreateConsumerDto{
		TenantID: tenantID.String(),
		Workers:  0, // Should default to 3
	}

	body, err := json.Marshal(createDto)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant created", response["message"])
}

// Test Create Tenant - Negative Cases
func TestCreateTenantInvalidJSON(t *testing.T) {
	app, _ := setupTestApp()

	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid JSON", response["error"])
}

func TestCreateTenantInvalidTenantID(t *testing.T) {
	app, _ := setupTestApp()

	createDto := dto.CreateConsumerDto{
		TenantID: "invalid-uuid",
		Workers:  5,
	}

	body, err := json.Marshal(createDto)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid tenant_id", response["error"])
}

// Test Delete Tenant - Positive Cases
func TestDeleteTenantSuccess(t *testing.T) {
	app, _ := setupTestApp()

	// First create a tenant
	tenantID := uuid.New()
	createDto := dto.CreateConsumerDto{
		TenantID: tenantID.String(),
		Workers:  3,
	}

	body, err := json.Marshal(createDto)
	require.NoError(t, err)

	req1 := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	resp1, err := app.Test(req1, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Now delete the tenant
	req2 := httptest.NewRequest(http.MethodDelete, "/tenants/"+tenantID.String(), nil)
	resp2, err := app.Test(req2, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant stopped", response["message"])
}

// Test Delete Tenant - Negative Cases
func TestDeleteTenantInvalidTenantID(t *testing.T) {
	app, _ := setupTestApp()

	req := httptest.NewRequest(http.MethodDelete, "/tenants/invalid-uuid", nil)
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid tenant_id", response["error"])
}

// Test Update Concurrency - Positive Cases
func TestUpdateConcurrencySuccess(t *testing.T) {
	app, _ := setupTestApp()

	// First create a tenant
	tenantID := uuid.New()
	createDto := dto.CreateConsumerDto{
		TenantID: tenantID.String(),
		Workers:  3,
	}

	body, err := json.Marshal(createDto)
	require.NoError(t, err)

	req1 := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	resp1, err := app.Test(req1, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Now update concurrency
	updateDto := map[string]int{
		"workers": 8,
	}

	updateBody, err := json.Marshal(updateDto)
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodPut, "/tenants/"+tenantID.String()+"/config/concurrency", bytes.NewReader(updateBody))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := app.Test(req2, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Concurrency updated", response["message"])
}

// Test Update Concurrency - Negative Cases
func TestUpdateConcurrencyInvalidTenantID(t *testing.T) {
	app, _ := setupTestApp()

	updateDto := map[string]int{
		"workers": 8,
	}

	updateBody, err := json.Marshal(updateDto)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/tenants/invalid-uuid/config/concurrency", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid tenant_id", response["error"])
}

func TestUpdateConcurrencyInvalidJSON(t *testing.T) {
	app, _ := setupTestApp()

	tenantID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/tenants/"+tenantID.String()+"/config/concurrency", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid JSON", response["error"])
}

func TestUpdateConcurrencyZeroWorkers(t *testing.T) {
	app, _ := setupTestApp()

	tenantID := uuid.New()
	updateDto := map[string]int{
		"workers": 0,
	}

	updateBody, err := json.Marshal(updateDto)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/tenants/"+tenantID.String()+"/config/concurrency", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response fiber.Map
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	assert.Equal(t, "Workers must be > 0", response["error"])
}
