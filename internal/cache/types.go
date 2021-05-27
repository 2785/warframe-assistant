package cache

type Cache interface {
	Set(key string, val interface{}) error
	Get(key string) (interface{}, bool)
	Drop(key string) error
}
