package js2x

import (
	"bytes"
	"io"
	"log"
	"os"
)

// Summary ...
type Summary struct {
	LineIndex  int
	OutputPath string
	Buff       bytes.Buffer
}

func (sm *Summary) Write() {
	var r io.Reader
	r = &sm.Buff
	file, err := os.Create(sm.OutputPath)

	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(file, r); err != nil {
		log.Fatal(err)
	}
}
