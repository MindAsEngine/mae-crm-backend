package domain

type Task struct {
    ID         int    `json:"id"`
    LeadID     int    `json:"lead_id"`
    EmployeeID int    `json:"employee_id"`
    Title      string `json:"title"`
    Status     string `json:"status"`
    CreatedAt  string `json:"created_at"`
}

type TaskRequest struct {
    LeadIDs    []int `json:"lead_ids"`    // Идентификаторы заявок
    EmployeeID int   `json:"employee_id"` // Ответственный сотрудник
}
