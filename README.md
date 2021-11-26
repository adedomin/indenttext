# IndentText

Simple Configuration Format that tries to be easy to use and understand at a glance.
Unlike other formats, IndentText does not have any types other than ascii-like text data.

### Example

```go
import (
	"github.com/adedomin/indenttext"
	"strings"
	"fmt"
)

func main() {
	myConfigFile := strings.NewReader(`
Key: Value
List of Values:
  1
  2
  3
  4
:
# this is a comment, empty lines are also ignored
# unless they are escaped with a content begin marker (')
'
'# this is not a comment

':
  a list with no name-value
:

':
  nested:
    list:
      type
    :
  :
:

# to escape a list start delimiter (:)
# use a content end marker (')
this is not a list of values:'
`)
	err := indenttext.Parse(
		myConfigFile,
		func (parents []string, item string, typeof indenttext.ItemType) bool {
			fmt.Printf("%s, %s - %s\n", parents, item, typeof)
			return false // return true to cancel parsing
		},
	)
	if err != nil {
		panic(err)
	}
}
```

### Delimiters

```
\n             - newline; record delimiter
'[rest]        - content-start-marker; escapes: leading spaces, comments, name-value, single quote at start
[rest]'        - content-end-marker; escapes: single quote at end of value and start-of-list
[content]:     - start-of-list of records, with key as content.
':             - anonymous-list
:              - end-of-list
#[rest]        - comment
[key]: [value] - name-value-pair.
```
### Notable escapes

```
# to escape name-value pairs, start the pair with a content start marker
'not name: value pair
name: value pair

# to escape a list end delimiter, use a content end marker, you can use both a start and end for look
:'
# or
':'

# NOT
':
:
# ': is a "anonymous list" where the key of this list is the empty string.
```

You cannot escape content with newlines.
The best alternative that exists is a list of values joined with newlines on parsing.

e.g.

```
my newline riddled content:
  line 1
  line 2
  more content
  blah
  trailing newline
  '
:
```
Where the empty value, `'` at the end, can be used to add trailing newlines.

### Todo

  1. More comprehensive tests.
  2. Better documentation
  3. Serialization functions

### Inspired by

[Deco - Delimiter Collision Free Format](https://github.com/Enhex/Deco)
package main

