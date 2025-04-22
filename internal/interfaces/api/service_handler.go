package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"api-gateway-sample/internal/application/dto"
	"api-gateway-sample/pkg/errors"
)

// ServiceHandler handles HTTP requests for service management
type ServiceHandler struct {
	serviceUseCase ServiceUseCase
}

// NewServiceHandler creates a new ServiceHandler instance
func NewServiceHandler(serviceUseCase ServiceUseCase) *ServiceHandler {
	return &ServiceHandler{
		serviceUseCase: serviceUseCase,
	}
}

// RegisterRoutes registers the service routes
func (h *ServiceHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/services", h.CreateService).Methods(http.MethodPost)
	router.HandleFunc("/services", h.ListServices).Methods(http.MethodGet)
	router.HandleFunc("/services/{id}", h.GetService).Methods(http.MethodGet)
	router.HandleFunc("/services/{id}", h.UpdateService).Methods(http.MethodPut)
	router.HandleFunc("/services/{id}", h.DeleteService).Methods(http.MethodDelete)
	router.HandleFunc("/services/name/{name}", h.FindServiceByName).Methods(http.MethodGet)
}

// CreateService handles service creation requests
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	service, err := h.serviceUseCase.CreateService(r.Context(), &req)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			http.Error(w, "Service already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(service)
}

// GetService handles service retrieval requests
func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	service, err := h.serviceUseCase.GetService(r.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// UpdateService handles service update requests
func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	service, err := h.serviceUseCase.UpdateService(r.Context(), id, &req)
	if err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		if errors.IsAlreadyExists(err) {
			http.Error(w, "Service name already taken", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// DeleteService handles service deletion requests
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.serviceUseCase.DeleteService(r.Context(), id); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListServices handles service listing requests
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.serviceUseCase.ListServices(r.Context())
	if err != nil {
		http.Error(w, "Failed to list services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// FindServiceByName handles service lookup by name
func (h *ServiceHandler) FindServiceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	service, err := h.serviceUseCase.FindServiceByName(r.Context(), name)
	if err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}
