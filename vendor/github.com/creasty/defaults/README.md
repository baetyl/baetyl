defaults
========

[![Build Status](https://travis-ci.org/creasty/defaults.svg?branch=master)](https://travis-ci.org/creasty/defaults)
[![codecov](https://codecov.io/gh/creasty/defaults/branch/master/graph/badge.svg)](https://codecov.io/gh/creasty/defaults)
[![GitHub release](https://img.shields.io/github/release/creasty/defaults.svg)](https://github.com/creasty/defaults/releases)
[![License](https://img.shields.io/github/license/creasty/defaults.svg)](./LICENSE)

Initialize members in a struct with a default value

- Supports almost all kind of types
  - Scalar types
    - `int/8/16/32/64`, `uint/8/16/32/64`, `float32/64`
    - `uintptr`, `bool`, `string`
  - Complex types
    - `map`, `slice`, `struct`
  - Aliased types
    - `time.Duration`
    - e.g., `type Enum string`
  - Pointer types
    - e.g., `*SampleStruct`, `*int`
- Recursively initializes fields in a struct
- Dynamically sets default values by [`defaults.Setter`](./setter.go) interface
- Preserves non-initial values from being reset with a default value


Usage
-----

```go
type Gender string

type Sample struct {
	Name   string `default:"John Smith"`
	Age    int    `default:"27"`
	Gender Gender `default:"m"`

	Slice       []string       `default:"[]"`
	SliceByJSON []int          `default:"[1, 2, 3]"` // Supports JSON format
	Map         map[string]int `default:"{}"`
	MapByJSON   map[string]int `default:"{\"foo\": 123}"`

	Struct    OtherStruct  `default:"{}"`
	StructPtr *OtherStruct `default:"{\"Foo\": 123}"`

	NoTag  OtherStruct               // Recurses into a nested struct even without a tag
	OptOut OtherStruct `default:"-"` // Opt-out
}

type OtherStruct struct {
	Hello  string `default:"world"` // Tags in a nested struct also work
	Foo    string `default:"-"`
	Random int    `default:"-"`
}

// SetDefaults implements defaults.Setter interface
func (s *OtherStruct) SetDefaults() {
	if (defaults.CanUpdate(s.Random)) { // Check if the value is initial (recommended)
		s.Random = rand.Int() // Set a value dynamically
	}
}
```

```go
obj := &Sample{}
if err := defaults.Set(obj); err != nil {
	panic(err)
}
```
