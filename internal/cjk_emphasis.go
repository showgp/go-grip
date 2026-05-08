package internal

import (
	"bytes"
	"unicode/utf8"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type strongRange struct {
	start        int
	contentStart int
	contentEnd   int
	end          int
}

func promoteCJKStrongEmphasis(root gast.Node, source []byte) {
	// CommonMark requires spaces around emphasis in some CJK punctuation cases.
	// Promote those leftover literal delimiters after goldmark has parsed safer spans.
	parents := make([]gast.Node, 0)
	_ = gast.Walk(root, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if entering && node.HasChildren() {
			parents = append(parents, node)
		}
		return gast.WalkContinue, nil
	})

	for _, parent := range parents {
		promoteCJKStrongTextRuns(parent, source)
	}
}

func promoteCJKStrongTextRuns(parent gast.Node, source []byte) {
	for child := parent.FirstChild(); child != nil; {
		first, ok := child.(*gast.Text)
		if !ok || first.IsRaw() {
			child = child.NextSibling()
			continue
		}

		last := first
		for next := last.NextSibling(); isMergeableTextRunNode(last, next); next = last.NextSibling() {
			last = next.(*gast.Text)
		}

		next := last.NextSibling()
		if bytes.Contains(source[first.Segment.Start:last.Segment.Stop], []byte("**")) {
			promoteCJKStrongTextRun(parent, first, last, source)
		}
		child = next
	}
}

func isMergeableTextRunNode(previous *gast.Text, next gast.Node) bool {
	nextText, ok := next.(*gast.Text)
	return ok &&
		!previous.SoftLineBreak() &&
		!previous.HardLineBreak() &&
		!nextText.IsRaw() &&
		previous.Segment.Stop == nextText.Segment.Start
}

func promoteCJKStrongTextRun(parent gast.Node, first *gast.Text, last *gast.Text, source []byte) {
	value := source[first.Segment.Start:last.Segment.Stop]
	ranges := findCJKStrongRanges(value)
	if len(ranges) == 0 {
		return
	}

	cursor := 0
	for _, strong := range ranges {
		insertTextSegmentBefore(parent, first, first.Segment.Start+cursor, first.Segment.Start+strong.start)

		emphasis := gast.NewEmphasis(2)
		emphasis.AppendChild(emphasis, gast.NewTextSegment(text.NewSegment(
			first.Segment.Start+strong.contentStart,
			first.Segment.Start+strong.contentEnd,
		)))
		parent.InsertBefore(parent, first, emphasis)

		cursor = strong.end
	}
	trailingText := insertTextSegmentBefore(parent, first, first.Segment.Start+cursor, last.Segment.Stop)

	if last.SoftLineBreak() || last.HardLineBreak() {
		lineBreakCarrier := trailingText
		if lineBreakCarrier == nil {
			lineBreakCarrier = gast.NewTextSegment(text.NewSegment(last.Segment.Stop, last.Segment.Stop))
			parent.InsertBefore(parent, first, lineBreakCarrier)
		}
		lineBreakCarrier.SetSoftLineBreak(last.SoftLineBreak())
		lineBreakCarrier.SetHardLineBreak(last.HardLineBreak())
	}

	after := last.NextSibling()
	for child := gast.Node(first); child != nil && child != after; {
		next := child.NextSibling()
		parent.RemoveChild(parent, child)
		child = next
	}
}

func insertTextSegmentBefore(parent gast.Node, anchor gast.Node, start int, stop int) *gast.Text {
	if start >= stop {
		return nil
	}

	node := gast.NewTextSegment(text.NewSegment(start, stop))
	parent.InsertBefore(parent, anchor, node)
	return node
}

func findCJKStrongRanges(value []byte) []strongRange {
	ranges := make([]strongRange, 0)
	cursor := 0
	for cursor < len(value) {
		opener := bytes.Index(value[cursor:], []byte("**"))
		if opener < 0 {
			break
		}
		opener += cursor
		if isEscapedDelimiter(value, opener) {
			cursor = opener + 2
			continue
		}

		contentStart := opener + 2
		search := contentStart
		matched := false
		for search < len(value) {
			closer := bytes.Index(value[search:], []byte("**"))
			if closer < 0 {
				break
			}
			closer += search
			if closer == contentStart || isEscapedDelimiter(value, closer) {
				search = closer + 2
				continue
			}

			end := closer + 2
			if isCJKStrongRange(value, opener, contentStart, closer, end) {
				ranges = append(ranges, strongRange{
					start:        opener,
					contentStart: contentStart,
					contentEnd:   closer,
					end:          end,
				})
				cursor = end
				matched = true
				break
			}
			break
		}
		if !matched {
			cursor = opener + 2
		}
	}
	return ranges
}

func isCJKStrongRange(value []byte, opener int, contentStart int, contentEnd int, end int) bool {
	if contentStart >= contentEnd {
		return false
	}

	first, firstSize := utf8.DecodeRune(value[contentStart:contentEnd])
	last, lastSize := utf8.DecodeLastRune(value[contentStart:contentEnd])
	if first == utf8.RuneError && firstSize == 0 || last == utf8.RuneError && lastSize == 0 {
		return false
	}
	if util.IsSpaceRune(first) || util.IsSpaceRune(last) {
		return false
	}

	before, hasBefore := previousRune(value, opener)
	after, hasAfter := nextRune(value, end)
	openingNeedsCJKSpace := hasBefore && util.IsEastAsianWideRune(before) && util.IsPunctRune(first)
	closingNeedsCJKSpace := hasAfter && util.IsEastAsianWideRune(after) && util.IsPunctRune(last)
	return openingNeedsCJKSpace || closingNeedsCJKSpace
}

func isEscapedDelimiter(value []byte, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && value[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func previousRune(value []byte, index int) (rune, bool) {
	if index <= 0 {
		return 0, false
	}
	r, size := utf8.DecodeLastRune(value[:index])
	return r, size > 0
}

func nextRune(value []byte, index int) (rune, bool) {
	if index >= len(value) {
		return 0, false
	}
	r, size := utf8.DecodeRune(value[index:])
	return r, size > 0
}
