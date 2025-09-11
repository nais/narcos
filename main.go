package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/narcos/internal/application"
)

func main() {
	if err := application.Run(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
