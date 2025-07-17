package services

import (
	"aswadwk/messaging-task-go/dto"
	"aswadwk/messaging-task-go/internal/models"
	"aswadwk/messaging-task-go/internal/repositories"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type TenantManager struct {
	rabbit            *RabbitMQ
	consumers         map[string]*TenantConsumer
	mu                sync.Mutex
	messageRepository repositories.MessageRepository
}

// TenantConsumer menyimpan control untuk setiap tenant
type TenantConsumer struct {
	stopChan   chan struct{}
	doneChan   chan struct{}
	workerPool *WorkerPool // Optional: kalau kamu pakai worker pool
}

// NewTenantManager inisialisasi manager
func NewTenantManager(
	rabbit *RabbitMQ,
	messageRepo repositories.MessageRepository,
) *TenantManager {
	return &TenantManager{
		rabbit:            rabbit,
		consumers:         make(map[string]*TenantConsumer),
		messageRepository: messageRepo,
	}
}

// StartTenantConsumer membuat queue & mulai consumer baru
func (tm *TenantManager) StartTenantConsumer(ctx context.Context, tenantID uuid.UUID, concurrency int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	id := tenantID.String()
	if _, exists := tm.consumers[id]; exists {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("tenant %s already exists", id))
	}

	queueName := fmt.Sprintf("tenant_%s_queue", id)

	// Declare queue
	if _, err := tm.rabbit.DeclareQueue(queueName); err != nil {
		return err
	}

	// Start consuming
	consumerTag := fmt.Sprintf("consumer_%s", id)
	msgs, err := tm.rabbit.ConsumeMessages(queueName, consumerTag)
	if err != nil {
		return err
	}

	stop := make(chan struct{})
	done := make(chan struct{})

	// OPTIONAL: Buat worker pool per tenant
	pool := NewWorkerPool(concurrency)

	// Jalankan goroutine consumer
	go func() {
		log.Printf("[TenantManager] Consumer started for tenant %s", id)
		defer close(done)

		for {
			select {
			case msg := <-msgs:
				if pool != nil {
					pool.Submit(func() {
						tm.handleMessage(id, msg)
					})
				} else {
					tm.handleMessage(id, msg)
				}
			case <-stop:
				log.Printf("[TenantManager] Stopping consumer for tenant %s", id)
				return
			}
		}
	}()

	tm.consumers[id] = &TenantConsumer{
		stopChan:   stop,
		doneChan:   done,
		workerPool: pool,
	}

	return nil
}

// StopTenantConsumer menghentikan consumer tenant
func (tm *TenantManager) StopTenantConsumer(tenantID uuid.UUID) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	id := tenantID.String()
	consumer, ok := tm.consumers[id]
	if !ok {
		return fmt.Errorf("tenant %s not found", id)
	}

	// Stop worker pool
	if consumer.workerPool != nil {
		consumer.workerPool.Stop()
	}

	// Stop consuming goroutine
	close(consumer.stopChan)
	<-consumer.doneChan

	// Optional: Hapus queue (kalau memang mau hapus)
	queueName := fmt.Sprintf("tenant_%s_queue", id)
	if err := tm.rabbit.DeleteQueue(queueName); err != nil {
		log.Printf("[TenantManager] Failed to delete queue: %v", err)
	}

	delete(tm.consumers, id)
	log.Printf("[TenantManager] Consumer for tenant %s stopped", id)

	return nil
}

// GracefulShutdown mematikan semua tenant consumer
func (tm *TenantManager) GracefulShutdown() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for tenantID, consumer := range tm.consumers {
		log.Printf("[Shutdown] Stopping tenant: %s", tenantID)
		if consumer.workerPool != nil {
			consumer.workerPool.Stop()
		}
		close(consumer.stopChan)
		<-consumer.doneChan
	}
}

func (tm *TenantManager) CreatePartition(tenantID uuid.UUID) error {
	if err := tm.messageRepository.CreatePartition(tenantID); err != nil {
		return fmt.Errorf("failed to create partition for tenant %s: %w", tenantID, err)
	}

	log.Printf("[TenantManager] Partition created for tenant %s", tenantID)
	return nil
}

func (tm *TenantManager) GetMessages(cursor int) (dto.QueryResponse, error) {
	return tm.messageRepository.GetMessages(cursor)
}

func (tm *TenantManager) handleMessage(tenantID string, msg amqp.Delivery) {
	log.Printf("[Tenant %s] Received: %s", tenantID, msg.Body)

	tm.messageRepository.Store(dto.NewMessageDto{
		TenantID: tenantID,
		Payload:  models.JSONB{"content": string(msg.Body)},
	})
}
