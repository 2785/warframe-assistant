package meta

import "go.uber.org/zap"

type Cache interface {
	Set(key string, val interface{}) error
	Get(key string) (interface{}, bool)
}

type CacheService struct {
	s Service
	c Cache
	l *zap.Logger
}

func NewWithCache(s Service, c Cache, l *zap.Logger) *CacheService {
	return &CacheService{s, c, l}
}

func (s *CacheService) GetRoleRequirementForGuild(action string, gid string) (string, error) {
	if val, ok := s.c.Get(gid + ":" + action); ok {
		s, ok := val.(string)
		if ok {
			return s, nil
		}
	}

	role_id, err := s.s.GetRoleRequirementForGuild(action, gid)
	if err != nil {
		return "", err
	}

	err = s.c.Set(gid+":"+action, role_id)
	if err != nil {
		s.l.Error("could not add entry into cache", zap.Error(err))
	}

	return role_id, nil
}
