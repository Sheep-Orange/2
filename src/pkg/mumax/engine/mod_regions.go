//  This file is part of MuMax, a high-performance micromagnetic simulator.
//  Copyright 2011  Arne Vansteenkiste and Ben Van de Wiele.
//  Use of this source code is governed by the GNU General Public License version 3
//  (as published by the Free Software Foundation) that can be found in the license.txt file.
//  Note that you are welcome to modify this code under the condition that you do not remove any 
//  copyright notices and prominently state that you modified it, giving a relevant date.

package engine

// Author: Rémy Lassalle-Balier

import ()

// Register this module
func init() {
	RegisterModule(&ModRegions{})
}

// Magnetization module.
type ModRegions struct{}

func (x ModRegions) Description() string {
	return "regionDefinition: regions"
}

func (x ModRegions) Name() string {
	return "regions"
}

func (x ModRegions) Load(e *Engine) {

	e.AddQuant("regionDefinition", SCALAR, MASK, Unit(""), "regions")

	//Regions := e.Quant("regionDefinition")
	//m.updater = &normUpdater{m: m, Msat: Msat}
}