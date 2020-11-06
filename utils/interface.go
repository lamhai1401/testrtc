package utils

// Fwdm linter
type Fwdm interface {
	Register(fwdID string, clientID string, handler func(wrapper *Wrapper) error)
	Unregister(inputs ...string)
	AddNewForwarder(id string) *Forwarder
	RemoveForwarder(id string)
	GetForwarder(id string) *Forwarder
	Push(id string, wrapper *Wrapper)
	GetKeys() []string
	Close()
}

// // Fwd linter
// type Fwd interface {
// 	Close()
// 	Push(wrapper *Wrapper)
// 	Register(clientID string, handler func(wrapper *Wrapper) error)
// 	UnRegister(clientID string)
// }
