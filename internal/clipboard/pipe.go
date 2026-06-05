package clipboard

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func pipeToCommand(bin string, args []string, data any) error {
	path, err := exec.LookPath(bin)
	if err != nil {
		return err
	}
	cmd := exec.Command(path, args...)
	switch v := data.(type) {
	case string:
		cmd.Stdin = strings.NewReader(v)
	case []byte:
		cmd.Stdin = bytes.NewReader(v)
	default:
		return fmt.Errorf("unsupported pipe data type")
	}
	if err := cmd.Run(); err != nil {
		log.Printf("clipboard (%s): %v", bin, err)
		return err
	}
	return nil
}

func outputFromCommand(bin string, args []string) ([]byte, error) {
	path, err := exec.LookPath(bin)
	if err != nil {
		return nil, err
	}
	// Read failures are routine while polling (e.g. image type absent); callers decide whether to log.
	out, err := exec.Command(path, args...).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func haveCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
