// Package js2x is the main package.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
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

// validatePath returns a path string; ~ is expanded to /User/user and trailing slashes are removed
func validatePath(p string) string {
	if t := strings.TrimPrefix(p, "~/"); t != p {
		u, err := user.Current()

		if err != nil {
			log.Fatalf("Unable to identify the current user")
		}

		t := strings.Join([]string{u.HomeDir, "/", t}, "")
		return strings.TrimSuffix(t, "/")
	}
	return strings.TrimSuffix(p, "/")
}

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

func lowerKebab(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, " ", "-", -1)
	return s
}

func trimLine(s string) string {
	ts := strings.TrimSpace(s)
	ts = strings.TrimSuffix(s, "\n")
	return ts
}

func (c *Config) getProcesses() (ps []Process) {
	for i := 0; i < len(c.Processes); i++ {
		cp := c.Processes[i]
		var p Process

		p.Target = validateTarget(cp.Target)
		p.InputPath = validatePath(cp.Input)
		p.OutputPath = validatePath(cp.Output)
		p.InputFile = validateInputPath(p.InputPath)

		ps = append(ps, p)
	}
	return ps
}

func (c *Config) getSummary() (sm Summary) {
	sm.OutputPath = validatePath(c.Summary)
	return sm
}

// Process ...
type Process struct {
	InputPath     string
	OutputPath    string
	Target        string
	InputFile     *os.File
	LinePrevious  string
	LineInput     string
	LineHeight    string
	LineOutput    string
	LineIndex     int
	SummaryInput  string
	SummaryOutput string
	Section       string
	Subsection    string
	Mode          string
	Buff          bytes.Buffer
}

func (p *Process) runProcess(sm *Summary) {
	scanner := bufio.NewScanner(p.InputFile)
	for scanner.Scan() {
		p.setLineInput(scanner.Text())
		p.setSection()
		p.setSubsection()
		p.writeToBuffer()
		sm.writeToBuffer(p)
	}
	p.writeToFile()
}

func (p *Process) setLineInput(s string) {
	p.LinePrevious = p.LineInput
	p.LineInput = s
	p.LineIndex++
}

func (p *Process) setSection() {
	if strings.Contains(p.LineInput, "!=== SKIP") {
		p.Section = "SKIP"
	}

	if strings.Contains(p.LineInput, "!=== NAV") {
		p.Section = "NAV"
	}

	if strings.Contains(p.LineInput, "!=== MAIN") {
		p.Section = "MAIN"
	}

	if strings.Contains(p.LineInput, "!=== DIRECT") {
		p.Section = "DIRECT"
	}
}

func (p *Process) setSubsection() {

	if strings.Contains(p.LineInput, "!===") {
		p.Subsection = "HEADER"
	} else if strings.Contains(p.LineInput, "===!") {
		p.Subsection = "FOOTER"
	} else {
		rgx := regexp.MustCompile(`#|-+`)

		switch p.Section {
		case "MAIN":

			if p.Subsection == "HEADER" && p.LineInput == "" {
				p.Subsection = "BLANK"
				p.LineHeight = ""
			}

			if strings.Contains(p.LineInput, "#") {
				p.Subsection = "LINK"
				p.LineHeight = rgx.FindString(p.LineInput)
			}

			if p.Subsection != "FUNC" && strings.Contains(p.LineInput, "// -") {
				p.Subsection = "LINK"
				p.LineHeight = rgx.FindString(p.LineInput)
			}

			if p.Subsection == "LINK" && p.LineInput == "" {
				p.Subsection = "BLANK"
				p.LineHeight = ""
			}

			if p.Subsection == "JSDOC_START" && p.LineInput != "" {
				p.Subsection = "JSDOC"
				p.LineHeight = ""
			}

			if p.Subsection == "BLANK" && strings.Contains(p.LineInput, "/**") {
				p.Subsection = "JSDOC_START"
			}

			if strings.Contains(p.LineInput, "*/") {
				p.Subsection = "JSDOC_END"
			}

			if p.Subsection == "JSDOC_END" && p.LineInput == "" {
				p.Subsection = "BLANK"
			}

			if p.Subsection == "FUNC_START" {
				p.Subsection = "FUNC"
			}

			if strings.Contains(p.LineInput, "function") {
				p.Subsection = "FUNC_START"
			}

			if p.Subsection == "FUNC" && p.LineInput == "}" {
				p.Subsection = "FUNC_END"
			}

			if p.Subsection == "FUNC_END" && p.LineInput == "" {
				p.Subsection = "BLANK"
				p.LineHeight = ""
			}

			if p.Subsection == "EX_START" && p.LineInput != "" {
				p.Subsection = "EX"
			}

			if p.Subsection == "BLANK" && strings.Contains(p.LineInput, "Logger.log") {
				p.Subsection = "EX_START"
			}

			if strings.Contains(p.LineInput, "!EX") {
				p.Subsection = "EX_END"
				p.LineHeight = ""
			}

			if p.Subsection == "EX_END" && p.LineInput == "" {
				p.Subsection = "BLANK"
			}

			if strings.Contains(p.LineInput, "===!") {
				p.Subsection = "FOOTER"
				p.LineHeight = ""
			}

		case "NAV":

			p.Subsection = "LINK"
			p.LineHeight = rgx.FindString(p.LineInput)

		case "SKIP":
			p.Subsection = "SKIP"
			p.LineHeight = ""

		case "DIRECT":
			if strings.Contains(p.LinePrevious, "!=== DIRECT") && strings.Contains(p.LinePrevious, p.Target) {
				p.Subsection = p.Target
			}
		}
	}
}

