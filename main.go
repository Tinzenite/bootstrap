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
	var err error
	if trusted {
		err = shared.MakeTinzeniteDir(path)
	} else {
		err = shared.MakeEncryptedDir(path)
	}
	// creation of structure error
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
	// we need to do some logic to detect where we can load stuff from
	trusted, err := isLoadable(path)
	if err != nil {
		return nil, err
	}
	// create object
	boot := &Bootstrap{
		path:   path,
		onDone: f}
	boot.cInterface = createChanInterface(boot)
	// load self peer from correct location
	var toxPeerDump *shared.ToxPeerDump
	if trusted {
		toxPeerDump, err = shared.LoadToxDumpFrom(path + "/" + shared.STORETOXDUMPDIR)
	} else {
		toxPeerDump, err = shared.LoadToxDumpFrom(path + "/" + shared.LOCALDIR)
	}
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
isLoadable returns an error if not loadable. The flag returns whether it is the
dir for a TRUSTED peer (so false if encrypted).
*/
func isLoadable(path string) (bool, error) {
	// we check based on available paths
	tinPath := path + "/" + shared.TINZENITEDIR + "/" + shared.LOCALDIR
	encPath := path + "/" + shared.LOCALDIR
	// first check if .tinzenite (so that visible dirs of enc won't cause false detection)
	if exists, _ := shared.DirectoryExists(tinPath); exists {
		return true, nil
	}
	// second check if encrypted
	if exists, _ := shared.DirectoryExists(encPath); exists {
		return false, nil
	}
	// if neither we're done, so say encrypted but error
	return false, shared.ErrNotTinzenite
}
