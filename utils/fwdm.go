package utils

import (
	"sync"

	"github.com/lamhai1401/gologs/logs"
)

type action struct {
	id      *string
	action  *string
	handler func(wrapper *Wrapper) error
	wg      *sync.WaitGroup
}

// ForwarderMannager controll all forwadrder manager
type ForwarderMannager struct {
	id          string       // name of forwader audio or video
	forwadrders *AdvanceMap  // manage all forwader
	actionChann chan *action // handle action add and remove
	isClosed    bool
	wg          sync.WaitGroup
	mutex       sync.RWMutex
}

// NewForwarderMannager create audio or video forwader
func NewForwarderMannager(id string) Fwdm {
	f := &ForwarderMannager{
		id:          id,
		forwadrders: NewAdvanceMap(),
		isClosed:    false,
		actionChann: make(chan *action, 100),
	}
	go f.serve()
	return f
}

// Close lstringer
func (f *ForwarderMannager) Close() {
	f.close()
}

// AddNewForwarder to add new forwarder with id
func (f *ForwarderMannager) AddNewForwarder(id string) *Forwarder {
	f.addAction(id, "add", &f.wg)
	return f.getForwarder(id)
}

// RemoveForwarder remove forwader with id
func (f *ForwarderMannager) RemoveForwarder(id string) {
	f.addAction(id, "remove", &f.wg)
}

// Push to wrapper to specific id
func (f *ForwarderMannager) Push(id string, wrapper *Wrapper) {
	if f.checkClose() {
		return
	}

	forwardfer := f.getForwarder(id)
	if forwardfer == nil {
		logs.Warn(id, " forwarder is nil. Cannot push")
		return
		// forwardfer = f.AddNewForwarder(id)
	}
	forwardfer.Push(wrapper)
}

// SetForwarder linter
func (f *ForwarderMannager) SetForwarder(id string, fw *Forwarder) {
	f.setForwarder(id, fw)
}

// GetForwarder get forwarder of this id is exist or not
func (f *ForwarderMannager) GetForwarder(id string) *Forwarder {
	return f.getForwarder(id)
}

// Unregister unregis clientId to specific forwarder
func (f *ForwarderMannager) Unregister(inputs ...string) {
	if forwardfer := f.getForwarder(inputs[0]); forwardfer != nil {
		forwardfer.UnRegister(inputs[1])
	}
}

// Register regis a client id to specific forwarder
func (f *ForwarderMannager) Register(id string, clientID string, handler func(wrapper *Wrapper) error) {
	if f.checkClose() {
		return
	}

	forwardfer := f.getForwarder(id)
	if forwardfer == nil {
		forwardfer = f.AddNewForwarder(id)
	}
	forwardfer.Register(clientID, handler)
}

// GetKeys return id of all forwarder
func (f *ForwarderMannager) GetKeys() []string {
	return f.getForwarders().GetKeys()
}
