package js2x

// Processes collects Process structs
type Processes []Process

// Run runs actions for all Process structs in Proceses
func (pcs Processes) Run(sm *Summary) {
	for i := 0; i < len(pcs); i++ {
		pcs[i].runProcess(sm)
	}
}
