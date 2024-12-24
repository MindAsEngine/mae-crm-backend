package audience

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"reporting-service/internal/domain"
	MysqlRepo "reporting-service/internal/repository/mysql"
	PostgreRepo "reporting-service/internal/repository/postgre"
)

type ExcelExporter struct {
	audienceRepo PostgreRepo.PostgresAudienceRepository
	mysqlRepo    MysqlRepo.MySQLAudienceRepository
	logger       *zap.Logger
}

func NewExcelExporter(repo PostgreRepo.PostgresAudienceRepository, mysqlrepo MysqlRepo.MySQLAudienceRepository, logger *zap.Logger) *ExcelExporter {
	return &ExcelExporter{
		audienceRepo: repo,
		mysqlRepo:    mysqlrepo,
		logger:       logger,
	}
}

func (e *ExcelExporter) ExportAudience(ctx context.Context, audienceID int64) (string, string, error) {

	// Get full audience data with applications
	audience, err := e.audienceRepo.GetByID(ctx, audienceID)
	if err != nil {
		return "", "", fmt.Errorf("get audience: %w", err)
	}

	// Get filter data
	filter, err := e.audienceRepo.GetFilterByAudienceId(ctx, audienceID)
	if err != nil {
		return "", "", fmt.Errorf("get filter: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	// Create Applications sheet
	mainSheet := "Applications"
	f.SetSheetName("Sheet1", mainSheet)

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
		return "", "", fmt.Errorf("EXPORTER get application ids: %w", err)
	}

	applications, err := e.mysqlRepo.ListApplicationsByIds(ctx, application_ids)

	if applications == nil {
		return "", "", fmt.Errorf("EXPORTER applications not found")
	}

	if err != nil {
		return "", "", fmt.Errorf("EXPORTER get applications: %w", err)
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
	dirName := fmt.Sprintf("AUDIENCE_%s_EXPORTS", audience.Name)
	exportPath := filepath.Join("export", dirName)

	// Ensure exports directory exists
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", "", fmt.Errorf("create export directory: %w", err)
	}

	// Create filename with timestamp
	fileName := fmt.Sprintf("audience_export_%s.xlsx",
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(exportPath, fileName)

	// Ensure exports directory exists
	if err := os.MkdirAll("export", 0755); err != nil {
		return "", "", fmt.Errorf("create exports directory: %w", err)
	}

	// Save file
	if err := f.SaveAs(filePath); err != nil {
		return "", "", fmt.Errorf("save file: %w", err)
	}

	return filePath, fileName, nil
}

func (e *ExcelExporter) ExportApplications(ctx context.Context, filter *domain.ApplicationFilterRequest) (string, string, error) {
	applications, err := e.mysqlRepo.ExportApplicationsWithFilters(ctx, filter)
	if err != nil {
		return "", "", fmt.Errorf("get applications: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	// Create sheet
	sheetName := "Applications"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{
		"ID", "Дата создания", "ФИО клиента", "Статус",
		"Телефон", "Менеджер", "Тип недвижимости",
		"Дней в статусе", "Проект",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// Style headers
	style, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#CCCCCC"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return "", "", fmt.Errorf("create style: %w", err)
	}
	f.SetRowStyle(sheetName, 1, 1, style)

	// Add data
	for i, app := range applications {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), app.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), app.CreatedAt.Format("02.01.2006 15:04"))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), app.ClientName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), app.StatusName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), app.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), app.ManagerName)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), app.PropertyType)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), app.StatusDuration)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), app.ProjectName)
	}

	// Set column widths
	columnWidths := map[string]float64{
		"A": 10, "B": 20, "C": 30, "D": 20,
		"E": 15, "F": 25, "G": 20, "H": 15,
		"I": 30,
	}
	for col, width := range columnWidths {
		f.SetColWidth(sheetName, col, col, width)
	}


    // Create unique directory
	dirName := "APPLICATIONS_EXPORTS"
	exportPath := filepath.Join("export", dirName)

	// Ensure exports directory exists
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", "", fmt.Errorf("create export directory: %w", err)
	}

	// Create filename with timestamp
	fileName := fmt.Sprintf("applications_export_%s.xlsx",
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(exportPath, fileName)

	// Ensure exports directory exists
	if err := os.MkdirAll("export", 0755); err != nil {
		return "", "", fmt.Errorf("create exports directory: %w", err)
	}

	// Save file
	if err := f.SaveAs(filePath); err != nil {
		return "", "", fmt.Errorf("save file: %w", err)
	}

	return filePath, fileName, nil
}