func (p *Process) writeToBuffer() {

	if p.Target == "README" {
		switch {
		case p.Section == "DIRECT" && p.Subsection == "README":
			p.lineWriteBufferDirect()
		case p.Subsection == "HEADER" || p.Subsection == "FOOTER":
			p.lineSkipBuffer()
		case p.Section == "NAV" && p.Subsection == "LINK":
			p.markdownNavLink()
		case p.Section == "MAIN" && p.Subsection == "LINK":
			p.markdownMainLink()
		case p.Subsection == "JSDOC_START":
			p.markdownJSDocStart()
		case p.Subsection == "EX_START" || p.Subsection == "EX":
			p.markdownEx()
		case p.Subsection == "EX_END":
			p.markdownExEnd()
		default:
			p.lineWriteBuffer()
		}

	} else if p.Target == "LIBRARY" {
		switch {
		case p.Section == "DIRECT" && p.Subsection == "LIBRARY":
			p.lineWriteBufferDirect()
		case p.Section == "MAIN" && p.Subsection == "BLANK":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "LINK":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "JSDOC_START":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "JSDOC":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "JSDOC_END":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "FUNC_START":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "FUNC":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "FUNC_END":
			p.lineWriteBuffer()
		default:
			p.lineSkipBuffer()
		}
	} else if p.Target == "QUICK-START" {
		switch {
		case p.Section == "DIRECT" && p.Subsection == "QUICK-START":
			p.lineWriteBufferDirect()
		case p.Section == "MAIN" && p.Subsection == "BLANK":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "LINK":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "JSDOC_START":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "JSDOC":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "JSDOC_END":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "FUNC_START":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "FUNC":
			p.lineWriteBuffer()
		case p.Section == "MAIN" && p.Subsection == "FUNC_END":
			p.lineWriteBuffer()
		case p.Subsection == "EX_START" || p.Subsection == "EX":
			p.lineWriteBuffer()
		case p.Subsection == "EX_END":
			p.quickstartExEnd()
		default:
			p.lineSkipBuffer()
		}
	}
}

func (p *Process) lineSkipBuffer() {
	p.LineOutput = ""
	p.LineHeight = ""
}

func (p *Process) lineWriteBufferDirect() {
	s := p.LineInput
	cs := strings.TrimSuffix(p.LineOutput, "\n")

	if cs == "" && s == "" {
	} else {
		s = strings.TrimPrefix(s, "// ")
		var b bytes.Buffer
		b.WriteString(s)
		b.WriteString("\n")
		p.LineOutput = b.String()
		p.Buff.WriteString(p.LineOutput)
	}
}

func (p *Process) lineWriteBuffer() {
	s := p.LineInput
	cs := strings.TrimSuffix(p.LineOutput, "\n")

	if cs == "" && s == "" {
	} else {
		var b bytes.Buffer
		b.WriteString(s)
		b.WriteString("\n")
		p.LineOutput = b.String()
		p.Buff.WriteString(p.LineOutput)
	}
}

