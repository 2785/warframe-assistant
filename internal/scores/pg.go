package scores

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var _ ScoresService = &PostgresService{}

type PostgresService struct {
	DB               *sqlx.DB
	Table            string
	Logger           *zap.Logger
	ScoresTableName  string
	UserIGNTableName string
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (ps *PostgresService) AddUnverified(userID, ign string, score int, proof string) (submissionID string, e error) {
	q := psql.Insert(ps.Table).
		Columns("userid", "ign", "score", "proof").
		Values(userID, ign, score, proof).
		Suffix("RETURNING submissionid").
		RunWith(ps.DB)

	id := ""
	err := q.QueryRow().Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ps *PostgresService) SubmitScore(userID, eventID, proof string, score int) (string, error) {
	q := psql.Insert(ps.ScoresTableName).
		Columns("user_id", "event_id", "score", "proof").
		Values(userID, eventID, score, proof).
		Suffix("RETURNING id").
		RunWith(ps.DB)

	id := ""
	err := q.QueryRow().Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (ps *PostgresService) GetOneUnverified() (submissionID, userID, ign string, score int, proof string, e error) {
	r := &Record{}
	q := psql.Select("*").From(ps.Table).Where(sq.Eq{"verified": false})

	query, args, err := q.ToSql()
	if err != nil {
		e = err
		return
	}

	err = ps.DB.Get(r, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			e = &ErrNoRecord{}
			return
		}
		e = err
		return
	}

	return r.ID, r.UserID, r.IGN, r.Score, r.Proof, nil
}

func (ps *PostgresService) GetOneUnverifiedSubmission(eventID string) (*ScoreSubmission, error) {
	r := &ScoreSubmission{}
	q := psql.Select("*").From(ps.ScoresTableName).Where(sq.Eq{"verified": false}, sq.Eq{"event_id": eventID})

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	err = ps.DB.Get(r, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			e := &ErrNoRecord{}
			return nil, e
		}
		return nil, err
	}

	return r, nil
}

func (ps *PostgresService) Verify(id string) error {
	q := psql.Update(ps.Table).Set("verified", true).Where(sq.Eq{"submissionid": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) VerifySubmission(id string) error {
	q := psql.Update(ps.ScoresTableName).Set("verified", true).Where(sq.Eq{"id": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) ScoreReport() ([]SummaryRecord, error) {

	r := []SummaryRecord{}
	q := psql.Select("ign", "userid", "SUM(score)").From(ps.Table).Where(sq.Eq{"verified": true}).GroupBy("ign", "userid").OrderBy("sum DESC")

	sql, args, err := q.ToSql()
	if err != nil {
		// this is a constant query, should never err
		panic(err)
	}

	err = ps.DB.Select(&r, sql, args...)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (ps *PostgresService) ScoreReportForEvent(eventID string) ([]ScoreSummary, error) {
	r := []ScoreSummary{}
	q := psql.Select("m.user_id", "m.score", "u.ign").
		FromSelect(psql.Select("user_id", "sum(score) as score").
			From(ps.ScoresTableName).
			Where(sq.Eq{"verified": true}, sq.Eq{"event_id": eventID}).
			GroupBy("user_id"), "m").
		LeftJoin(fmt.Sprintf("%s as u on m.user_id=u.id", ps.UserIGNTableName)).OrderBy("score DESC")

	sql, args, err := q.ToSql()
	if err != nil {
		// this is a constant query, should never err
		panic(err)
	}

	err = ps.DB.Select(&r, sql, args...)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (ps *PostgresService) VerificationStatus() (total, verified, pending int, e error) {
	q := psql.Select(
		"count(*) filter (where verified = TRUE) as verified",
		"count(*) filter (where verified = FALSE) as pending",
	).From(ps.Table)

	type d struct {
		Verified int `db:"verified"`
		Pending  int `db:"pending"`
	}

	dat := d{}

	query, args, err := q.ToSql()
	if err != nil {
		e = err
		return
	}

	err = ps.DB.Get(&dat, query, args...)

	if err != nil {
		e = err
		return
	}

	return dat.Verified + dat.Pending, dat.Verified, dat.Pending, nil
}

func (ps *PostgresService) VerificationStatusForEvent(eventID string) (total, verified, pending int, e error) {
	q := psql.Select(
		"count(*) filter (where verified = TRUE) as verified",
		"count(*) filter (where verified = FALSE) as pending",
	).From(ps.ScoresTableName).
		Where(sq.Eq{"event_id": eventID})

	type d struct {
		Verified int `db:"verified"`
		Pending  int `db:"pending"`
	}

	dat := d{}

	query, args, err := q.ToSql()
	if err != nil {
		e = err
		return
	}

	err = ps.DB.Get(&dat, query, args...)

	if err != nil {
		e = err
		return
	}

	return dat.Verified + dat.Pending, dat.Verified, dat.Pending, nil
}

func (ps *PostgresService) DeleteRecord(id string) error {
	q := psql.Delete(ps.Table).Where(sq.Eq{"submissionid": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) DeleteSubmission(id string) error {
	q := psql.Delete(ps.ScoresTableName).Where(sq.Eq{"id": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) UpdateScore(id string, score int) error {
	q := psql.Update(ps.Table).Set("score", score).Set("verified", true).Where(sq.Eq{"submissionid": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}

func (ps *PostgresService) UpdateSubmissionScore(id string, score int) error {
	q := psql.Update(ps.ScoresTableName).Set("score", score).Set("verified", true).Where(sq.Eq{"id": id})
	_, err := q.RunWith(ps.DB).Exec()
	return err
}
