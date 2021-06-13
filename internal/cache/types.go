package cache

import "errors"

type Cache interface {
	Set(key string, val interface{}) error
	Get(key string, val interface{}) error
	Drop(key string) error
}

var _ error = &ErrNoRecord{}

type ErrNoRecord struct{}

func (e *ErrNoRecord) Error() string {
	return "there's no record"
}

func AsErrNoRecord(e error) bool {
	nr := &ErrNoRecord{}
	return errors.As(e, &nr)
}
