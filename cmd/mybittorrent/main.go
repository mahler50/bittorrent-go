package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decode(s string, st int) (x any, i int, err error) {
	if st == len(s) {
		return nil, st, io.ErrUnexpectedEOF
	}

	i = st
	switch {
	case s[i] == 'l':
		return decodeList(s, i)
	case s[i] == 'i':
		return decodeInt(s, i)
	case unicode.IsDigit(rune(s[i])):
		return decodeString(s, i)
	case s[i] == 'd':
		return decodeDictionaries(s, i)
	default:
		return nil, st, fmt.Errorf("unexpected err at: %q", s[i])
	}
}

// decode string <length>:<string>
func decodeString(s string, st int) (x string, i int, err error) {
	var firstColonIndex int
	i = st

	for ; i < len(s); i++ {
		if s[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := s[st:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", st, fmt.Errorf("bad string length")
	}

	// skip ':'
	i++
	if i+length > len(s) {
		return "", st, fmt.Errorf("bad string: out of bounds")
	}

	return s[i : i+length], i + length, nil
}

// decode integer i<integer>e
func decodeInt(s string, st int) (x int, i int, err error) {
	i = st
	// skip 'i'
	i++

	for ; i < len(s); i++ {
		if s[i] == 'e' {
			break
		}
	}
	x, err = strconv.Atoi(s[st+1 : i])
	if err != nil {
		return 0, st, fmt.Errorf("bad integer")
	}
	// skip 'e'
	i++

	return x, i, nil
}

// decode list l<bencoded_elements>e
func decodeList(s string, st int) (l []any, i int, err error) {
	i = st
	// skip 'l'
	i++

	l = []any{}

	for {
		if i >= len(s) {
			return nil, st, fmt.Errorf("bad list")
		}

		if s[i] == 'e' {
			break
		}

		var x any
		x, i, err = decode(s, i)
		if err != nil {
			return nil, i, err
		}

		l = append(l, x)
	}
	// skip 'e'
	i++

	return l, i, nil
}

// decode dictionaries d<key1><value1>...<keyN><valueN>e
// key must be bencoded string and value could be any bencoded element
func decodeDictionaries(s string, st int) (dict map[string]any, i int, err error) {
	if st == len(s) {
		return nil, st, io.ErrUnexpectedEOF
	}

	if s[st] != 'd' {
		return nil, st, fmt.Errorf("dictionaries expected")
	}
	i = st
	// skip 'd'
	i++

	dict = map[string]any{}

	for {
		if i >= len(s) {
			return nil, st, fmt.Errorf("bad dictionaries")
		}

		if s[i] == 'e' {
			break
		}

		var key string
		key, i, err = decodeString(s, i)
		if err != nil {
			return nil, i, err
		}

		var value any
		value, i, err = decode(s, i)
		if err != nil {
			return nil, i, err
		}

		dict[key] = value
	}
	// skip 'e'
	i++

	return dict, i, nil
}

// print metainfo map
func printMetaInfo(dict map[string]any) {
	fmt.Printf("Tracker URL: %v\n", dict["announce"])

	info, ok := dict["info"].(map[string]any)
	if info == nil || !ok {
		log.Fatalf("No info section")
	}

	fmt.Printf("Length: %v\n", info["length"])
}

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		decoded, idx, err := decode(os.Args[2], 0)
		if err != nil {
			log.Fatalf("error: %v at %d\n", err, idx)
		}

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			log.Fatalf("error: encode to json%v\n", err)
		}
		fmt.Println(string(jsonOutput))
	case "info":
		bytes, err := os.ReadFile(os.Args[2])
		if err != nil {
			log.Fatalf("error: read file %v\n", err)
		}

		dict, idx, err := decodeDictionaries(string(bytes), 0)
		if err != nil {
			log.Fatalf("error: %v at %d\n", err, idx)
		}
		printMetaInfo(dict)
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
