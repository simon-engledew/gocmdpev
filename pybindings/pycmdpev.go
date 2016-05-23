package main

/*

#define Py_LIMITED_API
#include <Python.h>

PyObject * visualize(PyObject *, PyObject *);

// Workaround missing variadic function support
// https://github.com/golang/go/issues/975
int PyArg_ParseTuple_s(PyObject * args, const char **a) {
    return PyArg_ParseTuple(args, "s", a);
}

static PyMethodDef methods[] = {
    { "visualize", visualize, METH_VARARGS, "Visualise a JSON explain" },
    { NULL, NULL, 0, NULL }
};

static struct PyModuleDef module = {
   PyModuleDef_HEAD_INIT, "pycmdpev", NULL, -1, methods
};

PyMODINIT_FUNC
PyInit_pycmdpev(void)
{
    return PyModule_Create(&module);
}

*/
import "C"

import "errors"

func ArgsString(args *C.PyObject) (string, error) {
	var a *C.char

	if C.PyArg_ParseTuple_s(args, &a) == 0 {
		return "", errors.New("ArgumentError")
	}

	return C.GoString(a), nil
}

func main() {}
