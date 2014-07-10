package main

import (
	"encoding/json"
	"fmt"
)
import "github.com/cstrahan/go-watchman/cmd"

import "bytes"

func main() {
	obj := fromJSON(`
	["query", "/Users/charlesstrahan/go/src/github.com/cstrahan/go-watchman", {
		"expression": ["type", "f"],
		"fields": ["name"]
	}]`)

	res, err := cmd.Command("watchman", obj)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(toJSON(res))
}

func toJSON(obj interface{}) string {
	str, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(str)
}

func fromJSON(str string) interface{} {
	d := json.NewDecoder(bytes.NewReader([]byte(str)))
	var obj interface{}
	d.Decode(&obj)
	return obj
}
