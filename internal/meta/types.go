package meta

import (
	"errors"
	"time"
)

type Service interface {
	GetRoleRequirementForGuild(action string, gid string) (string, error)
	IGNService
	EventService
	ParticipationService
}

type IGNService interface {
	CreateIGN(userID, ign string) error
	GetIGN(userID string) (string, error)
	ListAllIGN() (map[string]string, error)
	UpdateIGN(userID, newIGN string) error
	DeleteRelation(userID string) error
}

type EventService interface {
	CreateEvent(
		name, eventType string,
		start, end time.Time,
		gid string,
		active bool,
	) (string, error)
	UpdateEvent(id, name, eventType string, start, end time.Time, gid string, active bool) error
	SetEventStatus(id string, status bool) error
	SetEventEndDate(id string, end time.Time) error
	GetEvent(id string) (*Event, error)
	ListAllEvent() ([]*Event, error)
	ListEventsForGuild(gid string) ([]*Event, error)
	ListActiveEventsForGuild(gid string) ([]*Event, error)
	DeleteEvent(id string) error
}

type ParticipationService interface {
	AddParticipation(userID, eventID string, participating bool) (string, error)
	DeleteParticipation(id string) error
	DeleteParticipationByUserAndEvent(uid, eid string) error
	ListUserForEvent(eid string) (map[string]string, map[string]string, error)
	UserInEvent(uid, eid string) (bool, error)
	SetParticipation(id string, status bool) error
	SetParticipationByUserAndEvent(uid, eid string, status bool) error
	GetParticipation(uid, eid string) (string, bool, error)
}

type Event struct {
	ID        string    `db:"id"`
	GID       string    `db:"guild_id"`
	Name      string    `db:"name"`
	Begin     time.Time `db:"start_date"`
	End       time.Time `db:"end_date"`
	Active    bool      `db:"active"`
	EventType string    `db:"event_type"`
}

var _ error = &ErrNoRecord{}

type ErrNoRecord struct{}

func (e *ErrNoRecord) Error() string {
	return "no records found"
}

func AsErrNoRecord(e error) bool {
	nr := &ErrNoRecord{}
	return errors.As(e, &nr)
}

var _ error = &ErrDuplicateEntry{}

type ErrDuplicateEntry struct{ M string }

func (e *ErrDuplicateEntry) Error() string {
	return "duplicate entry: " + e.M
}
