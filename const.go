package bootstrap

import (
	"errors"
	"time"
)

const (
	tickSpanResend = 11 * time.Second
	tickSpanNone   = 7 * time.Second
	tickSpanOnline = 1 * time.Minute
)

var (
	errNotBootstrapCapable = errors.New("can not bootstrap given peer")
)
