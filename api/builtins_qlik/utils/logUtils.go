package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type nullWriter struct {
}

func (nw *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func GetLogger(pluginName string) *log.Logger {
	var logWriter io.Writer
	if os.Getenv("QKP_LOG_STDOUT_ENABLED") == "true" {
		logWriter = os.Stdout
	} else if os.Getenv("QKP_LOG_ENABLED") == "true" {
		tmpFile, err := ioutil.TempFile(os.Getenv("QKP_LOG_DIR"), fmt.Sprintf("%v-*.log", pluginName))
		if err != nil {
			panic(fmt.Errorf("error initializing logging for plugin: %v, error: %v", pluginName, err))
		}
		logWriter = tmpFile
	} else {
		logWriter = &nullWriter{}
	}
	return log.New(logWriter, "", log.LstdFlags|log.LUTC|log.Lmicroseconds|log.Lshortfile)
}
