## structwalk [![Go Report Card](https://goreportcard.com/badge/github.com/xlab/structwalk)](https://goreportcard.com/report/github.com/xlab/structwalk) [![GoDoc](https://godoc.org/github.com/xlab/structwalk?status.svg)](https://godoc.org/github.com/xlab/structwalk)

Battle-tested Go struct and map traversal utilities. Has been around since 2016.

### FieldValue

Returns a value of a field in deeply nested structs, will traverse both maps and structs.

```go
structwalk.FieldValue(path string, in interface{}) (v interface{}, found bool)
```

Example:

```golang
var object = struct {
    Foo struct {
        Bar struct {
            Baz map[string]int
        }
    }
}{}
object.Foo.Bar.Baz = map[string]int{
    "Kek": 5,
}

value, found := structwalk.FieldValue("Foo.Bar.Baz.Kek", object)
// value = 5
// found = true
```

### GetterValue

A special case of `FieldValue`, that checks if for a field named `Foo`,
there is a method that is named `FooBytes()`. Useful to avoid allocations when accesing
string values that can be returned as a memory pointer instead (e.g. in 
[capnproto](https://github.com/glycerine/go-capnproto)).

```go
structwalk.GetterValue(path string, in interface{}) (v interface{}, found bool) 
```

### FieldList

Simply print the list of fields of a struct, recursively.

```go
structwalk.FieldList(in interface{}) []string
```

Example:

```go
var object = struct {
    Foo struct {
        Bar struct {
            Baz  map[string]int
            Baz2 string
        }
    }
}{}
object.Foo.Bar.Baz = map[string]int{
    "Kek": 5,
    "Lol": 4,
}

list := structwalk.FieldList(object)
// [
//  "Foo.Bar.Baz.Kek",
//  "Foo.Bar.Baz.Lol",
//  "Foo.Bar.Baz2"
// ]
```

### GetterList

Returns list of getter methods that accept no arguments (except the implicit pointer to the struct) and return one value. Can be used in templates to get a value that is not accessible by a method.

```go
structwalk.GetterList(in interface{}) []string
```

Example:

```go
type SomeDecorated struct{}

func (s SomeDecorated) Foo() string {
    return "foo"
}

func (s SomeDecorated) FooBytes() []byte {
    return []byte("foo")
}

func (s SomeDecorated) Bar() SomeStruct {
    return SomeStruct{}
}

list := structwalk.FieldList(SomeDecorated{})
// [
//   "Bar.Baz",
//   "Foo",
//   "FooBytes"
// ]
```

### License

MIT
