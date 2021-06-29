package scores

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var _ ScoresService = &PostgresService{}

type PostgresService struct {
	DB                     *sqlx.DB
	Logger                 *zap.Logger
	ScoresTableName        string
	ParticipationTableName string
	UserIGNTableName       string
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (ps *PostgresService) ClaimScore(
	pid string,
	score int,
	proof string,
) (submissionID string, e error) {
	q := psql.Insert(ps.ScoresTableName).
		Columns("participation_id", "score", "proof").
		Values(pid, score, proof).
		Suffix("RETURNING id").
		RunWith(ps.DB)

	id := ""
	err := q.QueryRow().Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ps *PostgresService) GetOneUnverified() (*ScoreRecord, error) {
	q := psql.Select("e.id as eid", "p.id as pid", "u.id as uid", "u.ign", "e.score", "e.proof", "e.verified").
		From(ps.ScoresTableName + " as e").
		LeftJoin(ps.ParticipationTableName + " as p on p.id = e.participation_id").
		LeftJoin(ps.UserIGNTableName + " as u on u.id = p.user_id").
		Where(sq.Eq{"verified": false}).
		Limit(1)

	record := &ScoreRecord{}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	err = ps.DB.Get(record, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNoRecord{}
		}
		return nil, err
	}

	return record, nil
}

func (ps *PostgresService) GetOneUnverifiedForEvent(eid string) (*ScoreRecord, error) {
	q := psql.Select("e.id as eid", "p.id as pid", "u.id as uid", "u.ign", "e.score", "e.proof", "e.verified").
		From(ps.ScoresTableName + " as e").
		LeftJoin(ps.ParticipationTableName + " as p on p.id = e.participation_id").
		LeftJoin(ps.UserIGNTableName + " as u on u.id = p.user_id").
		Where(sq.Eq{"p.event_id": eid, "verified": false})

	record := &ScoreRecord{}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	err = ps.DB.Get(record, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNoRecord{}
		}
		return nil, err
	}

	return record, nil
}

func (ps *PostgresService) Verify(sid string) error {
	q := psql.Update(ps.ScoresTableName).Set("verified", true).Where(sq.Eq{"id": sid})
	res, err := q.RunWith(ps.DB).Exec()
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return &ErrNoRecord{}
	}
	return nil
}

func (ps *PostgresService) MakeReportScoreSum(eid string) ([]SummaryRecord, error) {
	q := psql.Select("e.score", "u.id as uid", "u.ign").
		FromSelect(psql.Select("sum(e.score) as score", "e.uid").
			FromSelect(psql.Select("e.score", "p.user_id as uid").
				From(ps.ScoresTableName+" as e").
				LeftJoin(ps.ParticipationTableName+" as p on p.id = e.participation_id").
				Where(sq.Eq{"p.event_id": eid, "p.participating": true, "e.verified": true}),
				"e").GroupBy("e.uid"),
			"e").LeftJoin(ps.UserIGNTableName + " as u on u.id = e.uid").OrderBy("e.score desc")
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	leaderboard := []SummaryRecord{}
	err = ps.DB.Select(&leaderboard, query, args...)
	if err != nil {
		return nil, err
	}

	return leaderboard, nil
}

func (ps *PostgresService) MakeReportScoreTop(eid string) ([]SummaryRecord, error) {
	q := psql.Select("s.score", "u.id as uid", "u.ign").
		FromSelect(psql.Select("s.uid", "max(s.score) as score").
			FromSelect(psql.Select("s.participation_id as pid", "s.score as score", "p.user_id as uid").
				From(ps.ScoresTableName+" as s").
				LeftJoin(ps.ParticipationTableName+" as p on p.id = s.participation_id").
				Where(sq.Eq{"p.participating": true, "p.event_id": eid, "s.verified": true}),
				"s").
			GroupBy("s.uid"), "s").
		LeftJoin(ps.UserIGNTableName + " as u on u.id = s.uid")

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	leaderboard := []SummaryRecord{}
	err = ps.DB.Select(&leaderboard, query, args...)
	if err != nil {
		return nil, err
	}

	return leaderboard, nil
}

func (ps *PostgresService) VerificationStatus(eid string) (total, verified int, e error) {
	q := psql.Select(
		"count(case s.verified when TRUE then 1 else null end) as verified",
		"count(case s.verified when FALSE then 1 else null end) as pending",
	).FromSelect(
		psql.Select("s.verified").From(ps.ScoresTableName+" as s").
			LeftJoin(ps.ParticipationTableName+" as p on p.id = s.participation_id").
			Where(sq.Eq{"p.participating": true, "p.event_id": eid}),
		"s",
	)

	var done, pending int
	err := q.RunWith(ps.DB).QueryRow().Scan(&done, &pending)
	if err != nil {
		e = err
		return
	}

	return done + pending, done, nil
}

func (ps *PostgresService) DeleteScore(sid string) error {
	res, err := psql.Delete(ps.ScoresTableName).Where(sq.Eq{"id": sid}).
		RunWith(ps.DB).Exec()
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return &ErrNoRecord{}
	}

	return nil
}

func (ps *PostgresService) UpdateScoreAndVerify(sid string, score int) error {
	res, err := psql.Update(ps.ScoresTableName).
		Set("score", score).
		Set("verified", true).
		Where(sq.Eq{"id": sid}).
		RunWith(ps.DB).
		Exec()

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return &ErrNoRecord{}
	}

	return nil
}
