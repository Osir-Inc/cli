package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Formatter struct {
	jsonMode bool
	out      io.Writer
	errOut   io.Writer
}

func New(jsonMode bool) *Formatter {
	return &Formatter{
		jsonMode: jsonMode,
		out:      os.Stdout,
		errOut:   os.Stderr,
	}
}

func (f *Formatter) SetOut(w io.Writer)    { f.out = w }
func (f *Formatter) SetErr(w io.Writer)    { f.errOut = w }
func (f *Formatter) SetJSON(json bool)     { f.jsonMode = json }
func (f *Formatter) IsJSON() bool          { return f.jsonMode }

func (f *Formatter) PrintResult(v any) {
	if f.jsonMode {
		f.printJSON(v)
	} else {
		f.printText(v)
	}
}

func (f *Formatter) PrintSuccess(msg string) {
	if f.jsonMode {
		f.printJSON(map[string]string{"status": "success", "message": msg})
	} else {
		fmt.Fprintf(f.out, "[OK] %s\n", msg)
	}
}

func (f *Formatter) PrintError(msg string) {
	if f.jsonMode {
		data, _ := json.Marshal(map[string]string{"status": "error", "message": msg})
		fmt.Fprintln(f.errOut, string(data))
	} else {
		fmt.Fprintf(f.errOut, "[ERROR] %s\n", msg)
	}
}

func (f *Formatter) PrintTable(headers []string, rows [][]string) {
	if f.jsonMode {
		f.printJSON(map[string]any{"headers": headers, "rows": rows})
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	f.printRow(headers, widths)
	sep := make([]string, len(widths))
	for i, w := range widths {
		sep[i] = strings.Repeat("-", w)
	}
	f.printRow(sep, widths)

	// Print rows
	for _, row := range rows {
		f.printRow(row, widths)
	}
}

func (f *Formatter) PrintKeyValue(key, value string) {
	if f.jsonMode {
		f.printJSON(map[string]string{key: value})
	} else {
		fmt.Fprintf(f.out, "%-20s %s\n", key+":", value)
	}
}

func (f *Formatter) Println(msg string) {
	fmt.Fprintln(f.out, msg)
}

func (f *Formatter) Printf(format string, args ...any) {
	fmt.Fprintf(f.out, format, args...)
}

func (f *Formatter) printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		f.PrintError("Failed to serialize output: " + err.Error())
		return
	}
	fmt.Fprintln(f.out, string(data))
}

func (f *Formatter) printText(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(f.out, "%v\n", v)
		return
	}
	// Pretty-print JSON as key-value for text mode
	var m map[string]any
	if json.Unmarshal(data, &m) == nil {
		for k, val := range m {
			fmt.Fprintf(f.out, "%-20s %v\n", k+":", val)
		}
	} else {
		fmt.Fprintln(f.out, string(data))
	}
}

func (f *Formatter) printRow(cells []string, widths []int) {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		w := 0
		if i < len(widths) {
			w = widths[i]
		}
		parts[i] = fmt.Sprintf("%-*s", w, cell)
	}
	fmt.Fprintln(f.out, strings.Join(parts, "  "))
}
