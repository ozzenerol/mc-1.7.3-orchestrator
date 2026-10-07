package instance

import (
  "context"
  "os"
  "path/filepath"

  "github.com/jackc/pgx/v5/pgtype"
  "github.com/ozzenerol/mc-1.7.3-orchestrator/internal/config"
  "github.com/ozzenerol/mc-1.7.3-orchestrator/internal/db"
  "github.com/ozzenerol/mc-1.7.3-orchestrator/internal/utils"
)

const entity = "instance"

type ValidateInstanceParams struct {
  HostID         pgtype.UUID
  Port           int32
  MaxPlayers     int32
  DedicatedRamMb int32
  LogPath        string
}

type Service struct {
  q   *db.Queries
  cfg config.Config
}

func NewService(q *db.Queries) *Service {
  return &Service{
    q:   q,
    cfg: config.ReadOnlyInstance(),
  }
}

func (s *Service) Create(ctx context.Context, p db.CreateInstanceParams) (db.Instance, error) {
  params := ValidateInstanceParams{
    HostID:         p.HostID,
    Port:           p.Port,
    MaxPlayers:     p.MaxPlayers,
    DedicatedRamMb: p.DedicatedRamMb,
    LogPath:        p.LogPath,
  }

  if err := s.validate(params); err != nil {
    return db.Instance{}, err
  }

  inst, err := s.q.CreateInstance(ctx, p)
  return inst, utils.DBError(err, entity)
}

func (s *Service) Get(ctx context.Context, id pgtype.UUID) (db.Instance, error) {
  inst, err := s.q.GetInstance(ctx, id)
  return inst, utils.DBError(err, entity)
}

func (s *Service) List(ctx context.Context) ([]db.Instance, error) {
  instances, err := s.q.ListInstances(ctx)
  return instances, utils.DBError(err, entity)
}

func (s *Service) Update(ctx context.Context, id pgtype.UUID, p db.UpdateInstanceParams) (db.Instance, error) {
  params := ValidateInstanceParams{
    HostID:         p.HostID,
    Port:           p.Port,
    MaxPlayers:     p.MaxPlayers,
    DedicatedRamMb: p.DedicatedRamMb,
    LogPath:        p.LogPath,
  }

  if err := s.validate(params); err != nil {
    return db.Instance{}, err
  }

  updateParams := db.UpdateInstanceParams{
    ID:             id,
    HostID:         p.HostID,
    Port:           p.Port,
    MaxPlayers:     p.MaxPlayers,
    DedicatedRamMb: p.DedicatedRamMb,
    LogPath:        p.LogPath,
  }

  inst, err := s.q.UpdateInstance(ctx, updateParams)
  return inst, utils.DBError(err, entity)
}

func (s *Service) Delete(ctx context.Context, id pgtype.UUID) error {
  n, err := s.q.DeleteInstance(ctx, id)
  if err != nil {
    return utils.DBError(err, entity)
  }

  if n == 0 {
    return utils.NotFound(entity)
  }

  return nil
}

func (s *Service) validate(p ValidateInstanceParams) error {
  if p.Port < 0 || uint32(p.Port) < s.cfg.Flags.MinPort || uint32(p.Port) > s.cfg.Flags.MaxPort {
    return utils.InvalidInput("port must be between %d and %d", s.cfg.Flags.MinPort, s.cfg.Flags.MaxPort)
  }

  if p.MaxPlayers < 1 || uint32(p.MaxPlayers) > s.cfg.Flags.MaxPlayers {
    return utils.InvalidInput("max players must be between 1 and %d", s.cfg.Flags.MaxPlayers)
  }

  minRAM := s.cfg.Flags.MinRAM
  maxRAM := s.cfg.Flags.MaxRAM

  if p.DedicatedRamMb < 0 || uint32(p.DedicatedRamMb) < minRAM || uint32(p.DedicatedRamMb) > maxRAM {
    return utils.InvalidInput("dedicated ram must be between %d and %d", minRAM, maxRAM)
  }

  return validateLogPath(p.LogPath)
}

func validateLogPath(logPath string) error {
  if filepath.Ext(logPath) != ".log" {
    return utils.InvalidInput("log path %q must end in .log", logPath)
  }

  dir := filepath.Dir(logPath)
  info, err := os.Stat(dir)
  if err != nil {
    return utils.InvalidInput("log directory %q: %v", dir, err)
  }

  if !info.IsDir() {
    return utils.InvalidInput("%q is not a directory", dir)
  }

  return nil
}
