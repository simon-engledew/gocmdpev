package main

import "../gopev"

/*
#cgo pkg-config: python3
#define Py_LIMITED_API
#include <Python.h>
*/
import "C"

import (
	"fmt"
	"os"
)

//export visualize
func visualize(self, args *C.PyObject) *C.PyObject {
	input, err := ArgsString(args)

	if err != nil {
		return nil
	}

	err = gopev.Visualize(os.Stdout, []byte(input))

	if err != nil {
		C.PyErr_SetString(C.PyExc_RuntimeError, C.CString(fmt.Sprintf("%v", err)))
	}

	return &C._Py_NoneStruct
}