func (e *ExcelExporter) ExportCallCenterReport(report *domain.CallCenterReport) (string, string, error) {
    f := excelize.NewFile()
    defer f.Close()

    sheetName := "Sales Report"
    f.SetSheetName("Sheet1", sheetName)

    // Set headers style
    headerStyle, err := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#CCCCCC"}},
        Alignment: &excelize.Alignment{Horizontal: "center"},
    })
    if err != nil {
        return "", "", fmt.Errorf("create header style: %w", err)
    }

    // Set percent style
    percentStyle, err := f.NewStyle(&excelize.Style{
		NumFmt:    10,
        Alignment: &excelize.Alignment{Horizontal: "right"},
    })
    if err != nil {
        return "", "", fmt.Errorf("create percent style: %w", err)
    }

    // Write headers
    for i, header := range report.Headers {
        col := string(rune('A' + i))
        cell := fmt.Sprintf("%s1", col)
        f.SetCellValue(sheetName, cell, header.Title)
        f.SetColWidth(sheetName, col, col, 15)
    }
    f.SetRowStyle(sheetName, 1, 1, headerStyle)

    // Write data
    for i, data := range report.Data {
        row := i + 2
        metrics := []interface{}{
            data.ManagerName,
            data.TotalInquiries,
            data.TargetInquiries,
            data.TargetConversion / 100,
            data.AppointedVisits,
            data.VisitConversion / 100,
            data.CompletedVisits,
            data.VisitSuccess / 100,
            data.LeadToVisit / 100,
        }

        //if len(report.Headers) > 9 { // If optional columns are included
        //    metrics = append(metrics,
        //        data.Bookings,
        //        data.VisitToBooking / 100,
        //        data.Contracts,
        //        data.BookingToContract / 100,
        //        data.LeadToContract / 100,
        //    )
        //}

        for j, value := range metrics {
            cell := fmt.Sprintf("%s%d", string(rune('A'+j)), row)
            f.SetCellValue(sheetName, cell, value)
            
            // Apply percent style to conversion columns
            if j == 3 || j == 5 || j == 7 || j == 8 || j == 10 || j == 12 || j == 13 {
                f.SetCellStyle(sheetName, cell, cell, percentStyle)
            }
        }
    }

    // Write totals
    totalRow := len(report.Data) + 2
    totalStyle, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E0E0E0"}},
    })

    totals := []interface{}{
        "ИТОГО",
        report.Totals.TotalInquiries,
        report.Totals.TargetInquiries,
        report.Totals.TargetConversion / 100,
        report.Totals.AppointedVisits,
        report.Totals.VisitConversion / 100,
        report.Totals.CompletedVisits,
        report.Totals.VisitSuccess / 100,
        report.Totals.LeadToVisit / 100,
    }

    //if len(report.Headers) > 9 {
    //    totals = append(totals,
    //        report.Totals.Bookings,
    //        report.Totals.VisitToBooking / 100,
    //        report.Totals.Contracts,
    //        report.Totals.BookingToContract / 100,
    //        report.Totals.LeadToContract / 100,
    //    )
    //}

    for j, value := range totals {
        cell := fmt.Sprintf("%s%d", string(rune('A'+j)), totalRow)
        f.SetCellValue(sheetName, cell, value)
        f.SetCellStyle(sheetName, cell, cell, totalStyle)
        
        if j == 3 || j == 5 || j == 7 || j == 8 || j == 10 || j == 12 || j == 13 {
            f.SetCellStyle(sheetName, cell, cell, percentStyle)
        }
    }

    // Add anomalies sheet if present
    if len(report.Anomalies) > 0 {
        f.NewSheet("Anomalies")
        f.SetCellValue("Anomalies", "A1", "Обнаруженные аномалии")
        f.SetCellStyle("Anomalies", "A1", "A1", headerStyle)
        
        for i, anomaly := range report.Anomalies {
            f.SetCellValue("Anomalies", fmt.Sprintf("A%d", i+2), anomaly)
        }
        
        f.SetColWidth("Anomalies", "A", "A", 50)
    }

    // Save file
    fileName := fmt.Sprintf("sales_report_%s.xlsx", 
        time.Now().Format("20060102_150405"))
    filePath := filepath.Join("export/CALL_CENTER_REPORTS", fileName)

    if err := os.MkdirAll("export/CALL_CENTER_REPORTS", 0755); err != nil {
        return "", "", fmt.Errorf("create exports directory: %w", err)
    }

    if err := f.SaveAs(filePath); err != nil {
        return "", "", fmt.Errorf("save excel file: %w", err)
    }

    return filePath, fileName, nil
}