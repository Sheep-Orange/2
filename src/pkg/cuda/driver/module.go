// Copyright 2011 Arne Vansteenkiste (barnex@gmail.com).  All rights reserved.
// Use of this source code is governed by a freeBSD
// license that can be found in the LICENSE.txt file.

package driver


// This file implements loading of CUDA ptx modules

//#include <cuda.h>
import "C"

import (
	"unsafe"
)


// Represents a CUDA CUmodule, a reference to executable device code.
type Module uintptr


// Loads a compute module from file
func ModuleLoad(fname string) Module {
	//fmt.Fprintln(os.Stderr, "driver.ModuleLoad", fname)
	var mod C.CUmodule
	err := Result(C.cuModuleLoad(&mod, C.CString(fname)))
	if err != SUCCESS {
		panic(err)
	}
	return Module(unsafe.Pointer(mod))
}


// Loads a compute module from string
func ModuleLoadData(image string) Module {
	var mod C.CUmodule
	err := Result(C.cuModuleLoadData(&mod, unsafe.Pointer(C.CString(image))))
	if err != SUCCESS {
		panic(err)
	}
	return Module(unsafe.Pointer(mod))
}


// Returns a Function handle
func ModuleGetFunction(module Module, name string) Function {
	var function C.CUfunction
	err := Result(C.cuModuleGetFunction(&function, C.CUmodule(unsafe.Pointer(module)), C.CString(name)))
	if err != SUCCESS {
		panic(err)
	}
	return Function(unsafe.Pointer(function))
}


func (m Module) GetFunction(name string) Function {
	return ModuleGetFunction(m, name)
}
