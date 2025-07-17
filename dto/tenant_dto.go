package dto

type CreateConsumerDto struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Workers  int    `json:"workers" validate:"required"`
}

type UpdateConcurrencyDto struct {
	Workers int `json:"workers" validate:"required"`
}
