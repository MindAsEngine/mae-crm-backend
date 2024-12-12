package domain

type LogEntry struct {
    Timestamp string `json:"timestamp"`
    Level     string `json:"level"`
    Service   string `json:"service"`
    Message   string `json:"message"`
}
