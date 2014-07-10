package cmd

import (
	"encoding/json"
	"errors"
	"net"
	"os/exec"

	"github.com/cstrahan/go-watchman/bser"
)

func Command(watchmanPath string, cmd interface{}) (interface{}, error) {
	sockname, err := GetSockName(watchmanPath)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", sockname)
	if err != nil {
		return nil, err
	}

	val, err := bser.Decode(conn)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func GetSockName(watchmanPath string) (string, error) {
	cmd := exec.Command(watchmanPath, "get-sockname")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Run(); err != nil {
		return "", err
	}

	d := json.NewDecoder(out)
	var obj interface{}
	err = d.Decode(&obj)
	if err != nil {
		return "", err
	}

	err = getError(obj)
	if err != nil {
		return "", err
	}

	sockname := obj.(map[string]interface{})["sockname"].(string)
	return sockname, nil
}

func getError(obj interface{}) error {
	m, ok := obj.(map[string]interface{})
	if !ok {
		errors.New("expected json object")
	}

	err := m["error"]
	if err != nil {
		return errors.New(err.(string))
	}

	return nil
}
