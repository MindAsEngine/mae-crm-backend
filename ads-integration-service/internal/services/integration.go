package services

import (
	"ads-integration-service/internal/repository"
	"ads-integration-service/internal/workers"
	"log"
	//"github.com/gin-gonic/gin"
)

type IntegrationService struct {
    Repo *repository.IntegrationRepo
    MQ   *workers.RabbitMQ
}


func NewIntegrationService(repo *repository.IntegrationRepo, mq *workers.RabbitMQ) *IntegrationService {
    return &IntegrationService{Repo: repo, MQ: mq}
}


// Реализация интерфейса TaskHandler
func (s *IntegrationService) ProcessUploadMsg(UploadMsgID int) error {
    // Пример обработки задачи
    integration, err := s.Repo.GetIntegrationByID(UploadMsgID)
    if err != nil {
        log.Printf("Failed to fetch integration UploadMsg %d: %v", UploadMsgID, err)
        return err
    }

    log.Printf("Processing UploadMsg: %+v", integration)
    // Здесь логика интеграции с рекламными кабинетами 
    // TODO: send api request to ads cabinets
    return nil
}
