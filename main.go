package bootstrap

import (
	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

func CreateBootstrap(path, localPeerName string) (*Bootstrap, error) {
	if shared.IsTinzenite(path) {
		return nil, shared.ErrIsTinzenite
	}
	// build structure
	err := shared.MakeDotTinzenite(path)
	if err != nil {
		return nil, err
	}
	// create object
	boot := &Bootstrap{}
	boot.cInterface = createChanInterface(boot)
	channel, err := channel.Create(localPeerName, nil, boot.cInterface)
	if err != nil {
		return nil, err
	}
	boot.channel = channel
	// make peer (at correct location!)
	// peer := shared.
	return nil, shared.ErrUnsupported
}

func LoadBootstrap(path string) (*Bootstrap, error) {
	if !shared.IsTinzenite(path) {
		return nil, shared.ErrNotTinzenite
	}
	return nil, shared.ErrUnsupported
}
