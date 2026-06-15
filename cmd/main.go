package main

import (
	"github.com/richd0tcom/yadm/interrnal/config"
	"github.com/richd0tcom/yadm/interrnal/storage"
)

// "fmt"
// "os"


func main() {
    // if len(os.Args) < 2 {
    //     fmt.Println("Usage: tracker <command>")
    //     os.Exit(1)
    // }
    
    // command := os.Args[1]
    
    // switch command {
    // case "snapshot":
    //     // handleSnapshot()
	// 	break
    // case "list":
    //     // handleList()
	// 	break
    // default:
    //     fmt.Printf("Unknown command: %s\n", command)
    //     os.Exit(1)
    // }

	config.LoadConfig("~/.dotfile-tracker/config.toml")

    storage.NewBlobStore("~/.dotfile-tracker/")
}