package cmd

import (
	"encoding/json"
	"errors"
	"os/exec"
)

func Query(root string) []Result {
}

func GetSockName() (string, error) {
	cmd := exec.Command("watchman", "get-sockname")
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
	err := m["error"]
	if err != nil {
		return errors.New(err.(string))
	}

	return nil
}
