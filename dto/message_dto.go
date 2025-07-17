package dto

import "time"

type NewMessageDto struct {
	TenantID string         `json:"tenant_id" validate:"required"`
	Payload  map[string]any `json:"payload" validate:"required"`
}

type MessageDto struct {
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id"`
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
}

type MessageResponseDto struct {
	Data       []MessageDto `json:"data"`
	NextCursor string       `json:"next_cursor"`
}
