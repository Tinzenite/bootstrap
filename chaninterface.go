package bootstrap

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/tinzenite/model"
	"github.com/tinzenite/shared"
)

type chaninterface struct {
	// reference back to Bootstrap
	boot *Bootstrap
	// model reference NOTE: once created it means that a bootstrap is in progress!
	model *model.Model
}

func createChanInterface(boot *Bootstrap) *chaninterface {
	return &chaninterface{boot: boot}
}

func (c *chaninterface) OnNewConnection(address, message string) {
	log.Println("NewConnection:", address[:8], "ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	log.Println("MSG from", address[:8], ":", message)
}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	filename := address + "." + name
	log.Println("AllowFile: writing file as", filename, ".")
	return true, c.boot.path + "/" + shared.TINZENITEDIR + "/" + shared.RECEIVINGDIR + "/" + filename
}

func (c *chaninterface) OnFileReceived(address, path, name string) {
	log.Println("TODO: received", name, "at", path)
	// split filename to get identification
	check := strings.Split(name, ".")[0]
	identification := strings.Split(name, ".")[1]
	if check != address {
		log.Println("Filename is mismatched!")
		return
	}
	if c.model != nil {
		log.Println("Receiving file!")
	} else {
		log.Println("Receiving model!") // <-- should only be called once!
	}
	// if not model this is an update --> handle it
	if identification != shared.IDMODEL {
		log.Println("Doesn't seem to be a model, do special stuff!")
		// rename to correct name for model
		err := os.Rename(path, c.boot.path+"/"+shared.TINZENITEDIR+"/"+shared.TEMPDIR+"/"+identification)
		if err != nil {
			log.Println("Failed to move file to temp: " + err.Error())
			return
		}
		// apply
		log.Println("TODO need to store the updatemessages... sigh")
		return
		/*
			obj := &shared.ObjectInfo{}
			um := shared.CreateUpdateMessage(shared.OpCreate, *obj)
			c.model.ApplyUpdateMessage(&um)
			// done
			return
		*/
	}
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
	// make a model of the local stuff
	m, err := model.Create(c.boot.path, c.boot.peer.Identification)
	if err != nil {
		log.Println("ReModel:", err)
		return
	}
	c.model = m
	// apply what is already here
	err = c.model.Update()
	if err != nil {
		log.Println("ReModel:", err)
		return
	}
	// get difference in updates
	var updateLists []*shared.UpdateMessage
	updateLists, err = c.model.BootstrapModel(foreignModel)
	if err != nil {
		log.Println("ReModel:", err)
		return
	}
	// pretend that the updatemessage came from outside here
	for _, um := range updateLists {
		// create & modify must first fetch file
		rm := shared.CreateRequestMessage(shared.ReObject, um.Object.Identification)
		// request file and apply update on success
		c.boot.channel.Send(address, rm.String())
	}
}

func (c *chaninterface) OnConnected(address string) {
	log.Println("Connected:", address[:8])
}
