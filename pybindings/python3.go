package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/simon-engledew/gocmdpev/pev"
)

/*
#cgo pkg-config: python3
#define Py_LIMITED_API
#include <Python.h>
*/
import "C"

//export visualize
func visualize(self, args *C.PyObject) *C.PyObject {
	input, err := ArgsString(args)

	if err != nil {
		return nil
	}

	err = pev.Visualize(os.Stdout, strings.NewReader(input))

	if err != nil {
		C.PyErr_SetString(C.PyExc_RuntimeError, C.CString(fmt.Sprintf("%v", err)))
	}

	return &C._Py_NoneStruct
}
