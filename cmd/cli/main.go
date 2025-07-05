package main

import (
	"fmt"
	"rss-reader/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		fmt.Printf("Failed to launch: %v", err)
		return
	}
}
