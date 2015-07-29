/*
Package bootstrap implements the capability to connect to an existing and online
Tinzenite peer network.

TODO: add encryption bootstrap capabilities
*/
package bootstrap

import (
	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Create returns a struct that will allow to bootstrap to an existing
Tinzenite network. To actually start bootstrapping call Bootstrap.Start(address).
*/
func Create(path, localPeerName string, f Success) (*Bootstrap, error) {
	if shared.IsTinzenite(path) {
		return nil, shared.ErrIsTinzenite
	}
	// build structure
	err := shared.MakeDotTinzenite(path)
	if err != nil {
		return nil, err
	}
	// create object
	boot := &Bootstrap{
		path:   path,
		onDone: f}
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
	// bg stuff
	boot.wg.Add(1)
	boot.stop = make(chan bool, 1)
	go boot.run()
	return boot, nil
}

/*
Load tries to load the given directory as a bootstrap object, allowing it to
connect to an existing network. To actually start bootstrapping call
Bootstrap.Start(address). NOTE: will fail if already connected to other peers!

TODO: strictly speaking we only need the selfpeer... look into this?
*/
func Load(path string, f Success) (*Bootstrap, error) {
	if !shared.IsTinzenite(path) {
		return nil, shared.ErrNotTinzenite
	}
	if !checkValidBootstrap(path) {
		return nil, errNotBootstrapCapable
	}
	// create object
	boot := &Bootstrap{
		path:   path,
		onDone: f}
	boot.cInterface = createChanInterface(boot)
	// load
	toxPeerDump, err := shared.LoadToxDump(path)
	if err != nil {
		return nil, err
	}
	boot.peer = toxPeerDump.SelfPeer
	channel, err := channel.Create(boot.peer.Name, toxPeerDump.ToxData, boot.cInterface)
	if err != nil {
		return nil, err
	}
	boot.channel = channel
	// bg stuff
	boot.wg.Add(1)
	boot.stop = make(chan bool, 1)
	go boot.run()
	return boot, nil
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
