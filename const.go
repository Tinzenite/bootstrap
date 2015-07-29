package bootstrap

import "errors"

var (
	errNotBootstrapCapable = errors.New("can not bootstrap given peer")
)
