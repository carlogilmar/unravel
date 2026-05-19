package git

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const contextLines = 3

func computeHunks(oldContent, newContent string) []Hunk {
	dmp := diffmatchpatch.New()
	a, b, lineArray := dmp.DiffLinesToRunes(oldContent, newContent)
	diffs := dmp.DiffMainRunes(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	// Flatten into per-line operations with stable line numbering.
	type op struct {
		kind LineKind
		text string
	}
	var ops []op
	for _, d := range diffs {
		lines := splitKeepNL(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			for _, l := range lines {
				ops = append(ops, op{LineContext, l})
			}
		case diffmatchpatch.DiffInsert:
			for _, l := range lines {
				ops = append(ops, op{LineAdd, l})
			}
		case diffmatchpatch.DiffDelete:
			for _, l := range lines {
				ops = append(ops, op{LineDel, l})
			}
		}
	}

	// Locate change regions, expand by contextLines on each side, coalesce.
	type region struct{ start, end int } // inclusive indices into ops
	var regions []region
	for i := 0; i < len(ops); i++ {
		if ops[i].kind == LineContext {
			continue
		}
		j := i
		for j+1 < len(ops) && ops[j+1].kind != LineContext {
			j++
		}
		regions = append(regions, region{i, j})
		i = j
	}
	if len(regions) == 0 {
		return nil
	}

	expanded := make([]region, 0, len(regions))
	for _, r := range regions {
		s := r.start - contextLines
		if s < 0 {
			s = 0
		}
		e := r.end + contextLines
		if e >= len(ops) {
			e = len(ops) - 1
		}
		if len(expanded) > 0 && s <= expanded[len(expanded)-1].end+1 {
			expanded[len(expanded)-1].end = e
		} else {
			expanded = append(expanded, region{s, e})
		}
	}

	var hunks []Hunk
	oldLine, newLine := 1, 1
	cursor := 0
	for idx, r := range expanded {
		for cursor < r.start {
			switch ops[cursor].kind {
			case LineContext:
				oldLine++
				newLine++
			case LineAdd:
				newLine++
			case LineDel:
				oldLine++
			}
			cursor++
		}

		oldStart, newStart := oldLine, newLine
		var oldCount, newCount int
		var lines []Line
		for cursor <= r.end {
			o := ops[cursor]
			line := Line{Kind: o.kind, Content: stripNL(o.text)}
			switch o.kind {
			case LineContext:
				line.OldNum, line.NewNum = oldLine, newLine
				oldLine++
				newLine++
				oldCount++
				newCount++
			case LineAdd:
				line.NewNum = newLine
				newLine++
				newCount++
			case LineDel:
				line.OldNum = oldLine
				oldLine++
				oldCount++
			}
			lines = append(lines, line)
			cursor++
		}

		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@", oldStart, oldCount, newStart, newCount)
		hunks = append(hunks, Hunk{
			ID:       hunkID(header, idx),
			Header:   header,
			OldStart: oldStart,
			NewStart: newStart,
			Lines:    lines,
		})
	}
	return hunks
}

func splitKeepNL(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.SplitAfter(s, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func stripNL(s string) string {
	return strings.TrimRight(s, "\n")
}
