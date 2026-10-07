package host

import (
  "context"
  "net"
  "strings"

  "github.com/jackc/pgx/v5/pgtype"
  "github.com/ozzenerol/mc-1.7.3-orchestrator/internal/db"
  "github.com/ozzenerol/mc-1.7.3-orchestrator/internal/utils"
)

const entity = "host"

type Service struct {
  q *db.Queries
}

func NewService(q *db.Queries) *Service {
  return &Service{
    q: q,
  }
}

func (s *Service) Create(ctx context.Context, name string, ip string) (db.Host, error) {
  if err := validate(name, ip); err != nil {
    return db.Host{}, err
  }

  params := db.CreateHostParams{
    Name: name,
    Ip:   ip,
  }

  h, err := s.q.CreateHost(ctx, params)
  return h, utils.DBError(err, entity)
}

func (s *Service) Get(ctx context.Context, id pgtype.UUID) (db.Host, error) {
  h, err := s.q.GetHost(ctx, id)
  return h, utils.DBError(err, entity)
}

func (s *Service) List(ctx context.Context) ([]db.Host, error) {
  hosts, err := s.q.ListHosts(ctx)
  return hosts, utils.DBError(err, entity)
}

func (s *Service) Update(ctx context.Context, id pgtype.UUID, name string, ip string) (db.Host, error) {
  if err := validate(name, ip); err != nil {
    return db.Host{}, err
  }

  params := db.UpdateHostParams{
    ID:   id,
    Name: name,
    Ip:   ip,
  }

  h, err := s.q.UpdateHost(ctx, params)
  return h, utils.DBError(err, entity)
}

func (s *Service) Delete(ctx context.Context, id pgtype.UUID) error {
  n, err := s.q.DeleteHost(ctx, id)
  if err != nil {
    return utils.DBError(err, entity)
  }

  if n == 0 {
    return utils.NotFound(entity)
  }

  return nil
}

func validate(name string, ip string) error {
  if strings.TrimSpace(name) == "" {
    return utils.InvalidInput("name is required")
  }

  if ip != "localhost" && net.ParseIP(ip) == nil {
    return utils.InvalidInput("invalid ip %q", ip)
  }

  return nil
}
