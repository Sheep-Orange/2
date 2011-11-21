//  This file is part of MuMax, a high-performance micromagnetic simulator.
//  Copyright 2011  Arne Vansteenkiste and Ben Van de Wiele.
//  Use of this source code is governed by the GNU General Public License version 3
//  (as published by the Free Software Foundation) that can be found in the license.txt file.
//  Note that you are welcome to modify this code under the condition that you do not remove any 
//  copyright notices and prominently state that you modified it, giving a relevant date.

package engine

// Module for exchange constant
// Author: Arne Vansteenkiste

import ()

// Register this module
func init() {
	RegisterModule(&ModAExchange{})
}

// Module for exchange constant
type ModAExchange struct{}

func (x ModAExchange) Description() string {
	return "Exchange constant [J/m]"
}

func (x ModAExchange) Name() string {
	return "aexchange"
}

func (x ModAExchange) Load(e *Engine) {
	e.AddQuant("Aex", SCALAR, MASK, Unit("J/m"), "exchange coefficient")
}