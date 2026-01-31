package main

import (
	"config"
	"fmt"
)

func main() {
	cfg := config.Get()
	fmt.Println(cfg)
}
