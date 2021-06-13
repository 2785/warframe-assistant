package meta

import (
	"github.com/2785/warframe-assistant/internal/cache"
	"go.uber.org/zap"
)

type CacheService struct {
	c cache.Cache
	l *zap.Logger

	Service
}

func NewWithCache(s Service, c cache.Cache, l *zap.Logger) *CacheService {
	return &CacheService{c, l, s}
}

func (s *CacheService) GetRoleRequirementForGuild(action string, gid string) (string, error) {
	rid := ""

	err := s.c.Get(gid+":"+action, &rid)
	if err == nil {
		return rid, nil
	}

	role_id, err := s.Service.GetRoleRequirementForGuild(action, gid)
	if err != nil {
		return "", err
	}

	err = s.c.Set(gid+":"+action, role_id)
	if err != nil {
		s.l.Error("could not add entry into cache", zap.Error(err))
	}

	return role_id, nil
}
