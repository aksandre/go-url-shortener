package config

type ConfigTypeInterface interface {
	installFlags()
	GetAddrServer() *addressServer
	GetHostShortLink() *hostShortLink
}

type сonfigType struct {
	addrServer    *addressServer
	hostShortLink *hostShortLink
}

func (ct *сonfigType) installFlags() {
	ct.addrServer = AddressServerFlag
	ct.hostShortLink = HostShortLinkFlag
}

func (ct *сonfigType) GetAddrServer() *addressServer {
	return ct.addrServer
}

func (ct *сonfigType) GetHostShortLink() *hostShortLink {
	return ct.hostShortLink
}

func NewConfigApp() ConfigTypeInterface {
	// переменная конфига
	configApp := &сonfigType{}
	configApp.installFlags()
	return configApp
}
