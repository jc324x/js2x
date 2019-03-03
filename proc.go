package js2x

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jychri/goku/brf"
)

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
		b.WriteString(brf.LowerKebab(s))
		b.WriteString("-1)\n")
		b.WriteString("=====\n")
	case "-":
		b.WriteString("* [")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(brf.LowerKebab(s))
		b.WriteString(")\n")
	case "--":
		b.WriteString("  * [")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(brf.LowerKebab(s))
		b.WriteString(")\n")
	case "---":
		b.WriteString("   * [")
		b.WriteString(s)
		b.WriteString("](#")
		b.WriteString(brf.LowerKebab(s))
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

func (sm *Summary) writeToBuffer(p *Process) {

	p.SummaryInput = brf.Trim(p.LineInput)
	p.SummaryOutput = brf.Trim(p.LineOutput)

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
