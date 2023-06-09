/*
Copyright 2023 Michael Bridgen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
)

func printTokens(s string) {
	parts, err := parseInterpolation(s)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%#v\n", parts) // for sanity checking
	for i := range parts {
		fmt.Println(parts[i].String())
	}
}

func Example_parseInterpolation_norefs() {
	printTokens("this is just a string")
	// Output:
	// "this is just a string"
}

func Example_parseInterpolation_justref() {
	printTokens("${var}")
	// Output:
	// ${var}
}

func Example_parseInterpolation_escapes() {
	printTokens("text $and $${escaped dollar")
	// Output:
	// "text $and ${escaped dollar"
}

func Example_parseInterpolation_embedref() {
	printTokens("text${var}more")
	// Output:
	// "text"
	// ${var}
	// "more"
}
