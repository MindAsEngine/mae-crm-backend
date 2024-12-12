package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reporting-service/internal/domain"
	"reporting-service/internal/repository"
	"strconv"
	//"time"

	"github.com/gin-gonic/gin"
	//"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type AudienceService struct {
	Repo   repository.AudienceRepo
	Cache  repository.RedisCache
	Broker repository.RabbitMQ
}

// Конструктор
func NewAudienceService(repo *repository.AudienceRepo, cache *repository.RedisCache, broker *repository.RabbitMQ) *AudienceService {
	return &AudienceService{
		Repo:   *repo,
		Cache:  *cache,
		Broker: *broker,
	}
}

func (s *AudienceService) CreateAudience(c *gin.Context) {
	var audience domain.Audience
	if err := c.ShouldBindJSON(&audience); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		//TODO repo usage here; Title as Identficator
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "created"})
}

func (s *AudienceService) GetAudiences(c *gin.Context) {
	audiences, err := s.Repo.GetAudiences()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve audiences"})
		return
	}
	c.JSON(http.StatusOK, audiences)
}

func (s *AudienceService) UpdateAudiencesDaily(ctx context.Context) error {
    // Находим все существующие аудитории
    query := `SELECT id, 
                     creation_date_from, 
                     creation_date_to, 
                     statuses, 
                     rejection_reasons, 
                     non_target_reasons 
              FROM audiences`
    
    var audiences []domain.Audience

    err := s.Repo.DB.SelectContext(ctx, &audiences, query)
    if err != nil {
        return fmt.Errorf("failed to fetch audiences: %w", err)
    }

    // Обновляем каждую аудиторию
    for _, aud := range audiences {
		
        err := s.Repo.UpdateAudienceRequests(ctx, aud.ID)
        if err != nil {
            log.Printf("Failed to update audience %v: %v", aud.ID, err)
            // Продолжаем обновление остальных аудиторий
            continue
        }
    }

    return nil
}


func (s *AudienceService) ExportAudience(c *gin.Context) {
	id := c.Param("id")
	// TODO: вызов репа чтобы табличку отдал
	c.JSON(http.StatusOK, gin.H{"status": "exported", "id": id})
}

func (s *AudienceService) PublishUploadMsg(IntegrationID int) error {
	return s.Broker.Channel.Publish(
		"",               // Exchange
		"UploadMessages", // Routing key
		false,            // Mandatory
		false,            // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(strconv.Itoa(IntegrationID)),
		},
	)
}

func (s *AudienceService) GetIntegrations(c *gin.Context)  {
	//TODO: implement this func
}


func (s *AudienceService) CreateIntegration(c *gin.Context) {
	var integration domain.Integration
	if err := c.ShouldBindJSON(&integration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.Repo.CreateIntegration(&integration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create integration"})
		return
	}

	// Публикация задачи в очередь
	if err := s.PublishUploadMsg(integration.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish upload message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "created", "integration_id": integration.ID})
}

func (s *AudienceService) DeleteIntegration(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	if err := s.Repo.DeleteIntegration(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
