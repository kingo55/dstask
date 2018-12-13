package dstask

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strings"
)

const (
	// keep it readable
	TABLE_MAX_WIDTH = 160
)

// should use a better console library after first POC

/// display list of filtered tasks with context and filter
func (ts *TaskSet) Display() {
	table := NewTable(
		"ID",
		"Priority",
		"Tags",
		"Project",
		"Summary",
	)

	for _, t := range ts.Tasks {
		table.AddRow(
			// id should be at least 2 chars wide to match column header
			// (headers can be truncated)
			fmt.Sprintf("%-2d", t.id),
			t.Priority,
			strings.Join(t.Tags," "),
			t.Project,
			t.Summary,
		)
	}

	rowsRendered := table.Render(10)

	if rowsRendered == len(ts.Tasks) {
		fmt.Printf("\n%v tasks.\n", len(ts.Tasks))
	} else {
		fmt.Printf("\n%v tasks, truncated to fit terminal.\n", len(ts.Tasks))
	}
}

// display a single task in detail, with numbered subtasks
func (t *Task) Display() {

}

type Table struct {
	Header []string
	Rows [][]string
	MaxColWidths []int
	TermWidth int
	TermHeight int
}

// header may  havetruncated words
func NewTable(header ...string) *Table {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		ExitFail("Not a TTY")
	}

	return &Table{
		Header: header,
		MaxColWidths: make([]int, len(header)),
		TermWidth: int(ws.Col),
		TermHeight: int(ws.Row),
	}
}

func (t *Table) AddRow(row ...string) {
	if len(row) != len(t.Header) {
		panic("Row is incorrect length")
	}

	for i, cell := range(row) {
		if t.MaxColWidths[i] < len(cell) {
			t.MaxColWidths[i] = len(cell)
		}
	}

	t.Rows = append(t.Rows,row)
}

// get widths appropriate to the terminal size and TABLE_MAX_WIDTH
// cells may require padding or truncation. Cell padding of 1char between
// fields recommended -- not included.
func (t *Table) calcColWidths(gap int) []int {
	target := TABLE_MAX_WIDTH

	if t.TermWidth < target {
		target = t.TermWidth
	}

	colWidths := t.MaxColWidths[:]

	// account for gaps
	target -= gap * len(colWidths) - 1

	for SumInts(colWidths...) > target {
		// find max col width index
		var max, maxi int

		for i,w := range(colWidths) {
			if w > max {
				max = w
				maxi = i
			}
		}

		// decrement, if 0 abort
		if colWidths[maxi] == 0 {
			break
		}
		colWidths[maxi] = colWidths[maxi] - 1
	}

	return colWidths
}

// render table, returning count of rows rendered
func (t *Table) Render(gap int) int {
	// TODO: ansi colours
	// TODO alternate row colours (tw)

	widths := t.calcColWidths(2)
	maxRows := t.TermHeight - gap
	rows := append([][]string{t.Header}, t.Rows...)

	for i, row := range(rows) {
		cells := row[:]
		for i, w := range(widths) {
			cells[i] = FixStr(cells[i], w)
		}
		fmt.Println(strings.Join(cells, "  "))

		if i > maxRows {
			return i
		}
	}

	return len(t.Rows)
}
