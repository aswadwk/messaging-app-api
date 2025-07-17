package repositories

import (
	"aswadwk/messaging-task-go/dto"
	"aswadwk/messaging-task-go/internal/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Store(message dto.NewMessageDto) error
	CreatePartition(tenantID uuid.UUID) error
	DropPartition(tenantID uuid.UUID) error
	GetMessages(cursor int) (dto.QueryResponse, error)
}

type messageRepository struct {
	db *gorm.DB
}

// GetMessages implements MessageRepository.
func (m *messageRepository) GetMessages(cursor int) (dto.QueryResponse, error) {
	var messages []models.Message
	var total int64

	// Get total count
	if err := m.db.Model(&models.Message{}).Count(&total).Error; err != nil {
		return dto.QueryResponse{}, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error counting messages: %v", err))
	}

	// Calculate pagination
	perPage := 10
	offset := (cursor - 1) * perPage
	if cursor <= 0 {
		offset = 0
		cursor = 1
	}

	// Fetch messages
	if err := m.db.Model(&models.Message{}).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return dto.QueryResponse{}, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error retrieving messages: %v", err))
	}

	// Calculate last page
	lastPage := int(total) / perPage
	if int(total)%perPage > 0 {
		lastPage++
	}

	// Ensure data field is always an array, never null
	if messages == nil {
		messages = make([]models.Message, 0)
	}

	response := dto.QueryResponse{
		Total:    int(total),
		PerPage:  perPage,
		CurPage:  cursor,
		LastPage: lastPage,
		Data:     messages,
	}

	return response, nil
}

// CreatePartition implements MessageRepository.
func (m *messageRepository) CreatePartition(tenantID uuid.UUID) error {
	partitionName := `messages_tenant_` + tenantID.String()
	partitionNameQuoted := `"` + partitionName + `"`
	tenantIDStr := "'" + tenantID.String() + "'"
	query := "CREATE TABLE IF NOT EXISTS " + partitionNameQuoted + " PARTITION OF messages FOR VALUES IN (" + tenantIDStr + ")"
	if err := m.db.Exec(query).Error; err != nil {
		return err
	}
	return nil
}

// DropPartition implements MessageRepository.
func (m *messageRepository) DropPartition(tenantID uuid.UUID) error {
	partitionName := "messages_tenant_" + tenantID.String()
	query := "DROP TABLE IF EXISTS " + partitionName
	if err := m.db.Exec(query).Error; err != nil {
		return err
	}
	return nil
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// Store implements MessageRepository.
func (m *messageRepository) Store(message dto.NewMessageDto) error {
	ID, _ := uuid.NewV7()

	newMessage := models.Message{
		ID:       ID.String(),
		TenantID: message.TenantID,
		Payload:  message.Payload,
	}

	if err := m.db.Create(&newMessage).Error; err != nil {
		return err
	}

	return nil
}
