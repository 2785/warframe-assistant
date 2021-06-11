package discord

import "go.uber.org/zap"

func WithComponent(co string) zap.Field {
	return zap.String("co", co)
}

func WithHandler(h string) zap.Field {
	return zap.String("handler", h)
}

func WithCommand(c string) zap.Field {
	return zap.String("command", c)
}

func WithGuildID(gid string) zap.Field {
	return zap.String("gid", gid)
}

func WithChannelID(cid string) zap.Field {
	return zap.String("cid", cid)
}

func WithUserID(uid string) zap.Field {
	return zap.String("uid", uid)
}

func WithMessageID(mid string) zap.Field {
	return zap.String("mid", mid)
}

func WithSubmissionID(sid string) zap.Field {
	return zap.String("sid", sid)
}

func WithEventID(eid string) zap.Field {
	return zap.String("eid", eid)
}

func WithRoleID(rid string) zap.Field {
	return zap.String("rid", rid)
}
