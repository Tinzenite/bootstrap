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
Tinzenite network. Also it is CORRECT and DESIRED that a model for a trusted peer
+is not stored between runs to allow resetting if something goes wrong.
*/
type Bootstrap struct {
	path       string           // root path
	cInterface *chaninterface   // internal hidden struct for channel callbacks
	channel    *channel.Channel // tox communication channel
	peer       *shared.Peer     // self peer
	bootstrap  map[string]bool  // stores address of peers we need to bootstrap
	onDone     Success          // callback for when done
	wg         sync.WaitGroup   // stuff for background thread
	stop       chan bool        // stuff for background thread
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
	trusted := b.IsTrusted()
	var err error
	if trusted {
		err = shared.MakeTinzeniteDir(b.path)
	} else {
		err = shared.MakeEncryptedDir(b.path)
	}
	if err != nil {
		return err
	}
	// write self peer if TRUSTED peer. Encrypted don't write their own peer.
	if trusted {
		err = b.peer.StoreTo(b.path + "/" + shared.STOREPEERDIR)
	}
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
	// write toxpeerdump
	if trusted {
		err = toxPeerDump.StoreTo(b.path + "/" + shared.STORETOXDUMPDIR)
	} else {
		err = toxPeerDump.StoreTo(b.path + "/" + shared.LOCALDIR)
	}
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
			online, err := b.channel.IsAddressOnline(address)
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
IsTrusted can be used to read whether this bootstrap is creating an encrypted or
a trusted peer.
*/
func (b *Bootstrap) IsTrusted() bool {
	return b.peer.Trusted
}

/*
Close cleanly closes everything underlying.
*/
func (b *Bootstrap) Close() {
	// ensure that bg thread was closed
	select {
	case b.stop <- true:
		b.wg.Wait()
	default:
		// if closing doesn't work we've probably already closed it
	}
	// close channel
	b.channel.Close()
}

/*
Run is the background thread that keeps checking if it can bootstrap.
*/
func (b *Bootstrap) run() {
	defer func() { log.Println("Bootstrap:", "Background process stopped.") }()
	waitTicker := time.Tick(tickSpanNone)
	for {
		select {
		case <-b.stop:
			b.wg.Done()
			return
		case <-waitTicker:
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
				log.Println("WARNING: Multiple online! Will try connecting to ", addresses[0][:8], " only.")
			}
			// if not trusted, we are done once the connection has been accepted.
			if !b.IsTrusted() {
				// execute callback
				b.done()
				// stop bg thread
				b.stop <- false
				// go quit
				continue
			}
			log.Println("Initiating transfer of directory state.")
			// yo, we want to bootstrap!
			rm := shared.CreateRequestMessage(shared.OtModel, shared.IDMODEL)
			b.channel.Send(addresses[0], rm.JSON())
		} // select
	} // for
}

/*
done is called to execute the callback synchroniously and handle other things.
*/
func (b *Bootstrap) done() {
	// make sure background thread is done but if channel is blocked go on
	select {
	case b.stop <- true:
	default:
	}
	// store up to date tox information
	err := b.Store()
	if err != nil {
		log.Println("Failed to store:", err)
		// warn but continue to execute everything
	}
	// notify of done
	if b.onDone != nil {
		b.onDone()
	} else {
		log.Println("onDone is nil!")
	}
}
