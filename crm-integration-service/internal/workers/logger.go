package workers

import (
    "log"
    "time"
)

type Loggable interface {
    LogTaskResult(taskID int, result string) error
}

func StartLogger(service Loggable) {
    for {
        time.Sleep(10 * time.Second)
        log.Println("Background logging worker running...")
        // Здесь можно добавить обработку дополнительных логов
    }
}
