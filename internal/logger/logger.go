package logger

import (
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"io"
	"log"
	"os"
)

type proxyWriter struct {
	writer io.Writer
}

func (w *proxyWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)

	return w.writer.Write(p)
}

func InitLogger(outputAddr string) {
	writer, err := gelf.NewTCPWriter(outputAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(&proxyWriter{writer: writer})
}
