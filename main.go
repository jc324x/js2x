package js2x

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jychri/goku/tilde"
)

// setTarget, need for log.Fatalf?
// validateTargets
func validateTarget(m string) string {
	us := strings.ToUpper(m)
	switch us {
	case "README":
	case "LIBRARY":
	case "QUICK-START":
	default:
		log.Fatalf("%v is not a valid target", m)
	}
	return us
}

func validateInputPath(i string) *os.File {
	if _, err := os.Stat(i); os.IsNotExist(err) {
		log.Fatalf("No file at path %v", i)
	}

	file, err := os.OpenFile(i, os.O_RDWR, 0644)

	if err != nil {
		log.Fatalf("Unable to open file at %v", i)
	}
	return file
}

// --> main fns

// Init ...
func Init() (pcs Processes, sm Summary) {
	path := tilde.Abs("~/.js2x.json")

	var content []byte
	var err error
	var conf Config

	if content, err = ioutil.ReadFile(path); err != nil {
		log.Fatalf("Fatal: Can't read ~/.js2x.json")
	}

	if err := json.Unmarshal(content, &conf); err != nil {
		log.Fatalf("Fatal: Can't unmarshal ~/.js2x.json")
	}

	pcs = conf.getProcesses()
	sm = conf.getSummary()

	return pcs, sm
}

func main() {
	ps, sm := Init()
	ps.Run(&sm)
	sm.Write()
}
