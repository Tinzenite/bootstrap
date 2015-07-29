package bootstrap

import (
	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
CreateBootstrap returns a struct that will allow to bootstrap to an existing
Tinzenite network.
*/
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
	boot := &Bootstrap{path: path}
	boot.cInterface = createChanInterface(boot)
	channel, err := channel.Create(localPeerName, nil, boot.cInterface)
	if err != nil {
		return nil, err
	}
	boot.channel = channel
	// get address for peer
	address, err := boot.channel.Address()
	if err != nil {
		return nil, err
	}
	// make peer (at correct location!)
	peer, err := shared.CreatePeer(localPeerName, address)
	if err != nil {
		return nil, err
	}
	boot.peer = peer
	return boot, nil
}

/*
LoadBootstrap tries to load the given directory as a bootstrap object, allowing
it to connect to an existing network. NOTE: will fail if already connected to
other peers!
*/
func LoadBootstrap(path string) (*Bootstrap, error) {
	if !shared.IsTinzenite(path) {
		return nil, shared.ErrNotTinzenite
	}
	if !checkValidBootstrap(path) {
		return nil, errNotBootstrapCapable
	}
	/*TODO check for one peer and channel stuff for success - this would allow
	bootstrapping even non bootstrap directories.*/
	return nil, shared.ErrUnsupported
}

/*
checkValidBootstrap from given path. To be valid: must be .TINZENITEDIR, must
have only one peer (hopefully itself), and local/self.json must exit.
*/
func checkValidBootstrap(path string) bool {
	if !shared.IsTinzenite(path) {
		return false
	}
	tinzenPath := path + "/" + shared.TINZENITEDIR
	peerAmount, err := shared.CountFiles(tinzenPath + "/" + shared.ORGDIR + "/" + shared.PEERSDIR)
	if err != nil {
		return false
	}
	return peerAmount == 1 && shared.FileExists(tinzenPath+"/"+shared.LOCALDIR+"/"+shared.SELFPEERJSON)
}
