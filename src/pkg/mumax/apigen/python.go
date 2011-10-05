//  This file is part of MuMax, a high-performance micromagnetic simulator.
//  Copyright 2011  Arne Vansteenkiste and Ben Van de Wiele.
//  Use of this source code is governed by the GNU General Public License version 3
//  (as published by the Free Software Foundation) that can be found in the license.txt file.
//  Note that you are welcome to modify this code under the condition that you do not remove any 
//  copyright notices and prominently state that you modified it, giving a relevant date.


package apigen

import (
	"io"
	"reflect"
	"fmt"
)

type Python struct{}

func (p *Python) Filename() string {
	return "mumax2.py"
}

func (p *Python) Comment() string {
	return "#"
}

func (p *Python) WriteHeader(out io.Writer) {
	fmt.Fprintln(out, `
import os
import json

infifo = 0
outfifo = 0
initialized = 0
outputdir = ""

## Initializes the communication with mumax2.
def init():
	global infifo
	global outfifo
	global outputdir
	# get the output directory from environment
	outputdir=os.environ["MUMAX2_OUTPUTDIR"] + "/"
	# signal our intent to open the fifos
	handshake=open(outputdir + 'handshake', 'w')
	handshake.close()
	# the order in which the fifos are opened matters
	infifo=open(outputdir + 'out.fifo', 'r') # mumax's out is our in
	outfifo=open(outputdir + 'in.fifo', 'w') # mumax's in is our out
	initialized = 1

## Calls a mumax2 command and returns the result as string.
def call(command, args):
	if (initialized == 0):
		init()
	outfifo.write(json.dumps([command, args]))
	outfifo.write('\n')
	outfifo.flush()
	return json.loads(infifo.readline())
`)
}

func (p *Python) WriteFooter(out io.Writer) {

}

func (p *Python) WriteFunc(out io.Writer, name string, comment []string, argNames []string, argTypes []reflect.Type, returnTypes []reflect.Type) {
	defer func() {
		err := recover()
		if err != nil {
			//fmt.Println("WriteFunc ", name, comment, argNames, err)
		}
	}()
	fmt.Fprintln(out)
	fmt.Fprintf(out, pyDocComment(comment))
	fmt.Fprint(out, "def ", name, "(")

	args := ""
	for i := range argTypes {
		if i != 0 {
			args += ", "
		}
		args += argNames[i]
	}
	fmt.Fprintln(out, args, "):")

	fmt.Fprintf(out, `	ret = call("%s", [%s])`, name, args)
	fmt.Fprint(out, "\n	return ")
	for i := range returnTypes {
		if i != 0 {
			fmt.Fprint(out, ", ")
		}
		fmt.Fprintf(out, `%v(ret[%v])`, python_convert[returnTypes[i].String()], i)
	}
	fmt.Fprintln(out)
	//fmt.Fprintln(out, fmt.Sprintf(`	return %s(call("%s", [%s])[0])`, python_convert[retType], name, args)) // single return value only
}

var (
	// maps go types to python types	
	python_convert map[string]string = map[string]string{"int": "int",
		"float32": "float",
		"float64": "float",
		"string":  "str",
		"bool":    "bool",
		"":        ""}
)

func pyDocComment(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	str := "#"
	for _, l := range lines {
		str += "#" + l + "\n"
	}
	return str
}
