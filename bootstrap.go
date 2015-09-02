package bootstrap

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Success is the callback that will be called once the bootstrap is complete.
*/
type Success func()

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
	// callback for when done
	onDone Success
	// stuff for background thread
	wg   sync.WaitGroup
	stop chan bool
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
Address returns the full address of this peer.
*/
func (b *Bootstrap) Address() (string, error) {
	return b.channel.ConnectionAddress()
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

/*
PrintStatus returns a formatted string of the peer status.
*/
func (b *Bootstrap) PrintStatus() string {
	var out string
	out += "Online:\n"
	addresses, err := b.channel.FriendAddresses()
	if err != nil {
		out += "channel.FriendAddresses failed!"
	} else {
		var count int
		for _, address := range addresses {
			online, err := b.channel.IsOnline(address)
			var insert string
			if err != nil {
				insert = "ERROR"
			} else {
				insert = fmt.Sprintf("%v", online)
			}
			out += address[:16] + " :: " + insert + "\n"
			count++
		}
		out += "Total friends: " + fmt.Sprintf("%d", count)
	}
	return out
}

/*
Close cleanly closes everything underlying.
*/
func (b *Bootstrap) Close() {
	// send stop signal
	b.stop <- true
	// wait for it to close
	b.wg.Wait()
	// finally close channel
	b.channel.Close()
}

/*
Run is the background thread that keeps checking if it can bootstrap.
*/
func (b *Bootstrap) run() {
	defer func() { log.Println("Bootstrap:", "Background process stopped.") }()
	online := false
	var interval time.Duration
	for {
		// this ensures 2 different tick spans depending on whether someone is online or not
		if online {
			interval = tickSpanOnline
		} else {
			interval = tickSpanNone
		}
		select {
		case <-b.stop:
			b.wg.Done()
			return
		case <-time.Tick(interval):
			online = false
			addresses, err := b.channel.OnlineAddresses()
			if err != nil {
				log.Println("Check:", err)
				break
			}
			if len(addresses) == 0 {
				log.Println("None available yet.")
				break
			}
			if len(addresses) > 1 {
				// since we'll always only try connecting to one, warn specifically!
				log.Println("WARNING: Multiple online! Will try connecting to ", addresses[0][:16], " only.")
			}
			online = true
			// yo, we want to bootstrap!
			rm := shared.CreateRequestMessage(shared.ReModel, shared.IDMODEL)
			b.channel.Send(addresses[0], rm.JSON())
		} // select
	} // for
}
