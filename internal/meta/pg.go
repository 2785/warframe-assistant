package meta

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var _ Service = &PostgresService{}

type PostgresService struct {
	DB                 *sqlx.DB
	ActionRoleTable    string
	Logger             *zap.Logger
	IGNTable           string
	EventsTable        string
	ParticipationTable string
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const pgErrUniqueConstraintViolation string = "23505"

// GetRoleRequirementForGuild returns empty string if there's no requirement
func (ps *PostgresService) GetRoleRequirementForGuild(action string, gid string) (string, error) {
	q := psql.Select("role_id").From(ps.ActionRoleTable).Where(sq.Eq{"guild_id": gid, "action": action})
	roleID := ""
	err := q.RunWith(ps.DB).Scan(&roleID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return roleID, nil
}

// IGN relation CRUD
func (ps *PostgresService) CreateIGN(userID, ign string) error {
	q := psql.Insert(ps.IGNTable).Columns("id", "ign").Values(userID, ign)
	_, err := q.RunWith(ps.DB).Exec()
	if err != nil {
		pqErr := &pq.Error{}
		if errors.As(err, &pqErr) {
			if string(pqErr.Code) == pgErrUniqueConstraintViolation {
				return &ErrDuplicateEntry{fmt.Sprintf("user with id '%s' already has an IGN registered", userID)}
			}
		}
		return err
	}
	return nil
}

func (ps *PostgresService) GetIGN(userID string) (string, error) {
	q := psql.Select("ign").From(ps.IGNTable).Where(sq.Eq{"id": userID})
	ign := ""
	err := q.RunWith(ps.DB).Scan(&ign)
	if err == sql.ErrNoRows {
		return "", &ErrNoRecord{}
	}

	return ign, err
}

func (ps *PostgresService) ListAllIGN() (map[string]string, error) {
	q := psql.Select("id", "ign").From(ps.IGNTable)
	type m struct {
		ID  string `db:"id"`
		IGN string `db:"ign"`
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	igns := []m{}
	err = ps.DB.Select(&igns, query, args...)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]string, len(igns))
	for _, v := range igns {
		mapping[v.ID] = v.IGN
	}

	return mapping, nil
}

func (ps *PostgresService) UpdateIGN(userID, newIGN string) error {
	q := psql.Update(ps.IGNTable).Set("ign", newIGN).Where(sq.Eq{"id": userID})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) DeleteRelation(userID string) error {
	q := psql.Delete(ps.IGNTable).Where(sq.Eq{"id": userID})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

// Event CRUD

func (ps *PostgresService) CreateEvent(name, eventType string, start, end time.Time, gid string, active bool) (string, error) {
	q := psql.Insert(ps.EventsTable).
		Columns("guild_id", "name", "start_date", "end_date", "active", "event_type").
		Values(gid, name, start, end, active, eventType).Suffix("RETURNING id")

	id := ""
	err := q.RunWith(ps.DB).QueryRow().Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (ps *PostgresService) UpdateEvent(id, name, eventType string, start, end time.Time, gid string, active bool) error {
	q := psql.Update(ps.EventsTable).SetMap(map[string]interface{}{
		"guild_id":   gid,
		"name":       name,
		"start_date": start,
		"end_date":   end,
		"active":     active,
		"event_type": eventType,
	}).Where(sq.Eq{"id": id})

	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) SetEventStatus(id string, status bool) error {
	q := psql.Update(ps.EventsTable).Set("active", status).Where(sq.Eq{"id": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) SetEventEndDate(id string, end time.Time) error {
	q := psql.Update(ps.EventsTable).Set("end_date", end).Where(sq.Eq{"id": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) GetEvent(id string) (*Event, error) {
	q := psql.Select("id", "guild_id", "name", "start_date", "end_date", "active", "event_type").From(ps.EventsTable).Where(sq.Eq{"id": id})
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	event := &Event{}

	err = ps.DB.Get(event, query, args...)

	if err != nil {
		return nil, err
	}

	return event, nil
}

func (ps *PostgresService) ListAllEvent() ([]*Event, error) {
	q := psql.Select("id", "guild_id", "name", "start_date", "end_date", "active", "event_type").From(ps.EventsTable)
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	events := []*Event{}

	err = ps.DB.Select(&events, query, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (ps *PostgresService) ListEventsForGuild(gid string) ([]*Event, error) {
	q := psql.Select("id", "guild_id", "name", "start_date", "end_date", "active", "event_type").From(ps.EventsTable).Where(sq.Eq{"guild_id": gid})
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	events := []*Event{}

	err = ps.DB.Select(&events, query, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (ps *PostgresService) ListActiveEventsForGuild(gid string) ([]*Event, error) {
	q := psql.Select("id", "guild_id", "name", "start_date", "end_date", "active", "event_type").From(ps.EventsTable).Where(sq.Eq{"guild_id": gid, "active": true})
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	events := []*Event{}

	err = ps.DB.Select(&events, query, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (ps *PostgresService) DeleteEvent(id string) error {
	q := psql.Delete(ps.EventsTable).Where(sq.Eq{"id": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

// Participation Crud

func (ps *PostgresService) AddParticipation(userID, eventID string) (string, error) {
	q := psql.Insert(ps.ParticipationTable).Columns("user_id", "event_id").Values(userID, eventID).Suffix("RETURNING id")
	// this query can result in error 23503 (foreign key) and 23505 (unique key constraint)

	id := ""
	err := q.RunWith(ps.DB).QueryRow().Scan(&id)
	if err != nil {
		pqErr := &pq.Error{}
		if errors.As(err, &pqErr) {
			if string(pqErr.Code) == pgErrUniqueConstraintViolation {
				return "", &ErrDuplicateEntry{fmt.Sprintf("user with id '%s' is already in event '%s'", userID, eventID)}
			}
		}
		return "", err
	}

	return id, nil
}

func (ps *PostgresService) DeleteParticipation(id string) error {
	q := psql.Delete(ps.ParticipationTable).Where(sq.Eq{"id": id})
	res, err := q.RunWith(ps.DB).Exec()

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return &ErrNoRecord{}
	}

	return nil
}

func (ps *PostgresService) DeleteParticipationByUserAndEvent(uid, eid string) error {
	q := psql.Delete(ps.ParticipationTable).Where(sq.Eq{"user_id": uid, "event_id": eid})
	res, err := q.RunWith(ps.DB).Exec()

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return &ErrNoRecord{}
	}

	return nil
}

func (ps *PostgresService) ListUserForEvent(eid string) (map[string]string, error) {
	q := psql.Select("p.user_id", "u.ign").
		FromSelect(sq.Select("user_id").
			From(ps.ParticipationTable).
			Where(sq.Eq{"event_id": eid}), "p").
		LeftJoin(fmt.Sprintf("%s as u on p.user_id=u.id", ps.IGNTable))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	type participant struct {
		UID string `db:"user_id"`
		IGN string `db:"ign"`
	}

	participants := []participant{}

	err = ps.DB.Select(&participants, query, args...)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(participants))
	for _, v := range participants {
		out[v.UID] = v.IGN
	}
	return out, nil
}

func (ps *PostgresService) UserInEvent(uid, eid string) (bool, error) {
	q := psql.Select("1").From(ps.ParticipationTable).Where(sq.Eq{"user_id": uid, "event_id": eid})

	query, args, err := q.ToSql()
	if err != nil {
		return false, err
	}

	out := 0

	err = ps.DB.Get(&out, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
