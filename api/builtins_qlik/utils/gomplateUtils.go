package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v3"
)

func RunGomplate(dataSource string, pwd string, env []string, template string, logger *log.Logger) ([]byte, error) {
	var opts gomplate.Config
	opts.DataSources = []string{fmt.Sprintf("data=%s", filepath.Join(pwd, dataSource))}
	opts.Input = template
	opts.LDelim = "(("
	opts.RDelim = "))"

	for _, envVar := range env {
		if envVarParts := strings.Split(envVar, "="); len(envVarParts) == 2 {
			if err := os.Setenv(envVarParts[0], envVarParts[1]); err != nil {
				logger.Printf("error setting env variable: %v=%v, error: %v\n", envVarParts[0], envVarParts[1], err)
			}
		}
	}

	var buffer bytes.Buffer
	opts.Out = &buffer
	if err := gomplate.RunTemplates(&opts); err != nil {
		logger.Printf("error calling gomplate API with config: %v, error: %v\n", opts.String(), err)
		return nil, err
	}
	return buffer.Bytes(), nil
}
