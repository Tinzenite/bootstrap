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

func (c *chaninterface) OnConnected(address string) {}
