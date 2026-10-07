package instance

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ozzenerol/mc-1.7.3-orchestrator/internal/db"
	"github.com/ozzenerol/mc-1.7.3-orchestrator/internal/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
  mux.HandleFunc("POST /instances", h.create)
	mux.HandleFunc("GET /instances", h.list)
  mux.HandleFunc("GET /instances/{id}", h.get)
  mux.HandleFunc("PUT /instances/{id}", h.update)
  mux.HandleFunc("DELETE /instances/{id}", h.delete)
}

type instanceCreateRequest struct {
	HostID         pgtype.UUID
	Port           int32
	MaxPlayers     int32
	DedicatedRamMb int32
	LogPath        string
}

type instanceUpdateRequest struct {
	ID             pgtype.UUID
	HostID         pgtype.UUID
	Port           int32
	MaxPlayers     int32
	DedicatedRamMb int32
	LogPath        string
}

type instanceResponse struct {
	ID             pgtype.UUID
	HostID         pgtype.UUID
	Port           int32
	MaxPlayers     int32
	DedicatedRamMb int32
	LogPath        string
}

func toResponse(ins db.Instance) instanceResponse {
	return instanceResponse {
		ID: ins.ID,
		HostID: ins.HostID,
		Port: ins.Port,
		MaxPlayers: ins.MaxPlayers,
		DedicatedRamMb: ins.DedicatedRamMb,
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req instanceCreateRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	params := db.CreateInstanceParams {
		HostID:         req.HostID,
    Port:           req.Port,
    MaxPlayers:     req.MaxPlayers,
    DedicatedRamMb: req.DedicatedRamMb,
    LogPath:        req.LogPath,
	}

	ins, err := h.svc.Create(r.Context(), params)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, toResponse(ins))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	instances, err := h.svc.List(r.Context())
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	resp := make([]instanceResponse, len(instances))
	for i, ins := range instances {
		resp[i] = toResponse(ins)
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ParseID(r)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	ins, err := h.svc.Get(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, toResponse(ins))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := utils.ParseID(r)
  if err != nil {
    utils.WriteError(w, err)
    return
  }
  var req instanceUpdateRequest
  if err := utils.DecodeJSON(r, &req); err != nil {
    utils.WriteError(w, err)
    return
  }

	params := db.UpdateInstanceParams {
		ID:							req.ID,
		HostID:         req.HostID,
    Port:           req.Port,
    MaxPlayers:     req.MaxPlayers,
    DedicatedRamMb: req.DedicatedRamMb,
    LogPath:        req.LogPath,
	}


  ins, err := h.svc.Update(r.Context(), id, params)
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  utils.WriteJSON(w, http.StatusOK, toResponse(ins))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
  id, err := utils.ParseID(r)
  if err != nil {
    utils.WriteError(w, err)
    return
  }
	
  if err := h.svc.Delete(r.Context(), id); err != nil {
    utils.WriteError(w, err)
    return
  }

  w.WriteHeader(http.StatusNoContent)
}
