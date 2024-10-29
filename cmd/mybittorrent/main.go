package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

type MetaInfo struct {
	Name        string `bencode:"name"`
	Pieces      string `bencode:"pieces"`
	Length      int64  `bencode:"length"`
	PieceLength int64  `bencode:"piece length"`
}

type Meta struct {
	Announce string   `bencode:"announce"`
	Info     MetaInfo `bencode:"info"`
}

func main() {
	command := os.Args[1]
	switch command {
	case "decode":
		val := os.Args[2]

		decoded, err := bencode.Decode(strings.NewReader(val))
		if err != nil {
			panic(err)
		}
		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonOutput))
	case "info":
		fileName := os.Args[2]
		file, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		var meta Meta
		if err := bencode.Unmarshal(file, &meta); err != nil {
			panic(err)
		}
		fmt.Println("Tracker URL:", meta.Announce)
		fmt.Println("Length:", meta.Info.Length)

		h := sha1.New()
		if err := bencode.Marshal(h, meta.Info); err != nil {
			panic(err)
		}
		fmt.Printf("Info Hash: %x\n", h.Sum(nil))
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
