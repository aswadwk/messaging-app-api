package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

type JSONB map[string]any

func (j *JSONB) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to scan JSONB value")
	}

	return json.Unmarshal(bytes, &j)
}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Message struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Payload   JSONB     `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
