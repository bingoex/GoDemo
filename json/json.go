package main

import (
	"encoding/json"
	"fmt"
)

type Nest struct {
	Foo int
	Bar string
}

type TestJson struct {
	Str  string `json:"strkey"`
	Int  int
	Nest Nest `json:"nest"`
}

func StructToJson() {
	t := TestJson{
		Str: "abcdefg",
		Int: 1,
		Nest: Nest{
			Foo: 22,
			Bar: "haha",
		},
	}

	s, _ := json.MarshalIndent(&t, "", "\t")
	fmt.Printf("t to json :\n%s\n\n", string(s))
}

func main() {
	StructToJson()

	// json str to struct
	type Foo struct {
		K   int `json:"k"`
		kkk string
	}
	var f Foo
	json.Unmarshal([]byte(`{"k": 1, "kkk": "fdsafdsafdsf"}`), &f)
	fmt.Printf("f = '%+v'\n", f)

	bar := Foo{K:5, kkk:"fdsafsafasf"}
	fmt.Printf("bar = '%v'\n", bar)
	fmt.Printf("bar = '%+v'\n", bar)
}
