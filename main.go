package main

import (
	"blog/internal/app"
	"blog/internal/pkg/port"
	"fmt"
)

func main() {
	aPort := port.FindAvailablePort(8089)
	router, err := app.Setup()
	if err != nil {
		panic(fmt.Sprintf("service setup failed:", err.Error()))
	}
	router.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", aPort))
	// if err != nil {
	// 	panic(fmt.Sprintf("service listen %d failed: %d", aPort, err.Error()))
	// }
}
