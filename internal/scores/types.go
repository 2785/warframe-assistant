package scores

import "errors"

type ScoresService interface {
	ClaimScore(pid string, score int, proof string) (submissionID string, e error)
	GetOneUnverified() (*ScoreRecord, error)
	GetOneUnverifiedForEvent(eid string) (*ScoreRecord, error)
	Verify(sid string) error
	MakeReportScoreSum(eid string) ([]SummaryRecord, error)
	MakeReportScoreTop(eid string) ([]SummaryRecord, error)
	VerificationStatus(eid string) (total, verified int, e error)
	DeleteScore(sid string) error
	UpdateScoreAndVerify(sid string, score int) error
}

type ScoreRecord struct {
	ID       string `db:"eid"`
	PID      string `db:"pid"`
	UID      string `db:"uid"`
	IGN      string `db:"ign"`
	Score    int    `db:"score"`
	Proof    string `db:"proof"`
	Verified bool   `db:"verified"`
}

type SummaryRecord struct {
	UID   string `db:"uid"`
	IGN   string `db:"ign"`
	Score int    `db:"score"`
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
