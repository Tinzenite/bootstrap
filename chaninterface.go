package bootstrap

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tinzenite/model"
	"github.com/tinzenite/shared"
)

type chaninterface struct {
	boot        *Bootstrap                       // reference back to Bootstrap
	sendAddress string                           // address of connected peer as soon as messages arrive
	model       *model.Model                     // model reference NOTE: once created it means that a bootstrap is in progress! Also it is CORRECT and DESIRED that the model is not stored between runs.
	messages    map[string]*shared.UpdateMessage // we need to remember all update messages so that we can apply them when received
	wg          sync.WaitGroup                   // stuff for background thread
	stop        chan bool                        // stuff for background thread
	mutex       sync.Mutex                       // required for map of update messages
}

func createChanInterface(boot *Bootstrap) *chaninterface {
	cinterface := &chaninterface{
		boot:     boot,
		model:    nil,
		messages: make(map[string]*shared.UpdateMessage)}
	// start bg thread for trying to resend
	cinterface.stop = make(chan bool, 1)
	cinterface.wg.Add(1)
	go cinterface.run()
	return cinterface
}

func (c *chaninterface) OnFriendRequest(address, message string) {
	log.Println("NewConnection:", address[:8], "ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	log.Println("Bootstrap received message from", address[:8], ":", message)
}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	filename := address + "." + name
	return true, c.boot.path + "/" + shared.TINZENITEDIR + "/" + shared.RECEIVINGDIR + "/" + filename
}

func (c *chaninterface) OnFileReceived(address, path, name string) {
	// split filename to get identification
	check := strings.Split(name, ".")[0]
	identification := strings.Split(name, ".")[1]
	if check != address {
		log.Println("Filename is mismatched!")
		return
	}
	// whether we allow accepting a file or model depends on whether we already have a model here...
	if c.model != nil {
		// safe guard
		if len(c.messages) == 0 {
			log.Println("No update messages available! Ignoring file.")
			return
		}
		err := c.onFile(path, identification)
		if err != nil {
			log.Println("onFile:", err)
		}
	} else {
		// safe guard
		if identification != shared.IDMODEL {
			log.Println("Expecting model! Ignoring file.")
			return
		}
		log.Println("Receiving model!") // <-- should only be called once!
		// no need to keep sending check in background
		c.boot.stop <- false
		// go due it
		err := c.onModel(address, path)
		if err != nil {
			log.Println("onModel:", err)
		}
	}
}

func (c *chaninterface) OnFileCanceled(address, path string) {
	log.Println("File transfer was canceled by " + address[:8] + "!")
}

func (c *chaninterface) OnConnected(address string) {
	// remember which peer we are connected to so that we know to whom to send our requests
	c.sendAddress = address
	log.Println("Connected:", address[:8])
}

// ---------------------- NO CALLBACKS BEYOND THIS POINT -----------------------

func (c *chaninterface) onModel(address, path string) error {
	// read model file and remove it
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil {
		log.Println("Failed to remove temp model file:", err)
		// not strictly critical so no return here
	}
	// unmarshal
	foreignModel := &shared.ObjectInfo{}
	err = json.Unmarshal(data, foreignModel)
	if err != nil {
		return err
	}
	// make a model of the local stuff
	storePath := c.boot.path + "/" + shared.TINZENITEDIR + "/" + shared.LOCALDIR
	m, err := model.Create(c.boot.path, c.boot.peer.Identification, storePath)
	if err != nil {
		return err
	}
	c.model = m
	// apply what is already here
	err = c.model.Update()
	if err != nil {
		return err
	}
	// get difference in updates
	var updateLists []*shared.UpdateMessage
	updateLists, err = c.model.Bootstrap(foreignModel)
	if err != nil {
		return err
	}
	log.Println("Need to apply", len(updateLists), "updates.")
	// pretend that the updatemessage came from outside here
	for _, um := range updateLists {
		// directories can be applied directly
		if um.Object.Directory {
			dirPath := m.RootPath + "/" + um.Object.Path
			// if the dir doesn't exist, make it
			if exists, _ := shared.DirectoryExists(dirPath); !exists {
				err := shared.MakeDirectory(dirPath)
				if err != nil {
					log.Println("Failed applying dir:", err)
				}
			}
			// apply to model
			err = c.model.ApplyUpdateMessage(um)
			// ignore merge conflicts as they are to be overwritten anyway
			if err != nil && err != shared.ErrConflict {
				return err
			}
			continue
		}
		// write update messages which must be fetched
		c.messages[um.Object.Identification] = um
	}
	return nil
}

func (c *chaninterface) onFile(path, identification string) error {
	// on receiving a file we work on the map, so lock for duration of function
	c.mutex.Lock()
	defer func() { c.mutex.Unlock() }()
	// see if we have a corresponding update message
	um, exists := c.messages[identification]
	if !exists {
		log.Println("Can not apply file", identification, "!")
		return os.Remove(path)
	}
	// remove from list since we're applying it
	delete(c.messages, identification)
	// rename to correct name for model
	err := os.Rename(path, c.boot.path+"/"+shared.TINZENITEDIR+"/"+shared.TEMPDIR+"/"+identification)
	if err != nil {
		return err
	}
	// apply
	err = c.model.ApplyUpdateMessage(um)
	// ignore merge conflicts as they are to be overwritten anyway
	if err != nil && err != shared.ErrConflict {
		return err
	}
	// detect when done to call success callback
	if len(c.messages) == 0 {
		// write directory to DIRECTORYLIST because it is now a valid TINZENITEDIR
		err := shared.WriteDirectoryList(c.boot.path)
		if err != nil {
			log.Println("Failed to write path to", shared.DIRECTORYLIST, "!")
			// not a critical error but log in case clients can't find the dir
		}
		// finish our own bg thread
		c.stop <- true
		c.wg.Wait()
		log.Println("Resend completed successfully.")
		// execute callback
		c.boot.done()
		/* NOTE: it is important that if the bootstrap was successful, DO NOT
		CALL boot.Close() from within this method! */
		// done so return nil
		return nil
	}
	return nil
}

func (c *chaninterface) run() {
	// resend every so often
	resendTicker := time.Tick(tickSpanResend)
	for {
		select {
		case <-c.stop:
			c.wg.Done()
			return
		case <-resendTicker:
			c.sendOutstandingRequest(c.sendAddress)
		} // select
	} // for
}

/*
sendOutstandingRequest sends all request messages for update messages that have
not yet been applied. This ensures that no files are missed and the bootstrap
successfully completes.
*/
func (c *chaninterface) sendOutstandingRequest(address string) {
	// if no address given nothing to do so return immediately
	if address == "" {
		return
	}
	// lock and unlock map access for this function
	c.mutex.Lock()
	defer func() { c.mutex.Unlock() }()
	// determine messages available
	amount := len(c.messages)
	// if all have been applied this method can stop
	if amount == 0 {
		log.Println("WARNING: nothing to request!")
		return
	}
	// send requests
	log.Println("Requesting", amount, "files.")
	for _, um := range c.messages {
		// make sure we don't request directories
		if um.Object.Directory {
			log.Println("WARNING: trying to request directory, ignoring!")
			continue
		}
		// prepare request file
		rm := shared.CreateRequestMessage(shared.OtObject, um.Object.Identification)
		// send request file
		c.boot.channel.Send(address, rm.JSON())
	}
}
