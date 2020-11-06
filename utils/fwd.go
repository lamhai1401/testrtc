package utils

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/gologs/logs"
	log "github.com/lamhai1401/gologs/logs"
	"github.com/pion/rtp"
)

// Wrapper linter
type Wrapper struct {
	Pkg    rtp.Packet // save rtp packet
	Data   []byte     `json:"rtp"`    // packet to write
	Kind   string     `json:"kind"`   // audio or video
	SeatID int        `json:"seatID"` // stream id number 1-2-3-4
	Type   string     `json:"type"`   // type off wrapper data - ok - ping - pong
}

// Forwarder linter
type Forwarder struct {
	id          string                                  // stream id
	isClosed    bool                                    // makesure is closed
	clients     *AdvanceMap                             // clientID - channel
	handlers    map[string]func(wrapper *Wrapper) error // to save handler
	actionChann chan *action                            // handle action add and remove, close
	msgChann    chan *Wrapper
	mutex       sync.RWMutex
}

// NewForwarder return new forwarder
func NewForwarder(id string) *Forwarder {
	f := &Forwarder{
		id:          id,
		actionChann: make(chan *action, 100),
		clients:     NewAdvanceMap(),
		handlers:    make(map[string]func(wrapper *Wrapper) error),
		isClosed:    false,
		msgChann:    make(chan *Wrapper),
	}

	f.serve()
	return f
}

// Close linter
func (f *Forwarder) Close() {
	if f.checkClose() {
		return
	}
	if chann := f.getActionChann(); chann != nil {
		close := "close"
		a := &action{
			action: &close,
		}
		chann <- a
	}
}

// Push new wrapper to server chan
func (f *Forwarder) Push(wrapper *Wrapper) {
	if f.checkClose() {
		f.info("fwd was closed")
		return
	}
	if chann := f.getMsgChann(); chann != nil {
		chann <- wrapper
	}
}

// Register new client
func (f *Forwarder) Register(clientID string, handler func(wrapper *Wrapper) error) {
	if chann := f.getActionChann(); chann != nil && !f.checkClose() {
		add := "add"
		chann <- &action{
			id:      &clientID,
			action:  &add,
			handler: handler,
		}
	}
}

func (f *Forwarder) addNewClient(clientID string, handler func(wrapper *Wrapper) error) {
	if f.checkClose() {
		f.info("fwd was closed")
		return
	}

	// remove client if exist
	if chann := f.getClient(clientID); chann != nil {
		f.UnRegister(clientID)
	}

	f.setClient(clientID, make(chan *Wrapper, 1000))
	f.setHandler(clientID, handler)

	go f.collectData(clientID)
}

func (f *Forwarder) collectData(clientID string) {
	var handler func(w *Wrapper) error
	var err error
	chann := f.getData(clientID)

	for {
		if f.checkClose() {
			f.info("fwd was closed")
			return
		}
		w, open := <-chann
		if !open {
			// fmt.Println("out collectData")
			return
		}

		handler = f.getHandler(clientID)
		if handler == nil {
			f.info(fmt.Sprintf("%s handler is nil. Close for loop", clientID))
			return
		}

		if err = handler(&w); err != nil {
			log.Error(fmt.Sprintf("%s handler err: %v", clientID, err))
			return
		}

		w = Wrapper{} // clear mem
		handler = nil
		err = nil
	}
}

func (f *Forwarder) getData(clientID string) <-chan Wrapper {
	c := make(chan Wrapper, 1000)
	// var dumpWrapper *Wrapper

	// dumpWrapper := &Wrapper{
	// 	Kind: "string",
	// 	Type: "string",
	// }

	// parent := context.Background()
	// timeout := 3 * time.Second
	// var ctx context.Context
	// var cancel context.CancelFunc

	chann := f.getClient(clientID)

	go func() {
		defer close(c)
		// for {
		// 	ctx, cancel = context.WithTimeout(parent, timeout)
		// 	select {
		// 	case w, open := <-chann:
		// 		if !open {
		// 			// fmt.Println("out getData")
		// 			return
		// 		}
		// 		c <- *w
		// 		dumpWrapper = w
		// 		break
		// 	case <-ctx.Done():
		// 		if dumpWrapper != nil {
		// 			c <- *dumpWrapper
		// 		}
		// 		break
		// 	}
		// 	cancel()
		// 	ctx = nil
		// 	cancel = nil
		// }
		var w *Wrapper
		var open bool
		for {
			w, open = <-chann
			if !open {
				logs.Info(fmt.Sprintf("%s channel was closed", clientID))
				return
			}
			c <- *w
			w = nil
		}
	}()

	return c
}

// UnRegister linter
func (f *Forwarder) UnRegister(clientID string) {
	if f.checkClose() {
		return
	}

	if chann := f.getActionChann(); chann != nil {
		remove := "remove"
		a := &action{
			action: &remove,
			id:     &clientID,
		}
		chann <- a
	}
}

// transfer old fwd to new fwd
func (f *Forwarder) transfer(fw *Forwarder) {
	if f.checkClose() {
		f.info("fwd was closed")
		return
	}

	if clients := f.getClients(); clients != nil {
		tmp := clients.Capture()

		for k, v := range tmp {
			handler, ok := v.(func(wrapper *Wrapper) error)
			if ok {
				fw.setClient(k, make(chan *Wrapper, 1000))
				fw.setHandler(k, handler)
				fw.collectData(k)
			}
		}
	}
}
