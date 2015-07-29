package bootstrap

type chaninterface struct {
	// reference back to Bootstrap
	boot *Bootstrap
}

func createChanInterface(boot *Bootstrap) *chaninterface {
	return &chaninterface{boot: boot}
}

func (c *chaninterface) OnNewConnection(address, message string) {}

func (c *chaninterface) OnMessage(address, message string) {}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	return false, ""
}

func (c *chaninterface) OnFileReceived(address, path, name string) {}

func (c *chaninterface) OnConnected(address string) {
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
