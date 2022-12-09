package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/cybergarage/go-cbor/cbor"
)

func main() {
	// fmt.Println("DECODE:")
	// cborObjs := []string{
	// 	// "0a",
	// 	// "1903e8",
	// 	// "3903e7",
	// 	// "fb3ff199999999999a",
	// 	// "f90001",
	// 	// "f4",
	// 	// "f5",
	// 	// "f6",
	// 	// "c074323031332d30332d32315432303a30343a30305a",
	// 	// "4449455446",
	// 	// "6449455446",
	// 	// "83010203",
	// 	"a161616141",
	// }
	// for _, cborObj := range cborObjs {
	// 	cborBytes, _ := hex.DecodeString(cborObj)
	// 	decoder := cbor.NewDecoder(bytes.NewReader(cborBytes))
	// 	goObj, _ := decoder.Decode()
	// 	m, ok := goObj.(map[string]any)
	// 	if !ok {
	// 		fmt.Printf("%v - %T\n", goObj)
	// 	}
	// 	for k, v := range m {
	// 		fmt.Printf("MAP %v:%v - %T\n", k, v, v)
	// 	}
	// 	// fmt.Printf("%v - %T\n", goObj)
	// }

	fmt.Println("\nENCODE:")

	// goTimeObj, _ := time.Parse(time.RFC3339, "2013-03-21T20:04:00Z")
	goObjs := []any{
		// uint(1000),
		// int(-1000),
		// float32(100000.0),
		// float64(-4.1),
		// false,
		// true,
		// nil,
		// goTimeObj,
		// []byte("IETF"),
		// "IETF",
		[]int{1, 2, 3},
		map[string]any{"a": "A"},
	}
	for _, goObj := range goObjs {
		var w bytes.Buffer
		encoder := cbor.NewEncoder(&w)
		encoder.Encode(goObj)
		cborBytes := w.Bytes()
		fmt.Printf("%s\n", hex.EncodeToString(cborBytes))
	}

}
