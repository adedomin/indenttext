/*
 * Copyright (c) 2021 Anthony DeDominic <adedomin@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package indenttext

import (
	"bufio"
	"fmt"
	"io"
)

// Error with a message, line and column context
type ConfigError struct {
	context string
	line int
	col int
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf(
		"line:%d col:%d Error: %s",
		e.line, e.col, e.context,
	)
}

type ItemType int

const (
	Key ItemType = iota
	Value ItemType = iota
	Closed ItemType = iota
)

func (i ItemType) String() string {
	switch i {
	case Key: return "Key"
	case Value: return "Value"
	case Closed: return "Closed"
	default: return "INVALID"
	}
}

type Visitor func (parents []string, item string, typeof ItemType) bool

// Parser that stores the context, state and a Reader of the file being parsed
type parser struct {
	content *bufio.Reader
	line []byte
	lineno int
	lineLenLimit int
}

func (p *parser) nextLine() error {
	var prefix bool
	var err error
	p.line, prefix, err = p.content.ReadLine()
	if err != nil {
		return err
	}

	lim := len(p.line) // intentional byte length function
	if lim > p.lineLenLimit {
		return &ConfigError{
			context: "Line is too long",
			line:    p.lineno,
			col:     lim,
		}
	}

	if prefix {
		for prefix {
			var tline []byte
			tline, prefix, err = p.content.ReadLine()
			if err != nil && err != io.EOF {
				return err
			}

			lim += len(tline)
			if lim > p.lineLenLimit {
				return &ConfigError{
					context: "Line is too long",
					line:    p.lineno,
					col:     lim,
				}
			}
			p.line = append(p.line, tline...)
		}
	}

	p.lineno++

	return nil
}

func (p *parser) buildConfError(context string, col int) error {
	return &ConfigError{
		context: context,
		line:    p.lineno,
		col:     col,
	}
}

func (p *parser) iterParse(fn Visitor) error {
	var err error
	var stack []string
	
	for err = p.nextLine(); err == nil; err = p.nextLine() {
		i := 0
		// find start of content
		for ; i < len(p.line); i++ {
			if p.line[i] != ' ' && p.line[i] != '\t' {
				break
			}
		}
		// skip empty p.lines
		if i == len(p.line) {
			continue
		}

		start := i
		// start content marker
		// escapes leading whitespace and closing colon, e.g. ':
		foundContentStart := false
		if p.line[i] == '\'' {
			i++ // skip content_marker
			start = i
			foundContentStart = true
		// skip comments
		} else if p.line[i] == '#' {
			continue
		}

		// find end
		end := len(p.line)
		endToken := p.line[end-1]
		switch (endToken) {
		case ':':
			end = end - 1
		case '\'':
			if start != end {
				end = end -1
			}
		}
		
		// find name: value pair start (": ")
		nvPairNameEnd := -1
		foundNVPairMaybe := false
		// For the sake of purity, if a line begins with a
		// content start delimiter, it cannot be an Name-Value pair 
		if !foundContentStart {
			for ; i < end; i++ {
				if p.line[i] == ':' {
					foundNVPairMaybe = true
				} else if foundNVPairMaybe && p.line[i] == ' ' {
					nvPairNameEnd = i - 1
					break
				} else if foundNVPairMaybe {
					foundNVPairMaybe = false
				}
			}
		}

		switch (endToken) {
		case ':':
			if start == end && !foundContentStart { // Close
				if (len(stack) == 0) {
					return p.buildConfError("Too many compound terminators ':'", start)
				}

				poppedKey := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if fn(stack, poppedKey, Closed) {
					return p.buildConfError("Canceled", start)
				}
			} else { // Key
				newKey := string(p.line[start:end])
				if fn(stack, newKey, Key) {
					return p.buildConfError("Canceled", start)
				}
				stack = append(stack, newKey)
			}
		default:
			if nvPairNameEnd == -1 { // Value
				newVal := string(p.line[start:end])
				if fn(stack, newVal, Value) {
					return p.buildConfError("Canceled", start)
				}
			} else { // Name-Value Pair
				newKey := string(p.line[start:nvPairNameEnd])
				if fn(stack, newKey, Key) {
					return p.buildConfError("Canceled", start)
				}

				stack = append(stack, newKey)
				newVal := string(p.line[nvPairNameEnd+2:end])
				if fn(stack, newVal, Value) {
					return p.buildConfError("Canceled", start)
				}
				stack = stack[:len(stack)-1]

				if fn(stack, newKey, Closed) {
					return p.buildConfError("Canceled", start)
				}
			} 
		}
	}

	if err == io.EOF && len(stack) > 1 {
		return &ConfigError{
			context: "Unterminated compound group, not enough ':'",
			line: p.lineno,
			col: 0,
		}
	} else if err != io.EOF {
		return err
	} else {
		return nil
	}
}

// Parse a given readable file using the given visitor.
// The visitor will be visited with the current heirarchy of the tree
// and the value/type of the given item parsed.
// Visitor can cancel parsing by returning true.
func Parse(content io.Reader, visitor Visitor) error {
	h := &parser{
		content:      bufio.NewReader(content),
		line:         nil,
		lineno:       0,
		lineLenLimit: 256 * 1024, // 256KiB seems pretty reasonable sanity limit.
	}

	return h.iterParse(visitor)
}
