package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	//"gorm.io/gorm/logger"

	"reporting-service/internal/domain"
	"reporting-service/internal/services/audience"
)

type Response struct {
    Error   string      `json:"error,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
    audienceService *audience.Service
    logger         *zap.Logger
}

func NewHandler(audienceService *audience.Service, logger *zap.Logger) *Handler {
    return &Handler{
        audienceService: audienceService,
        logger:         logger,
    }
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
    api := r.PathPrefix("/api").Subrouter()
    
    // Audiences endpoints
    api.HandleFunc("/audiences", h.GetAudiences).Methods(http.MethodGet)
    api.HandleFunc("/audiences", h.CreateAudience).Methods(http.MethodPost)
    api.HandleFunc("/audiences/integrations", h.CreateIntegrations).Methods(http.MethodPost)
    api.HandleFunc("/audiences/{audienceId}", h.GetAudience).Methods(http.MethodGet)
    api.HandleFunc("/audiences/{audienceId}", h.DeleteAudience).Methods(http.MethodDelete)
    api.HandleFunc("/audiences/{audienceId}/disconnect", h.DisconnectAudience).Methods(http.MethodDelete)
    api.HandleFunc("/audiences/{audienceId}/export", h.ExportAudience).Methods(http.MethodGet)
}

func (h *Handler) GetAudiences(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    audiences, err := h.audienceService.List(ctx)
    if err != nil {
		log.Print("failed to get audiences","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
        h.errorResponse(w, "failed to get audiences", err, http.StatusInternalServerError)
        return
    }

    h.jsonResponse(w, audiences, http.StatusOK)
}

func (h *Handler) GetAudience(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)

    audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
    if err != nil {
        h.errorResponse(w, "invalid audience id", err, http.StatusBadRequest)
        return
    }
    audiences, err := h.audienceService.GetById(ctx, audienceID)
    if err != nil {
		log.Print("failed to get audiences","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
        h.errorResponse(w, "failed to get audiences", err, http.StatusInternalServerError)
        return
    }

    h.jsonResponse(w, audiences, http.StatusOK)
}

func (h *Handler) CreateIntegrations(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req domain.IntegrationsCreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", err, http.StatusBadRequest)
        return
    }

    integrations, err := h.audienceService.CreateIntegrations(ctx, req)
    if err != nil {
		log.Print("failed to create audience","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
        h.errorResponse(w, "failed to create audience", err, http.StatusInternalServerError)
        return
    }
    h.jsonResponse(w, integrations, http.StatusCreated)
}

func (h *Handler) CreateAudience(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req domain.AudienceCreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", err, http.StatusBadRequest)
        return
    }

    audience, err := h.audienceService.Create(ctx, req)
    if err != nil {
		log.Print("failed to create audience","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
        h.errorResponse(w, "failed to create audience", err, http.StatusInternalServerError)
        return
    }

    h.jsonResponse(w, audience, http.StatusCreated)
}

func (h *Handler) DeleteAudience(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    
    audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
    if err != nil {
        h.errorResponse(w, "invalid audience id", err, http.StatusBadRequest)
        return
    }

    if err := h.audienceService.Delete(ctx, audienceID); err != nil {
		log.Print("failed to delete audience","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
		h.errorResponse(w, "failed to delete audience", err, http.StatusInternalServerError)
        return
    }

    h.jsonResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *Handler) DisconnectAudience(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    
    audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
    if err != nil {
        h.errorResponse(w, "invalid audience id", err, http.StatusBadRequest)
        return
    }

    if err := h.audienceService.DisconnectAll(ctx, audienceID); err != nil {
		log.Print("failed to disconnect audience","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
		h.errorResponse(w, "failed to disconnect audience", err, http.StatusInternalServerError)
        return
    }

    h.jsonResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *Handler) ExportAudience(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    
    audienceID, err := strconv.ParseInt(vars["audienceId"], 10, 64)
    if err != nil {
        h.errorResponse(w, "invalid audience id", err, http.StatusBadRequest)
        return
    }

    filePath, err := h.audienceService.Export(ctx, audienceID)
    if err != nil {
		log.Print("failed to export audience","\nRequest: ",r,"\nResponce: ",w,"\nError: ",err)
		h.errorResponse(w, "failed to export audience", err, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", "attachment; filename=audience_export.xlsx")
    http.ServeFile(w, r, filePath)
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