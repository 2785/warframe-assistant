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

type ScoreSummary struct {
	UID        string `db:"user_id"`
	IGN        string `db:"ign"`
	TotalScore int    `db:"score"`
}

type ScoreSubmission struct {
	ID       string `db:"id"`
	UID      string `db:"user_id"`
	EID      string `db:"event_id"`
	Proof    string `db:"proof"`
	Verified bool   `db:"verified"`
	Score    int    `db:"score"`
}

var _ error = &ErrNoRecord{}

type ErrNoRecord struct{}

func (e *ErrNoRecord) Error() string {
	return "no records found"
}
