package js2x

import (
	"github.com/jychri/goku/tilde"
)

// Config is js2x.json unmarshalled
type Config struct {
	Summary   string `json:"summary"`
	Processes []struct {
		Target string `json:"target"`
		Input  string `json:"input"`
		Output string `json:"output"`
	} `json:"processes"`
}

func (c *Config) getProcesses() (ps []Process) {
	for i := 0; i < len(c.Processes); i++ {
		cp := c.Processes[i]
		var p Process

		p.Target = validateTarget(cp.Target)
		p.InputPath = tilde.Abs(cp.Input)
		p.OutputPath = tilde.Abs(cp.Output)
		p.InputFile = validateInputPath(p.InputPath)

		ps = append(ps, p)
	}
	return ps
}

func (c *Config) getSummary() (sm Summary) {
	sm.OutputPath = tilde.Abs(c.Summary)
	return sm
}
