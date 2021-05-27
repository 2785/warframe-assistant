package scores

type Repository interface {
}

type ScoresService interface {
	AddUnverified(userID, ign string, score int, proof string) (submissionID string, e error)
	GetOneUnverified() (submissionID, userID, ign string, score int, proof string, e error)
	Verify(submissionID string) error
	ScoreReport() ([]SummaryRecord, error)
	VerificationStatus() (total, verified, pending int, e error)
	DeleteRecord(submissionID string) error
	UpdateScore(submissionID string, newScore int) error
}

type Record struct {
	ID       string `db:"submissionid"`
	UserID   string `db:"userid"`
	IGN      string `db:"ign"`
	Score    int    `db:"score"`
	Proof    string `db:"proof"`
	Verified bool   `db:"verified"`
}

type SummaryRecord struct {
	UserID string `db:"userid"`
	IGN    string `db:"ign"`
	Sum    int    `db:"sum"`
}

var _ error = &ErrNoRecord{}

type ErrNoRecord struct{}

func (e *ErrNoRecord) Error() string {
	return "no records found"
}
