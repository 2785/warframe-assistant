package meta

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

var _ Service = &PostgresService{}

type PostgresService struct {
	DB     *sqlx.DB
	Table  string
	Logger *zap.Logger
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// GetRoleRequirementForGuild returns empty string if there's no requirement
func (ps *PostgresService) GetRoleRequirementForGuild(action string, gid string) (string, error) {
	q := psql.Select("role_id").From(ps.Table).Where(sq.Eq{"guild_id": gid, "action": action})
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
