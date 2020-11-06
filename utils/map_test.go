package utils

import (
	"fmt"
	"time"
)

func TestMap() {
	adMap := NewAdvanceMap()

	// go func() {
	// 	for {
	// 		time.Sleep(1 * time.Second)
	// 		fmt.Println("Len map before", adMap.Len())
	// 		fmt.Println("Capture map: ", adMap.Capture())
	// 		fmt.Println("Len map after", adMap.Len())
	// 	}
	// }()

	for i := 0; i <= 5; i++ {
		time.Sleep(1 * time.Second)
		adMap.Set(GenerateID(), GenerateID())
	}

	fmt.Println("Len map before", adMap.Len())
	fmt.Println("Capture map: ", adMap.Capture())
	fmt.Println("Len map after", adMap.Len())

	// for i := 0; i <= 5; i++ {
	// 	time.Sleep(1 * time.Second)
	// 	fmt.Println("Len map before", adMap.Len())
	// 	fmt.Println("Capture map: ", adMap.Capture())
	// 	fmt.Println("Len map after", adMap.Len())
	// }

	// go func() {
	// 	for {
	// 		time.Sleep(1 * time.Second)
	// 		adMap.Set(GenerateID(), GenerateID())
	// 	}
	// }()
	// go func() {
	// 	for {
	// 		time.Sleep(1000 * time.Millisecond)
	// 		adMap.Set(GenerateID(), GenerateID())
	// 	}
	// }()
	select {}
}
