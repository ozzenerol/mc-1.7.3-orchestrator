package utils

import (
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net/http"

  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgconn"
  "github.com/jackc/pgx/v5/pgtype"
)

var (
  ErrNotFound     = errors.New("not found")
  ErrInvalidInput = errors.New("invalid input")
  ErrConflict     = errors.New("already exists")
)

const (
  pgForeignKeyViolation = "23503"
  pgUniqueViolation     = "23505"
)

func NotFound(what string) error {
  return fmt.Errorf("%s %w", what, ErrNotFound)
}

func InvalidInput(format string, args ...any) error {
  return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

func DBError(err error, what string) error {
  var pgErr *pgconn.PgError
  switch {
  case errors.Is(err, pgx.ErrNoRows):
    return NotFound(what)
  case errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation:
    return InvalidInput("%s", pgErr.Detail)
  case errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation:
    return fmt.Errorf("%s %w", what, ErrConflict)
  }
  return err
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, err error) {
  switch {
  case errors.Is(err, ErrNotFound):
    http.Error(w, err.Error(), http.StatusNotFound)
  case errors.Is(err, ErrInvalidInput):
    http.Error(w, err.Error(), http.StatusBadRequest)
  case errors.Is(err, ErrConflict):
    http.Error(w, err.Error(), http.StatusConflict)
  default:
    log.Printf("internal error: %v", err)
    http.Error(w, "internal error", http.StatusInternalServerError)
  }
}

func DecodeJSON(r *http.Request, v any) error {
  if err := json.NewDecoder(r.Body).Decode(v); err != nil {
    return InvalidInput("invalid json")
  }
  return nil
}

func ParseID(r *http.Request) (pgtype.UUID, error) {
  var id pgtype.UUID
  if err := id.Scan(r.PathValue("id")); err != nil {
    return id, InvalidInput("invalid id %q", r.PathValue("id"))
  }
  return id, nil
}
