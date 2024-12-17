package audience

import (
    "context"
    "fmt"
    "path/filepath"
    "strings"
    "time"

    "github.com/xuri/excelize/v2"
    "go.uber.org/zap"

    PostgreRepo "reporting-service/internal/repository/postgre"
)

type ExcelExporter struct {
    audienceRepo PostgreRepo.PostgresAudienceRepository
    logger       *zap.Logger
}

func NewExcelExporter(repo PostgreRepo.PostgresAudienceRepository, logger *zap.Logger) *ExcelExporter {
    return &ExcelExporter{
        audienceRepo: repo,
        logger:      logger,
    }
}

func (e *ExcelExporter) ExportAudience(ctx context.Context, audienceID int64) (string, error) {
    audience, err := e.audienceRepo.GetByID(ctx, audienceID)
    if err != nil {
        return "", fmt.Errorf("get audience: %w", err)
    }

    f := excelize.NewFile()
    defer f.Close()

    mainSheet := "Audience Info"
    f.NewSheet(mainSheet)
    
    headers := []string{"Request ID", "Status", "Reason", "Created At", "Updated At"}
    for i, header := range headers {
        cell := fmt.Sprintf("%s1", string(rune('A'+i)))
        f.SetCellValue(mainSheet, cell, header)
    }

    style, err := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#CCCCCC"}},
    })
    if err != nil {
        return "", fmt.Errorf("create style: %w", err)
    }
    f.SetRowStyle(mainSheet, 1, 1, style)

    for i, req := range audience.Applications {
        row := i + 2
        f.SetCellValue(mainSheet, fmt.Sprintf("A%d", row), req.ID)
        f.SetCellValue(mainSheet, fmt.Sprintf("B%d", row), req.StatusName)
        f.SetCellValue(mainSheet, fmt.Sprintf("C%d", row), req.ReasonName)
        f.SetCellValue(mainSheet, fmt.Sprintf("D%d", row), req.CreatedAt.Format(time.RFC3339))
        f.SetCellValue(mainSheet, fmt.Sprintf("E%d", row), req.UpdatedAt.Format(time.RFC3339))
    }

    filterSheet := "Filter Info"
    f.NewSheet(filterSheet)
    
    f.SetCellValue(filterSheet, "A1", "Filter Parameter")
    f.SetCellValue(filterSheet, "B1", "Value")
    
    f.SetCellValue(filterSheet, "A2", "Date From")
    if audience.Filter.CreationDateFrom != nil {
        f.SetCellValue(filterSheet, "B2", audience.Filter.CreationDateFrom.Format(time.RFC3339))
    }
    
    f.SetCellValue(filterSheet, "A3", "Date To")
    if audience.Filter.CreationDateTo != nil {
        f.SetCellValue(filterSheet, "B3", audience.Filter.CreationDateTo.Format(time.RFC3339))
    }
    
    f.SetCellValue(filterSheet, "A4", "Statuses")
    f.SetCellValue(filterSheet, "B4", strings.Join(audience.Filter.StatusNames, ", "))

    f.SetCellValue(filterSheet, "A5", "Statuses")
    f.SetCellValue(filterSheet, "B5", strings.Join(audience.Filter.RegectionReasonNames, ", "))



    f.SetColWidth(mainSheet, "A", "E", 15)
    f.SetColWidth(filterSheet, "A", "B", 20)

    fileName := fmt.Sprintf("audience_%s_%s.xlsx", 
        string(audienceID)[:8],
        time.Now().Format("2006-01-02_15-04-05"))
    
    filePath := filepath.Join("exports", fileName)
    
    if err := f.SaveAs(filePath); err != nil {
        return "", fmt.Errorf("save file: %w", err)
    }

    return filePath, nil
}