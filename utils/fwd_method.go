package utils

import (
	"fmt"

	"github.com/lamhai1401/gologs/logs"
	log "github.com/lamhai1401/gologs/logs"
)

func (f *Forwarder) getID() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.id
}

func (f *Forwarder) checkClose() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.isClosed
}

func (f *Forwarder) setClose(state bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.isClosed = state
}

// close to close all serve
func (f *Forwarder) close() {
	if !f.checkClose() {
		f.setClose(true)
		f.closeClients()
		f.info(fmt.Sprintf("%s forwader was closed", f.getID()))
	}
}

// info to export log info
func (f *Forwarder) info(v ...interface{}) {
	log.Info(fmt.Sprintf("[%s] ", f.id), v)
}

// error to export error info
func (f *Forwarder) error(v ...interface{}) {
	log.Error(fmt.Sprintf("[%s] ", f.id), v)
}

func (f *Forwarder) getClient(clientID string) chan *Wrapper {
	if clients := f.getClients(); clients != nil {
		client, ok1 := clients.Get(clientID)
		if !ok1 {
			return nil
		}
		chann, ok2 := client.(chan *Wrapper)
		if ok2 {
			return chann
		}
	}
	return nil
}

func (f *Forwarder) setClient(clientID string, chann chan *Wrapper) {
	if clients := f.getClients(); clients != nil {
		clients.Set(clientID, chann)
	}
}

func (f *Forwarder) deleteClient(clientID string) {
	if clients := f.getClients(); clients != nil {
		clients.Delete(clientID)
	}
}

func (f *Forwarder) closeClient(clientID string) {
	if client := f.getClient(clientID); client != nil {
		f.deleteClient(clientID)
		f.deleteHandler(clientID)
		close(client)
		client = nil
		log.Info(fmt.Sprintf("Remove id %s from Forwarder id: %s done", clientID, f.getID()))
	}
}

func (f *Forwarder) closeClients() {
	if clients := f.getClients(); clients != nil {
		keys := clients.GetKeys()
		for _, key := range keys {
			f.closeClient(key)
		}
	}
}

func (f *Forwarder) getClients() *AdvanceMap {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.clients
}

func (f *Forwarder) setHandler(id string, handler func(w *Wrapper) error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.handlers[id] = handler
}

func (f *Forwarder) getHandler(id string) func(w *Wrapper) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.handlers[id]
}

func (f *Forwarder) deleteHandler(id string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.handlers, id)
}

func (f *Forwarder) getActionChann() chan *action {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.actionChann
}

// Serve to run
func (f *Forwarder) serve() {
	go func() {
		for {
			msg, open := <-f.msgChann
			if !open || f.checkClose() {
				return
			}
			f.forward(msg)
			msg = nil
		}
	}()

	go func() {
		for {
			action, open := <-f.actionChann
			if !open || f.checkClose() {
				return
			}
			switch *action.action {
			case "remove":
				f.closeClient(*action.id)
				break
			case "add":
				f.addNewClient(*action.id, action.handler)
				break
			case "close":
				f.close()
				return
			default:
				logs.Info("Nothing to do with this action: ", *action.action)
			}
		}
	}()
}

func (f *Forwarder) forward(wrapper *Wrapper) {
	if f.checkClose() {
		f.info(f.getID(), " fwd was closed")
		return
	}

	if clients := f.getClients(); clients != nil {
		clients.Iter(func(key, value interface{}) bool {
			chann, ok := value.(chan *Wrapper)
			if ok {
				chann <- wrapper
			}
			return true
		})
	}
}

func (f *Forwarder) getMsgChann() chan *Wrapper {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.msgChann
}
