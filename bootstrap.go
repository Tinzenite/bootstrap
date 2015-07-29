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
	// internal hidden struct for channel callbacks
	cInterface *chaninterface
	// tox communication channel
	channel *channel.Channel
	// stores address of peers we need to bootstrap
	bootstrap map[string]bool
	// self peer
	peer shared.Peer
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

func (b *Bootstrap) OnConnected(address string) {
	/*
		_, exists := c.bootstrap[address]
		if !exists {
			log.Println("Missing", address)
			// nope, doesn't need bootstrap
			return
		}
		// bootstrap
		rm := shared.CreateRequestMessage(shared.ReModel, IDMODEL)
		c.requestFile(address, rm, func(address, path string) {
			// read model file and remove it
			data, err := ioutil.ReadFile(path)
			if err != nil {
				log.Println("ReModel:", err)
				return
			}
			err = os.Remove(path)
			if err != nil {
				log.Println("ReModel:", err)
				// not strictly critical so no return here
			}
			// unmarshal
			foreignModel := &shared.ObjectInfo{}
			err = json.Unmarshal(data, foreignModel)
			if err != nil {
				log.Println("ReModel:", err)
				return
			}
			// get difference in updates
			var updateLists []*shared.UpdateMessage
			updateLists, err = c.tin.model.BootstrapModel(foreignModel)
			if err != nil {
				log.Println("ReModel:", err)
				return
			}
			// pretend that the updatemessage came from outside here
			for _, um := range updateLists {
				c.remoteUpdate(address, *um)
			}
			// bootstrap --> special behaviour, so call the finish method
			log.Println("BOOTSTRAPPED: HOW DO I APPLY IT?")
		})
	*/
}
