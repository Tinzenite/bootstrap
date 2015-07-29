package bootstrap

import "github.com/tinzenite/shared"

func CreateBootstrap(path string) (*Bootstrap, error) {
	if shared.IsTinzenite(path) {
		return nil, shared.ErrIsTinzenite
	}
	return nil, shared.ErrUnsupported
}

func LoadBootstrap(path string) (*Bootstrap, error) {
	if !shared.IsTinzenite(path) {
		return nil, shared.ErrNotTinzenite
	}
	return nil, shared.ErrUnsupported
}
