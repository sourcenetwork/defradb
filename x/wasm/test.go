package main

import (
	"fmt"
	"reflect"
)

func main() {
	m0 := make(map[string]any)
	m1 := &m0
	m0v := reflect.ValueOf(m0)
	m1v := reflect.ValueOf(m1)
	fmt.Println(uintptr(reflect.Indirect(m1v).RawType())) // prints 333
	fmt.Println(uintptr(m0v.RawType()))                   // print 365
}
