package util

import (
	"fmt"
	"github.com/weaveworks/ignite/pkg/logs"
	"os"
	"strings"
	"text/tabwriter"
)

type output struct {
	writer   *tabwriter.Writer
	isHeader bool
}

func NewOutput() *output {
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 1, '\t', 0)
	return &output{
		writer:   writer,
		isHeader: true,
	}
}

func (o *output) Write(input ...interface{}) {
	var sb strings.Builder
	for i, data := range input {
		// If the data conforms to fmt.Stringer, use .String()
		if stringer, ok := data.(fmt.Stringer); ok {
			sb.WriteString(stringer.String())
		} else {
			switch data.(type) {
			case string:
				sb.WriteString(fmt.Sprintf("%s", data))
			case int64:
				sb.WriteString(fmt.Sprintf("%d", data))
			default:
				sb.WriteString(fmt.Sprintf("%v", data))
			}
		}

		if i+1 < len(input) {
			sb.WriteString("\t")
		} else {
			sb.WriteString("\n")
		}

		// Just output the first column in quiet mode
		if logs.Quiet && i == 0 {
			sb.WriteString("\n")
			break
		}
	}

	if o.isHeader {
		o.isHeader = false
		// Return if we're in quiet mode and this was the header
		if logs.Quiet {
			return
		}
	}
	fmt.Fprint(o.writer, sb.String())
}

func (o *output) Flush() {
	o.writer.Flush()
}
