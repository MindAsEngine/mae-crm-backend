package api

import (
	"encoding/json"
	//"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	//"gorm.io/gorm/logger"

	"reporting-service/internal/domain"
	"reporting-service/internal/services/audience"
)

type Response struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

type Handler struct {
	audienceService *audience.Service
	logger          *zap.Logger
}

func NewHandler(audienceService *audience.Service, logger *zap.Logger) *Handler {
	return &Handler{
		audienceService: audienceService,
		logger:          logger,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)
	// Audiences endpoints
	api.HandleFunc("/audiences", h.GetAudiences).Methods(http.MethodGet)
	api.HandleFunc("/audiences", h.CreateAudience).Methods(http.MethodPost)
	api.HandleFunc("/audiences/integrations", h.CreateIntegrations).Methods(http.MethodPost)
	api.HandleFunc("/audiences/{audienceId}", h.GetAudience).Methods(http.MethodGet)
	api.HandleFunc("/audiences/{audienceId}", h.DeleteAudience).Methods(http.MethodDelete)
	api.HandleFunc("/audiences/{audienceId}/disconnect", h.DisconnectAudience).Methods(http.MethodDelete)
	api.HandleFunc("/audiences/{audienceId}/export", h.ExportAudience).Methods(http.MethodGet)
	
	api.HandleFunc("/applications/filters", h.GetAudienceFilters).Methods(http.MethodGet)
	api.HandleFunc("/applications", h.ListApplications).Methods(http.MethodGet)
	api.HandleFunc("/applications/export", h.ExportApplications).Methods(http.MethodGet)
	
	api.HandleFunc("/regions", h.GetRegions).Methods(http.MethodGet)
	api.HandleFunc("/regions/export", h.GetRegions).Methods(http.MethodGet)
	
	api.HandleFunc("/call-center", h.GetCallCenterReport).Methods(http.MethodGet)
	api.HandleFunc("/call-center/export", h.ExportCallCenterReport).Methods(http.MethodGet)

	api.HandleFunc("/speed", h.GetStatusDurationReport).Methods(http.MethodGet)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func (h *Handler) GetAudienceFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters, err := h.audienceService.GetFilters(ctx)
	if err != nil {
		h.errorResponse(w, "failed to get audience filters: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, filters, http.StatusOK)
}

func (h *Handler) GetAudiences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	audiences, err := h.audienceService.AudienceList(ctx)
	if err != nil {
		h.errorResponse(w, "failed to get audiences: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, audiences, http.StatusOK)
}

func (h *Handler) GetAudience(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
	if err != nil {
		h.errorResponse(w, "invalid audience id: "+err.Error(), err, http.StatusBadRequest)
		return
	}
	audiences, err := h.audienceService.GetById(ctx, audienceID)
	if err != nil {
		h.errorResponse(w, "failed to get audience by id: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, audiences, http.StatusOK)
}

func (h *Handler) CreateIntegrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.IntegrationsCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body: "+err.Error(), err, http.StatusBadRequest)
		return
	}

	integrations, err := h.audienceService.CreateIntegrations(ctx, req)
	if err != nil {
		h.errorResponse(w, "failed to create integration: "+err.Error(), err, http.StatusInternalServerError)
		return
	}
	h.jsonResponse(w, integrations, http.StatusCreated)
}

func (h *Handler) CreateAudience(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.AudienceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body: "+err.Error(), err, http.StatusBadRequest)
		return
	}

	audience, err := h.audienceService.Create(ctx, req)
	if err != nil {
		h.errorResponse(w, "failed to create audience: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, audience, http.StatusCreated)
}

func (h *Handler) DeleteAudience(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
	if err != nil {
		h.errorResponse(w, "invalid audience id: "+err.Error(), err, http.StatusBadRequest)
		return
	}

	if err := h.audienceService.Delete(ctx, audienceID); err != nil {
		h.errorResponse(w, "failed to delete audience: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *Handler) DisconnectAudience(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
	if err != nil {
		h.errorResponse(w, "invalid audience id: "+err.Error(), err, http.StatusBadRequest)
		return
	}

	if err := h.audienceService.DisconnectAll(ctx, audienceID); err != nil {
		h.errorResponse(w, "failed to disconnect audience: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *Handler) ExportAudience(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
	if err != nil {
		h.errorResponse(w, "invalid audience id: "+err.Error(), err, http.StatusBadRequest)
		return
	}

	filePath, fileName, err := h.audienceService.ExportAudience(ctx, audienceID)
	if err != nil {
		h.errorResponse(w, "failed to export audience: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filePath)
}

func (h *Handler) ListApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")

	pagination := &domain.PaginationRequest{}

	if page != "" {
		pageNum, err := strconv.Atoi(page)
		if err != nil || pageNum < 1 {
			h.errorResponse(w, "invalid page number", err, http.StatusBadRequest)
			return
		}
		pagination.Page = pageNum
	}

	if pageSize != "" {
		size, err := strconv.Atoi(pageSize)
		if err != nil || size < 1 {
			h.errorResponse(w, "invalid page size", err, http.StatusBadRequest)
			return
		}
		pagination.PageSize = size
	}

	time_from := time.Time{}
	time_to := time.Time{}
    err:= error(nil)
	if r.URL.Query().Get("start_date") != "" {
		time_from, err = time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
		if err != nil  {
			h.errorResponse(w, "invalid date_from format", err, http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("end_date") != "" {
		time_to, err = time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
		if err != nil {
			h.errorResponse(w, "invalid date_to format", err, http.StatusBadRequest)
			return
		}
	}
	filter := &domain.ApplicationFilterRequest{
		OrderField:     r.URL.Query().Get("order_field"),
		OrderDirection: r.URL.Query().Get("order_direction"),
		Status:         r.URL.Query().Get("status"),
		ProjectName:    r.URL.Query().Get("project_name"),
		PropertyType:   r.URL.Query().Get("property_type"),
	}
	if !time_from.IsZero() && !time_to.IsZero() {
		filter.StartDate = &time_from
		filter.EndDate = &time_to
	} 

	if daysInStatus := r.URL.Query().Get("days_in_status"); daysInStatus != "" {
		if days, err := strconv.Atoi(daysInStatus); err == nil {
			filter.StatusDuration = days
		}
	}

	response, err := h.audienceService.ListApplications(ctx, pagination, filter)
	if err != nil {
		h.errorResponse(w, "failed to get applications", err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, response, http.StatusOK)
}

func (h *Handler) ExportApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// vars := mux.Vars(r)
	time_from := time.Time{}
	time_to := time.Time{}
    err:= error(nil)
	if r.URL.Query().Get("start_date") != "" {
		time_from, err = time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
		if err != nil {
			h.errorResponse(w, "invalid date_from format", err, http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("end_date") != "" {
		time_to, err = time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
		if err != nil {
			h.errorResponse(w, "invalid date_to format", err, http.StatusBadRequest)
			return
		}
	}
	// Parse filter parameters
	filter := &domain.ApplicationFilterRequest{
		OrderField:     r.URL.Query().Get("order_field"),
		OrderDirection: r.URL.Query().Get("order_direction"),
		Status:         r.URL.Query().Get("status"),
		ProjectName:    r.URL.Query().Get("project_name"),
		PropertyType:   r.URL.Query().Get("property_type"),
	}
	if !time_from.IsZero() && !time_to.IsZero() {
		filter.StartDate = &time_from
		filter.EndDate = &time_to
	} 


	// Parse days_in_status if provided
	if daysStr := r.URL.Query().Get("days_in_status"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil {
			filter.StatusDuration = days
		} else {
			h.logger.Warn("invalid days_in_status parameter", zap.Error(err))
		}
	}

	filePath, fileName, err := h.audienceService.ExportApplications(ctx, *filter)
	if err != nil {
		h.errorResponse(w, "failed to export applications: "+err.Error(), err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filePath)
}

func (h *Handler) GetRegions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := &domain.RegionFilter{
		Search: r.URL.Query().Get("search"),
		Sort:   r.URL.Query().Get("sort"),
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		date, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			h.errorResponse(w, "invalid start date format", err, http.StatusBadRequest)
			return
		}
		filter.StartDate = &date
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		date, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			h.errorResponse(w, "invalid end date format", err, http.StatusBadRequest)
			return
		}
		filter.EndDate = &date
	}

	response, err := h.audienceService.GetRegions(ctx, filter)
	if err != nil {
		h.errorResponse(w, "failed to get regions data", err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, response, http.StatusOK)
}

func (h *Handler) GetCallCenterReport(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

	time_from := time.Now().AddDate(-1, 0, 0)
	time_to := time.Now()
    err:= error(nil)
	if r.URL.Query().Get("start_date") != "" {
		time_from, err = time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
		if err != nil  {
			h.errorResponse(w, "invalid date_from format", err, http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("end_date") != "" {
		time_to, err = time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
		if err != nil {
			h.errorResponse(w, "invalid date_to format", err, http.StatusBadRequest)
			return
		}
	}
	filter := &domain.CallCenterReportFilter{
		StartDate: &time_from,
		EndDate:   &time_to,}

    // Get report from service
    report, err := h.audienceService.GetCallCenterReport(ctx, filter)
    if err != nil {
        h.errorResponse(w, "failed to get sales report", err, http.StatusInternalServerError)
        return
    }

    // Return JSON by default
    h.jsonResponse(w, report, http.StatusOK)
}

func (h *Handler) ExportCallCenterReport(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

	time_from := time.Time{}
	time_to := time.Time{}
    err:= error(nil)
	if r.URL.Query().Get("start_date") != "" {
		time_from, err = time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
		if err != nil  {
			h.errorResponse(w, "invalid date_from format", err, http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("end_date") != "" {
		time_to, err = time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
		if err != nil {
			h.errorResponse(w, "invalid date_to format", err, http.StatusBadRequest)
			return
		}
	}
	filter := &domain.CallCenterReportFilter{
		StartDate: &time_from,
		EndDate:   &time_to,}


    // Get exported file path
    filePath, fileName, err := h.audienceService.ExportCallCenterReport(ctx, filter)
    if err != nil {
        h.errorResponse(w, "failed to export sales report", err, http.StatusInternalServerError)
        return
    }

    // Get filename from path


    // Set headers for file download
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filePath)
}

func (h *Handler) GetStatusDurationReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	time_from := time.Time{}
	time_to := time.Time{}
    err:= error(nil)
	if r.URL.Query().Get("start_date") != "" {
		time_from, err = time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
		if err != nil  {
			h.errorResponse(w, "invalid date_from format", err, http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Get("end_date") != "" {
		time_to, err = time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
		if err != nil {
			h.errorResponse(w, "invalid date_to format", err, http.StatusBadRequest)
			return
		}
	}
	trashold := 0
	if r.URL.Query().Get("over_threshold") != "" {
		trashold, err = strconv.Atoi(r.URL.Query().Get("over_threshold"))
		if err != nil {
			h.errorResponse(w, "invalid over_threshold format", err, http.StatusBadRequest)
			return
		}
	}

	filter := &domain.StatusDurationFilter{
		StartDate: &time_from,
		EndDate:   &time_to,
		ThresholdDays: trashold,
	}

	response, err := h.audienceService.GetSpeedReport(ctx, filter)
	if err != nil {
		h.errorResponse(w, "failed to get status duration report", err, http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, response, http.StatusOK)
}

func (h *Handler) errorResponse(w http.ResponseWriter, message string, err error, code int) {
	h.logger.Error(message,
		zap.Error(err),
		zap.Int("status_code", code))

	h.jsonResponse(w, Response{
		Error: message,
	}, code)
}

func (h *Handler) jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response",
			zap.Error(err))
	}
}
