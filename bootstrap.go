package bootstrap

import (
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

func (b *Bootstrap) Start(address string) error {
	/*
		// send own peer
		msg, err := json.Marshal(c.tin.selfpeer)
		if err != nil {
			return err
		}
		// send request
		err = c.tin.channel.RequestConnection(address, string(msg))
		if err != nil {
			return err
		}
		// if request is sent successfully, remember for bootstrap
		// format to legal address
		address = strings.ToLower(address)[:64]
		c.bootstrap[address] = true
		return nil
	*/
	return shared.ErrUnsupported
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
