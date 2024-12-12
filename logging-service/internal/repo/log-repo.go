package repo

import (
    "context"
    "log"
    "encoding/json"
    "logging-service/internal/domain"
    "strings"
    "github.com/elastic/go-elasticsearch/v8"
    "github.com/elastic/go-elasticsearch/v8/esapi"
)

type LogRepo struct {
    Client *elasticsearch.Client
}

// Инициализация Elasticsearch
func InitElasticsearch() *elasticsearch.Client {
    cfg := elasticsearch.Config{
        Addresses: []string{
            "http://elasticsearch:8088",
        },
    }

    es, err := elasticsearch.NewClient(cfg)
    if err != nil {
        log.Fatalf("Error creating Elasticsearch client: %v", err)
    }

    return es
}

// Поиск логов с фильтрацией
func (r *LogRepo) SearchLogs(query string) ([]domain.LogEntry, error) {
    var logs []domain.LogEntry

    req := esapi.SearchRequest{
        Index: []string{"logs"},
        Body:  strings.NewReader(query),
    }

    res, err := req.Do(context.Background(), r.Client)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    // Обработка ответа Elasticsearch
    if err := json.NewDecoder(res.Body).Decode(&logs); err != nil {
        return nil, err
    }

    return logs, nil
}
