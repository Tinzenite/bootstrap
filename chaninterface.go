package bootstrap

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/tinzenite/model"
	"github.com/tinzenite/shared"
)

type chaninterface struct {
	boot     *Bootstrap                       // reference back to Bootstrap
	model    *model.Model                     // model reference for trusted Bootstrap
	messages map[string]*shared.UpdateMessage // we need to remember all update messages so that we can apply them when received
	pushes   map[string]string                // push messages as identification with the associated name to store the peers as
}

func createChanInterface(boot *Bootstrap) *chaninterface {
	return &chaninterface{
		boot:     boot,
		model:    nil, // NOTE: once created it means that a bootstrap is in progress!
		messages: make(map[string]*shared.UpdateMessage),
		pushes:   make(map[string]string)}
}

func (c *chaninterface) OnFriendRequest(address, message string) {
	log.Println("NewConnection:", address[:8], "ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	// trusted are not expected to send any messages
	if c.boot.IsTrusted() {
		log.Println("Bootstrap received message from", address[:8], ":", message)
		return
	}
	// encrypted can receive push messages however
	v := &shared.Message{}
	err := json.Unmarshal([]byte(message), v)
	// if error most likely not JSON
	if err != nil {
		log.Println("Bootstrap received message from", address[:8], ":", message)
		return
	}
	// make sure only push is allowed
	if v.Type != shared.MsgPush {
		log.Println("Bootstrap received invalid message:", v.Type)
		return
	}
	// read push message
	pm := &shared.PushMessage{}
	err = json.Unmarshal([]byte(message), pm)
	if err != nil {
		log.Println("Bootstrap failed to read push message:", err.Error())
		return
	}
	// as soon as we received a valid push message we can stop notifying the peer
	c.boot.stop <- false
	// store in map by identification (length also signals when we are done)
	c.pushes[pm.Identification] = pm.Name
	// now wait for the files to be received...
}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	// we accept all files!
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
	// different behaviour if we are bootstrapping an encrypted peer
	if !c.boot.IsTrusted() {
		err := c.onPeerFile(path, identification)
		if err != nil {
			log.Println("Failed to apply received peer!", err.Error())
		}
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
		// no need to keep sending check in backgronud
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
		// files must be fetched first, so:
		// we have to remember the update messages because we'll need to apply them
		c.messages[um.Object.Identification] = um
		// create & modify must first fetch file
		rm := shared.CreateRequestMessage(shared.OtObject, um.Object.Identification)
		// request file and apply update on success
		c.boot.channel.Send(address, rm.JSON())
	}
	return nil
}

func (c *chaninterface) onFile(path, identification string) error {
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
		// execute callback
		c.boot.done()
		/* NOTE: it is important that if the bootstrap was successful, DO NOT
		CALL boot.Close() from within this method! */
		// done so return nil
		return nil
	}
	return nil
}

/*
onPeerFile handles the reception of a peer file. If all peers for which push
messages have been received have been fully received this also finishes the
bootstrap.
*/
func (c *chaninterface) onPeerFile(path, identification string) error {
	// retrieve name from pushes
	name, exists := c.pushes[identification]
	if !exists {
		return errors.New("no name for identification found")
	}
	// move file to correct location, named correctly
	err := os.Rename(path, c.boot.path+"/"+shared.ORGDIR+"/"+shared.PEERSDIR+"/"+name)
	if err != nil {
		return err
	}
	// check if we're done
	if len(c.pushes) == 0 {
		// this means we're done!
		c.boot.done()
	}
	return nil
}
