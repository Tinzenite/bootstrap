package bootstrap

import (
	"encoding/json"
	"log"

	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Bootstrap is a temporary peer object that allows to bootstrap into an existing
Tinzenite network.
*/
type Bootstrap struct {
	// root path
	path string
	// internal hidden struct for channel callbacks
	cInterface *chaninterface
	// tox communication channel
	channel *channel.Channel
	// self peer
	peer *shared.Peer
	// stores address of peers we need to bootstrap
	bootstrap map[string]bool
}

/*
Start begins a bootstrap process to the given address.
*/
func (b *Bootstrap) Start(address string) error {
	// send own peer
	msg, err := json.Marshal(b.peer)
	if err != nil {
		return err
	}
	// send request
	return b.channel.RequestConnection(address, string(msg))
}

/*
Check looks if the bootstrapped address is online and initiates the boostrap
process.

TODO: we can do this in the background... sigh
*/
func (b *Bootstrap) Check() {
	addresses, err := b.channel.OnlineAddresses()
	if err != nil {
		log.Println("Check:", err)
		return
	}
	if len(addresses) != 1 {
		/*TODO pick one? Randomly?*/
		log.Println("Multiple online!")
	}
	/*TODO start bootstrap*/
}

/*
Store writes a bootstrapped .TINZENITEDIR to disk. Call this if you want
persistant bootstrapping (and why wouldn't you?).
*/
func (b *Bootstrap) Store() error {
	err := shared.MakeDotTinzenite(b.path)
	if err != nil {
		return err
	}
	// write self peer
	err = b.peer.Store(b.path)
	if err != nil {
		return err
	}
	// store local peer info with toxdata
	toxData, err := b.channel.ToxData()
	if err != nil {
		return err
	}
	toxPeerDump := &shared.ToxPeerDump{
		SelfPeer: b.peer,
		ToxData:  toxData}
	err = toxPeerDump.Store(b.path)
	if err != nil {
		return err
	}
	return nil
}
