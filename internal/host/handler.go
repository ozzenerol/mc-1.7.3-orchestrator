package host

import (
  "net/http"
  "time"

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
  mux.HandleFunc("POST /hosts", h.create)
  mux.HandleFunc("GET /hosts", h.list)
  mux.HandleFunc("GET /hosts/{id}", h.get)
  mux.HandleFunc("PUT /hosts/{id}", h.update)
  mux.HandleFunc("DELETE /hosts/{id}", h.delete)
}

type hostRequest struct {
  Name string `json:"name"`
  IP   string `json:"ip"`
}

type hostResponse struct {
  ID        string    `json:"id"`
  Name      string    `json:"name"`
  IP        string    `json:"ip"`
  CreatedAt time.Time `json:"created_at"`
}

func toResponse(h db.Host) hostResponse {
  return hostResponse{
    ID:        h.ID.String(),
    Name:      h.Name,
    IP:        h.Ip,
    CreatedAt: h.CreatedAt.Time,
  }
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var req hostRequest
  if err := utils.DecodeJSON(r, &req); err != nil {
    utils.WriteError(w, err)
    return
  }

  host, err := h.svc.Create(r.Context(), req.Name, req.IP)
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  utils.WriteJSON(w, http.StatusCreated, toResponse(host))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
  hosts, err := h.svc.List(r.Context())
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  resp := make([]hostResponse, len(hosts))
  for i, host := range hosts {
    resp[i] = toResponse(host)
  }

  utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := utils.ParseID(r)
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  host, err := h.svc.Get(r.Context(), id)
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  utils.WriteJSON(w, http.StatusOK, toResponse(host))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := utils.ParseID(r)
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  var req hostRequest
  if err := utils.DecodeJSON(r, &req); err != nil {
    utils.WriteError(w, err)
    return
  }

  host, err := h.svc.Update(r.Context(), id, req.Name, req.IP)
  if err != nil {
    utils.WriteError(w, err)
    return
  }

  utils.WriteJSON(w, http.StatusOK, toResponse(host))
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
