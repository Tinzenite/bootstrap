package bootstrap

import (
	"errors"
	"time"
)

const (
	tickSpanNone   = 10 * time.Second
	tickSpanOnline = 1 * time.Minute
)

var (
	errNotBootstrapCapable = errors.New("can not bootstrap given peer")
)
