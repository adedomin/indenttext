# IndentText

Simple Configuration Format that tries to be easy to use and understand at a glance.
Unlike other formats, IndentText does not have any types other than ascii-like text data.

### Example

```
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
```

### Usage

```
// Where file contains:
// something: something else
myConfigFile, err := os.Open("file")
if err != nil {
    panic(err)
}

err := indenttext.Parse(
    myConfigFile,
    func (parents []string, item string, typeof ItemType) bool {
        fmt.Printf("%s - %s\n", item, typeof)
        return false // return true to cancel parsing
    },
)
if err != nil {
    panic(err)
}
```


### Todo

  1. More comprehensive tests.
  2. Better documentation
  3. Serialization functions

### Inspired by

[Deco - Delimiter Collision Free Format](https://github.com/Enhex/Deco)
