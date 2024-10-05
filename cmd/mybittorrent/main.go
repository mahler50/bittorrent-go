package main

import (
	"encoding/json"
	"fmt"
	"io"
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	//fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		decoded, idx, err := decode(os.Args[2], 0)
		if err != nil {
			fmt.Printf("error: %v at %d\n", err, idx)
			return
		}

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Printf("error: encode to json%v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
