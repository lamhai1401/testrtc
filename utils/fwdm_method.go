package utils

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/gologs/logs"
)

func (f *ForwarderMannager) getActionChann() chan *action {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.actionChann
}

func (f *ForwarderMannager) checkClose() bool {
	return f.isClosed
}

func (f *ForwarderMannager) setClose(state bool) {
	f.isClosed = state
}

func (f *ForwarderMannager) getForwarder(id string) *Forwarder {
	fwd, hasFwd := f.forwadrders.Get(id)
	if !hasFwd {
		// logs.Warn("Cannot find fwd with id ", id)
		return nil
	}
	if fwd != nil {
		fwd, ok := fwd.(*Forwarder)
		if !ok {
			logs.Warn("Cannot assert type fwd with id ", id)
			return nil
		}
		return fwd
	}
	return nil
}

func (f *ForwarderMannager) serve() {
	for {
		action, open := <-f.getActionChann()
		if !open || f.checkClose() {
			return
		}
		switch *action.action {
		case "add":
			f.addNewForwarder(*action.id, action.wg)
			break
		case "remove":
			f.removeForwarder(*action.id, action.wg)
			break
		default:
			logs.Info("Nothing to do with this action", *action.action)
		}
	}
}

func (f *ForwarderMannager) addAction(id string, act string, wg *sync.WaitGroup) {
	// if the counter != zero wait until it done
	wg.Wait()
	// add new wait
	if chann := f.getActionChann(); chann != nil {
		wg.Add(1)
		chann <- &action{
			action: &act,
			id:     &id,
			wg:     wg,
		}
		wg.Wait()
	}
}

func (f *ForwarderMannager) addNewForwarder(id string, wg *sync.WaitGroup) {
	// make sure decre the counter
	defer wg.Done()

	// create new
	newForwader := NewForwarder(id)

	// get existing forwarder
	oldFwd := f.getForwarder(id)
	if oldFwd != nil {
		// transfer data old -> new
		oldFwd.transfer(newForwader)
		// remove old fwd
		f.closeForwarder(id)
	}

	f.setForwarder(id, newForwader)
	logs.Info(fmt.Sprintf("Add New %s forwader successfully", id))
}

func (f *ForwarderMannager) removeForwarder(id string, wg *sync.WaitGroup) {
	defer wg.Done()
	f.closeForwarder(id)
}

func (f *ForwarderMannager) closeForwarder(id string) {
	if fw := f.getForwarder(id); fw != nil {
		f.deleteForwarder(id)
		fw.Close()
		fw = nil
	}
}

func (f *ForwarderMannager) setForwarder(id string, fw *Forwarder) {
	if fwds := f.getForwarders(); fwds != nil {
		fwds.Set(id, fw)
	}
}

func (f *ForwarderMannager) deleteForwarder(id string) {
	if fwds := f.getForwarders(); fwds != nil {
		fwds.Delete(id)
	}
}

func (f *ForwarderMannager) getForwarders() *AdvanceMap {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.forwadrders
}

func (f *ForwarderMannager) closeForwaders() {
	if fwds := f.getForwarders(); fwds != nil {
		keys := fwds.GetKeys()
		for _, key := range keys {
			f.deleteForwarder(key)
		}
	}
}

func (f *ForwarderMannager) close() {
	f.setClose(true)
	f.closeForwaders()
}
