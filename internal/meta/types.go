package meta

type Service interface {
	GetRoleRequirementForGuild(action string, gid string) (string, error)
}