func (p *Process) markdownNavLink() {
	s := strings.TrimPrefix(p.LineInput, "// | | ")

	strs := strings.SplitAfter(s, p.LineHeight)

	if len(strs) >= 1 {
		s = strs[1]
	}

	s = strings.TrimPrefix(s, " ")

	var b bytes.Buffer

	switch p.LineHeight {
	case "#":
		b.WriteString("\n")
		b.WriteString("[")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(lowerKebab(s))
		b.WriteString("-1)\n")
		b.WriteString("=====\n")
	case "-":
		b.WriteString("* [")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(lowerKebab(s))
		b.WriteString(")\n")
	case "--":
		b.WriteString("  * [")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(lowerKebab(s))
		b.WriteString(")\n")
	case "---":
		b.WriteString("   * [")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(lowerKebab(s))
		b.WriteString(")\n")
	}

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) markdownMainLink() {
	s := strings.TrimPrefix(p.LineInput, "// ")

	strs := strings.SplitAfter(s, p.LineHeight)

	if len(strs) >= 1 {
		s = strs[1]
	}

	s = strings.TrimPrefix(s, " ")

	var b bytes.Buffer

	switch p.LineHeight {
	case "#":
		b.WriteString("## ")
		b.WriteString(s)
		b.WriteString(" ##\n")
	case "-":
		b.WriteString("### ")
		b.WriteString(s)
		b.WriteString(" ###\n")
	case "--":
		b.WriteString("#### ")
		b.WriteString(s)
		b.WriteString(" ####\n")
	case "---":
		b.WriteString("##### ")
		b.WriteString(s)
		b.WriteString(" #####\n")
	}

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) markdownJSDocStart() {
	s := p.LineInput

	var b bytes.Buffer
	b.WriteString("```javascript\n")
	b.WriteString(s)
	b.WriteString("\n")

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) markdownEx() {
	s := p.LineInput

	if strings.Contains(s, "Logger.log") {
		s = strings.TrimPrefix(s, "// ")
	}

	if strings.Contains(s, "var") {
		s = strings.TrimPrefix(s, "// ")
	}

	var b bytes.Buffer

	b.WriteString(s)
	b.WriteString("\n")

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) markdownExEnd() {
	s := p.LineInput

	switch {
	case strings.Contains(s, "// Logger.log"):
		s = strings.TrimPrefix(s, "// ")
	case strings.Contains(s, "// var"):
		s = strings.TrimPrefix(s, "// ")
	}

	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "//!EX")

	var b bytes.Buffer
	b.WriteString(s)
	b.WriteString("\n")
	b.WriteString("```\n")

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) quickstartEx() {
	s := p.LineInput

	if strings.Contains(s, "Logger.log") {
		s = strings.TrimPrefix(s, "// ")
	}

	if strings.Contains(s, "var") {
		s = strings.TrimPrefix(s, "// ")
	}

	var b bytes.Buffer

	b.WriteString(s)
	b.WriteString("\n")

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) quickstartExEnd() {
	s := p.LineInput
	s = strings.TrimSuffix(s, "//!EX")

	var b bytes.Buffer
	b.WriteString(s)
	b.WriteString("\n")

	p.LineOutput = b.String()
	p.Buff.WriteString(p.LineOutput)
}

func (p *Process) writeToFile() {
	var r io.Reader
	r = &p.Buff
	file, err := os.Create(p.OutputPath)

	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(file, r); err != nil {
		log.Fatal(err)
	}
}

// Summary ...
type Summary struct {
	LineIndex  int
	OutputPath string
	Buff       bytes.Buffer
}

func (sm *Summary) writeToBuffer(p *Process) {

	p.SummaryInput = trimLine(p.LineInput)
	p.SummaryOutput = trimLine(p.LineOutput)

	var b bytes.Buffer

	if sm.LineIndex == 0 {
		b.WriteString("TOTAL, LINE, TARGET, SECTION, SUBSECTION,")
		b.WriteString("HEIGHT, INPUT, OUTPUT\n")
	}

	sm.LineIndex++

	b.WriteString(strconv.Itoa(sm.LineIndex))
	b.WriteString(",")
	b.WriteString(strconv.Itoa(p.LineIndex))
	b.WriteString(",")
	b.WriteString(p.Target)
	b.WriteString(",")
	b.WriteString(p.Section)
	b.WriteString(",")
	b.WriteString(p.Subsection)
	b.WriteString(",")

	if p.LineHeight != "" {
		b.WriteString(p.LineHeight)
	} else {
		b.WriteString("N/A")
	}
	b.WriteString(",")

	if p.SummaryInput != "" {
		b.WriteString(`"`)
		b.WriteString(p.SummaryInput)
		b.WriteString(`"`)
		b.WriteString(",")
	} else {
		b.WriteString("N/A,")
	}

	b.WriteString(`"`)
	b.WriteString(p.SummaryOutput)
	b.WriteString(`"`)
	b.WriteString("\n")

	sm.Buff.WriteString(b.String())
}

func (sm *Summary) writeToFile() {
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

func runProcesses(ps []Process, sm *Summary) {
	for i := 0; i < len(ps); i++ {
		ps[i].runProcess(sm)
	}
}

// --> main fns

func initializeRun() (ps []Process, sm Summary) {
	u, err := user.Current()

	if err != nil {
		log.Fatalf("Fatal: Can't identify the current user")
	}

	j := strings.Join([]string{u.HomeDir, "/.js2x.json"}, "")

	r, err := ioutil.ReadFile(j)

	if err != nil {
		log.Fatalf("Fatal: Can't read %v/.js2x.json", u.HomeDir)
	}

	var c Config

	if err := json.Unmarshal(r, &c); err != nil {
		log.Fatalf("Fatal: Can't unmarshal %v/.js2x.json", u.HomeDir)
	}

	ps = c.getProcesses()
	sm = c.getSummary()

	return ps, sm
}

func main() {
	ps, sm := initializeRun()
	runProcesses(ps, &sm)
	sm.writeToFile()
}
