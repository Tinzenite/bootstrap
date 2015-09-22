/*
Package bootstrap implements the capability to connect to an existing and online
Tinzenite peer network.

TODO: add encryption bootstrap capabilities
*/
package bootstrap

import (
	"io/ioutil"

	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Create returns a struct that will allow to bootstrap to an existing Tinzenite
network. To actually start bootstrapping call Bootstrap.Start(address).

Path: the absolute path to the directory. localPeerName: the user defined name
of this peer. trusted: whether this should be a trusted peer or an encrypted
one. f: the callback to call once the bootstrap has successfully run.
*/
func Create(path, localPeerName string, trusted bool, f Success) (*Bootstrap, error) {
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
	peer, err := shared.CreatePeer(localPeerName, address, trusted)
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
	toxPeerDump, err := shared.LoadToxDumpFrom(path + "/" + shared.STORETOXDUMPDIR)
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
	stats, err := ioutil.ReadDir(tinzenPath + "/" + shared.ORGDIR + "/" + shared.PEERSDIR)
	if err != nil {
		return false
	}
	exists, _ := shared.FileExists(tinzenPath + "/" + shared.LOCALDIR + "/" + shared.SELFPEERJSON)
	return len(stats) == 1 && exists
}
