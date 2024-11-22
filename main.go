package main

import (
	"fmt"
	"os"

	"github.com/kromiii/tbls-ask-agent-slack/cmd/embeddings"
	"github.com/kromiii/tbls-ask-agent-slack/cmd/server"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: tbls-ask-bot [server|embeddings]")
		return
	}

	switch args[0] {
	case "server":
		server.Run()
	case "embeddings":
		embeddings.Run()
	default:
		fmt.Println("Invalid command. Usage: tbls-ask-bot [server|embeddings]")
	}
}
