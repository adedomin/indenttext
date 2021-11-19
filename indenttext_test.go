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
	"reflect"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, i int, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Errorf(
		"C%d Received %v (type %v), Expected %v (type %v)",
		i, a, reflect.TypeOf(a), b, reflect.TypeOf(b),
	)
}

func last(s []string) string {
	if len(s) == 0 {
		return ""
	} else {
		return s[len(s)-1]
	}
}

type valTypeTuple struct {
	value string;
	typeof ItemType;
}

func TestBasicExample(t *testing.T) {
	// NV Pair and Conventional grammar
	testInputs := []string {
		"test: 123",
		`test:
  123
:`,
	}
	for iter, testCase := range testInputs {
		i := iter + 1
		tlen := Key
		err := Parse(
			strings.NewReader(testCase),
			func (parents []string, item string, ty ItemType) bool {
				if tlen != ty {
					t.Errorf(
						"C%d Parser Gave us a %s: %s, when we expected: %s.",
						i,
						ty,
						item,
						tlen,
					)
				}
				if ty == Key {
					assertEqual(t, i, len(parents), 0)
					assertEqual(t, i, item, "test")
				} else if ty == Value {
					// Make sure it is in the stack
					assertEqual(t, i, parents[0], "test")
					assertEqual(t, i, item, "123")
				} else if ty == Closed {
					assertEqual(t, i, len(parents), 0)
					assertEqual(t, i, item, "test")
				}
				tlen += 1
				return false
			},
		)
		if err != nil {
			t.Errorf("C%d %s", i, err)
		}
	}
}

func TestEmptyInput(t *testing.T) {
	err := Parse(
		strings.NewReader(""),
		func (p []string, i string, it ItemType) bool {
			t.Errorf(
				"Empty input should not call the visitor; got %s", i,
			)
			return false
		},
	)

	if err != nil {
		t.Error(err)
	}
}

func TestParsedEmptyInput(t *testing.T) {
	err := Parse(
		strings.NewReader(`# comment
      
           # blank lines with spaces
# etc
`),
		func (p []string, i string, it ItemType) bool {
			t.Errorf(
				"Empty input should not call the visitor; got %s", i,
			)
			return false
		},
	)

	if err != nil {
		t.Error(err)
	}
}

func ignoreInput(p []string, i string, it ItemType) bool {
	return false
}

func TestTooManyClosers(t *testing.T) {
	err := Parse(
		strings.NewReader(`this:
 is:
  a: test
 :
:
:`),
		ignoreInput,
	)

	if err == nil {
		t.Error("Should have terminated with a Parser error.")
	}
}

func TestNotEnoughClosers(t *testing.T) {
	err := Parse(
		strings.NewReader(`this:
 is:
  a: test`),
		ignoreInput,
	)

	if err == nil {
		t.Error("Should have terminated with a Parser error.")
	}
}

func TestAnonymousLists(t *testing.T) {
	cnt := 1
	err := Parse(
		strings.NewReader(`
':
  1
  ':
    2
    3
    4
  :
  5
:
':
  6
  7
  ':
    8
    9
  :
:`),
		func (p []string, i string, it ItemType) bool {
			switch it {
			case Key:
				assertEqual(t, 0, i, "")
			case Value:
				// +48 for start of ascii numerals
				assertEqual(t, cnt, i, string(rune(cnt+48)))
				cnt += 1
			}
			return false
		},
	)

	if err != nil {
		t.Error(err)
	}
}
