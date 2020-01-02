package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func RunGomplate(dataSource string, pwd string, env []string, temp string, logger *log.Logger) ([]byte, error) {
	dataLocation := filepath.Join(pwd, dataSource)
	data := fmt.Sprintf(`--datasource=data=%s`, dataLocation)
	from := fmt.Sprintf(`--in=%s`, temp)

	gomplateCmd := exec.Command("gomplate", `--left-delim=((`, `--right-delim=))`, data, from)

	gomplateCmd.Env = append(os.Environ(), env...)

	var out bytes.Buffer
	gomplateCmd.Stdout = &out
	err := gomplateCmd.Run()

	if err != nil {
		logger.Printf("error executing command: %v with args: %v, error: %v\n", gomplateCmd.Path, gomplateCmd.Args, err)
		return nil, err
	}
	return out.Bytes(), nil
}
