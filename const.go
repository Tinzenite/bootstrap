package bootstrap

import (
	"errors"
	"time"
)

const (
	tickSpanResend = 11 * time.Second
	tickSpanNone   = 7 * time.Second
)

var (
	errNotBootstrapCapable = errors.New("can not bootstrap given peer")
)
