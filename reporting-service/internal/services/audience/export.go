package audience

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	MysqlRepo "reporting-service/internal/repository/mysql"
	PostgreRepo "reporting-service/internal/repository/postgre"
)

type ExcelExporter struct {
    audienceRepo PostgreRepo.PostgresAudienceRepository
    mysqlRepo    MysqlRepo.MySQLAudienceRepository
    logger       *zap.Logger
}

func NewExcelExporter(repo PostgreRepo.PostgresAudienceRepository, mysqlrepo MysqlRepo.MySQLAudienceRepository , logger *zap.Logger) *ExcelExporter {
    return &ExcelExporter{
        audienceRepo: repo,
        mysqlRepo: mysqlrepo,
        logger:      logger,
    }
}

func (e *ExcelExporter) ExportAudience(ctx context.Context, audienceID int64) (string, error) {

    // Get full audience data with applications
    audience, err := e.audienceRepo.GetByID(ctx, audienceID)
    if err != nil {
        return "", fmt.Errorf("get audience: %w", err)
    }

    // Get filter data
    filter, err := e.audienceRepo.GetFilterByAudienceId(ctx, audienceID)
    if err != nil {
        return "", fmt.Errorf("get filter: %w", err)
    }

    f := excelize.NewFile()
    defer f.Close()

    // Create Applications sheet
    mainSheet := "Applications"
    f.NewSheet(mainSheet)
    
    // Set headers for applications
    headers := []string{
        "ID", "Status", "Status Name", "Reason", 
        "Manager ID", "Client ID", "Created At", "Updated At",
    }
    
    // Style headers
    for i, header := range headers {
        cell := fmt.Sprintf("%s1", string(rune('A'+i)))
        f.SetCellValue(mainSheet, cell, header)
    }

    style, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#CCCCCC"}},
    })
    f.SetRowStyle(mainSheet, 1, 1, style)

    application_ids, err := e.audienceRepo.GetApplicationIdsByAdienceId(ctx, audienceID)

    if err != nil {
        return "", fmt.Errorf("EXPORTER get application ids: %w", err)
    }

    applications, err := e.mysqlRepo.ListApplicationsByIds(ctx, application_ids)

    if err != nil {
        return "", fmt.Errorf("EXPORTER get applications: %w", err)
    }

    audience.Applications = applications
    // Add application data
    for i, app := range audience.Applications {
        row := i + 2
        f.SetCellValue(mainSheet, fmt.Sprintf("A%d", row), app.ID)
        f.SetCellValue(mainSheet, fmt.Sprintf("B%d", row), app.StatusID)
        f.SetCellValue(mainSheet, fmt.Sprintf("C%d", row), app.StatusName)
        f.SetCellValue(mainSheet, fmt.Sprintf("D%d", row), app.ReasonName)
        f.SetCellValue(mainSheet, fmt.Sprintf("E%d", row), app.ManagerID)
        f.SetCellValue(mainSheet, fmt.Sprintf("F%d", row), app.ClientID)
        f.SetCellValue(mainSheet, fmt.Sprintf("G%d", row), app.CreatedAt.Format(time.RFC3339))
        f.SetCellValue(mainSheet, fmt.Sprintf("H%d", row), app.UpdatedAt.Format(time.RFC3339))
    }

    // Create Audience Info sheet
    infoSheet := "Audience Info"
    f.NewSheet(infoSheet)
    
    // Add audience info
    f.SetCellValue(infoSheet, "A1", "Parameter")
    f.SetCellValue(infoSheet, "B1", "Value")
    
    f.SetCellValue(infoSheet, "A2", "Audience ID")
    f.SetCellValue(infoSheet, "B2", audience.ID)
    
    f.SetCellValue(infoSheet, "A3", "Name")
    f.SetCellValue(infoSheet, "B3", audience.Name)
    
    f.SetCellValue(infoSheet, "A4", "Created At")
    f.SetCellValue(infoSheet, "B4", audience.CreatedAt.Format(time.RFC3339))
    
    f.SetCellValue(infoSheet, "A5", "Updated At")
    f.SetCellValue(infoSheet, "B5", audience.UpdatedAt.Format(time.RFC3339))

    // Add filter info
    f.SetCellValue(infoSheet, "A7", "Filter Settings")
    f.SetCellValue(infoSheet, "A8", "Date From")
    if filter.CreationDateFrom != nil {
        f.SetCellValue(infoSheet, "B8", filter.CreationDateFrom.Format(time.RFC3339))
    }
    
    f.SetCellValue(infoSheet, "A9", "Date To")
    if filter.CreationDateTo != nil {
        f.SetCellValue(infoSheet, "B9", filter.CreationDateTo.Format(time.RFC3339))
    }
    
    f.SetCellValue(infoSheet, "A10", "Status Names")
    f.SetCellValue(infoSheet, "B10", strings.Join(filter.StatusNames, ", "))
    
    f.SetCellValue(infoSheet, "A11", "Rejection Reasons")
    f.SetCellValue(infoSheet, "B11", strings.Join(filter.RegectionReasonNames, ", "))
    
    f.SetCellValue(infoSheet, "A12", "Non-Target Reasons")
    f.SetCellValue(infoSheet, "B12", strings.Join(filter.NonTargetReasonNames, ", "))

    // Set column widths
    f.SetColWidth(mainSheet, "A", "H", 15)
    f.SetColWidth(infoSheet, "A", "B", 30)

     // Create unique directory
     dirName := fmt.Sprintf("AUDIENCE_%s", audience.Name)
     exportPath := filepath.Join("export", dirName)
 
     // Ensure exports directory exists
     if err := os.MkdirAll(exportPath, 0755); err != nil {
         return "", fmt.Errorf("create export directory: %w", err)
     }
 
     // Generate unique filename
     timestamp := time.Now().Format("20060102_150405")
     fileName := fmt.Sprintf("audience_%s_%s.xlsx", audience.Name, timestamp)
     filePath := filepath.Join(exportPath, fileName)
 
     // Clean up old exports (keep last 10)
     if err := e.cleanupOldExports("export", 10); err != nil {
         e.logger.Error("Failed to cleanup old exports", zap.Error(err))
     }

    // Save to unique path
    if err := f.SaveAs(filePath); err != nil {
        return "", fmt.Errorf("save file: %w", err)
    }

    return filePath, nil
}

func (e *ExcelExporter) cleanupOldExports(baseDir string, keep int) error {
    dirs, err := os.ReadDir(baseDir)
    if err != nil {
        return fmt.Errorf("read directory: %w", err)
    }

    // Sort directories by name (timestamp_audienceID)
    sort.Slice(dirs, func(i, j int) bool {
        return dirs[i].Name() > dirs[j].Name()
    })

    // Remove old directories
    for i := keep; i < len(dirs); i++ {
        path := filepath.Join(baseDir, dirs[i].Name())
        if err := os.RemoveAll(path); err != nil {
            return fmt.Errorf("remove old export: %w", err)
        }
    }

    return nil
}