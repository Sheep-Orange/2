//  This file is part of MuMax, a high-performance micromagnetic simulator.
//  Copyright 2011  Arne Vansteenkiste and Ben Van de Wiele.
//  Use of this source code is governed by the GNU General Public License version 3
//  (as published by the Free Software Foundation) that can be found in the license.txt file.
//  Note that you are welcome to modify this code under the condition that you do not remove any 
//  copyright notices and prominently state that you modified it, giving a relevant date.

package gpu

// Author: Arne Vansteenkiste

import (
	. "mumax/common"
	cu "cuda/driver"
	"flag"
)


const BIG = 32 * 1024 * 1024

func init() {
	flag.Parse()
	InitLogger("")
	cu.Init()
	//InitDebugGPUs()
	println("		*****  u s i n g    1    g p u  *******  ")
	InitMultiGPU([]int{0}, 0)
	SetPTXLookPath("../../../ptx")
}
