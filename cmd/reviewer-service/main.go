package main

import (
	"fmt"
	"reviewer-service/internal/config"
)

func main() {
	config := config.MustLoafConfig()

	fmt.Printf("config: %+v\n", config)
}
