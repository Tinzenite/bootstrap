package bootstrap

import (
	"errors"
	"time"
)

const tickSpan = 10 * time.Second

var (
	errNotBootstrapCapable = errors.New("can not bootstrap given peer")
)
