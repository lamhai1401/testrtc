package utils

import (
	"fmt"
	"time"
)

var (
	tmp = make([]string, 0)
)

// TestFwd linter
func TestFwd() {
	fwd := NewForwarder("test")
	addRoutine(fwd)
	go writeRoutine(fwd)

	time.Sleep(10 * time.Second)
	go removeRoutine(fwd)
	select {}
}

func writeRoutine(fwd *Forwarder) {
	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {
		fwd.Push(&Wrapper{
			Kind: fmt.Sprintf("Test men %v", ticker),
		})
	}
}

func addRoutine(fwd *Forwarder) {
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%v", i)
		time.Sleep(2 * time.Second)
		fwd.Register(id, func(wrapper *Wrapper) error {
			fmt.Println(fmt.Sprintf("%s has data: %s", id, wrapper.Kind))
			return nil
		})
	}
}

func removeRoutine(fwd *Forwarder) {
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		id := fmt.Sprintf("%v", i)
		fwd.UnRegister(id)
	}
}

func TestNewFwd() {
	fwd := NewForwarder("test")

	fwd.Register("clientID", func(wrapper *Wrapper) error {
		fmt.Println(wrapper.Kind)
		return nil
	})

	go func() {
		time.Sleep(2 * time.Second)
		fwd.Push(&Wrapper{
			Kind: "hoi test nhe",
		})
		// for {
		// 	time.Sleep(1 * time.Second)
		// 	fwd.Push(&Wrapper{
		// 		Kind: "hoi test nhe",
		// 	})
		// }
	}()

	go func() {
		time.Sleep(15 * time.Second)
		fwd.UnRegister("clientID")
	}()
	select {}
}
