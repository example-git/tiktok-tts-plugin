package main

import (
	"context"
	"os"

	tiktoktts "github.com/example-git/tiktok-tts-plugin"
)

func main() {
	if err := tiktoktts.Serve(context.Background(), nil, os.Stdin, os.Stdout); err != nil {
		os.Exit(1)
	}
}
