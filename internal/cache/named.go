package cache

var _ Cache = &NamedCache{}

type NamedCache struct {
	c      Cache
	prefix string
}

func Named(prefix string, c Cache) *NamedCache {
	return &NamedCache{c, prefix}
}

func (c *NamedCache) Set(key string, val interface{}) error {
	return c.c.Set(c.prefix+":"+key, val)
}

func (c *NamedCache) Get(key string, val interface{}) error {
	return c.c.Get(c.prefix+":"+key, val)
}

func (c *NamedCache) Once(key string, recv interface{}, do func() (interface{}, error)) error {
	return c.c.Once(c.prefix+":"+key, recv, do)
}

func (c *NamedCache) Drop(key string) error {
	return c.c.Drop(c.prefix + ":" + key)
}
