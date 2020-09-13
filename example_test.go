// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package littlebyte_test

import (
	"errors"
	"fmt"

	"github.com/magical/littlebyte"
)

func ExampleString_lengthPrefixed() {
	// This is an example of parsing length-prefixed data (as found in, for
	// example, TLS). Imagine a 16-bit prefixed series of 8-bit prefixed
	// strings.

	input := littlebyte.String([]byte{12, 0, 5, 'h', 'e', 'l', 'l', 'o', 5, 'w', 'o', 'r', 'l', 'd'})
	var result []string

	var values littlebyte.String
	if !input.ReadUint16LengthPrefixed(&values) ||
		!input.Empty() {
		panic("bad format")
	}

	for !values.Empty() {
		var value littlebyte.String
		if !values.ReadUint8LengthPrefixed(&value) {
			panic("bad format")
		}

		result = append(result, string(value))
	}

	// Output: []string{"hello", "world"}
	fmt.Printf("%#v\n", result)
}

func ExampleBuilder_lengthPrefixed() {
	// This is an example of building length-prefixed data (as found in,
	// for example, TLS). Imagine a 16-bit prefixed series of 8-bit
	// prefixed strings.
	input := []string{"hello", "world"}

	var b littlebyte.Builder
	b.AddUint16LengthPrefixed(func(b *littlebyte.Builder) {
		for _, value := range input {
			b.AddUint8LengthPrefixed(func(b *littlebyte.Builder) {
				b.AddBytes([]byte(value))
			})
		}
	})

	result, err := b.Bytes()
	if err != nil {
		panic(err)
	}

	// Output: 0c000568656c6c6f05776f726c64
	fmt.Printf("%x\n", result)
}

func ExampleBuilder_lengthPrefixOverflow() {
	// Writing more data that can be expressed by the length prefix results
	// in an error from Bytes().

	tooLarge := make([]byte, 256)

	var b littlebyte.Builder
	b.AddUint8LengthPrefixed(func(b *littlebyte.Builder) {
		b.AddBytes(tooLarge)
	})

	result, err := b.Bytes()
	fmt.Printf("len=%d err=%s\n", len(result), err)

	// Output: len=0 err=littlebyte: pending child length 256 exceeds 1-byte length prefix
}

func ExampleBuilderContinuation_errorHandling() {
	var b littlebyte.Builder
	// Continuations that panic with a BuildError will cause Bytes to
	// return the inner error.
	b.AddUint16LengthPrefixed(func(b *littlebyte.Builder) {
		b.AddUint32(0)
		panic(littlebyte.BuildError{Err: errors.New("example error")})
	})

	result, err := b.Bytes()
	fmt.Printf("len=%d err=%s\n", len(result), err)

	// Output: len=0 err=example error
}
