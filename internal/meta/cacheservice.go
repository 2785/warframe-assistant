package meta

import (
	"time"

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

	err := s.c.Once(gid+":"+action, &rid, func() (interface{}, error) {
		return s.Service.GetRoleRequirementForGuild(action, gid)
	})

	return rid, err
}

func (s *CacheService) GetIGN(userID string) (string, error) {
	ign := ""

	err := s.c.Once("ign:"+userID, &ign, func() (interface{}, error) {
		return s.Service.GetIGN(userID)
	})

	return ign, err
}

func (s *CacheService) UpdateIGN(userID, newIGN string) error {
	err := s.c.Drop("ign:" + userID)
	if err != nil {
		s.l.Error("could not delete entry from cache", zap.Error(err), zap.String("uid", userID))
	}
	return s.Service.UpdateIGN(userID, newIGN)
}

func (s *CacheService) DeleteRelation(userID string) error {
	err := s.c.Drop("ign:" + userID)
	if err != nil {
		s.l.Error("could not delete entry from cache", zap.Error(err), zap.String("uid", userID))
	}
	return s.Service.DeleteRelation(userID)
}

func (s *CacheService) UpdateEvent(id, name, eventType string,
	start, end time.Time,
	gid string,
	active bool,
) error {
	err := s.c.Drop("event:" + id)
	if err != nil {
		s.l.Error("could not delete entry from cache", zap.Error(err), zap.String("eid", id))
	}
	return s.Service.UpdateEvent(id, name, eventType, start, end, gid, active)
}

func (s *CacheService) SetEventStatus(id string, status bool) error {
	err := s.c.Drop("event:" + id)
	if err != nil {
		s.l.Error("could not delete entry from cache", zap.Error(err), zap.String("eid", id))
	}
	return s.Service.SetEventStatus(id, status)
}

func (s *CacheService) SetEventEndDate(id string, end time.Time) error {
	err := s.c.Drop("event:" + id)
	if err != nil {
		s.l.Error("could not delete entry from cache", zap.Error(err), zap.String("eid", id))
	}
	return s.Service.SetEventEndDate(id, end)
}

func (s *CacheService) GetEvent(id string) (*Event, error) {
	event := &Event{}
	err := s.c.Once("event:"+id, event, func() (interface{}, error) {
		return s.Service.GetEvent(id)
	})
	return event, err
}
