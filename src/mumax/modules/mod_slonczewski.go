//  This file is part of MuMax, a high-performance micromagnetic simulator.
//  Copyright 2011  Arne Vansteenkiste and Ben Van de Wiele.
//  Use of this source code is governed by the GNU General Public License version 3
//  (as published by the Free Software Foundation) that can be found in the license.txt file.
//  Note that you are welcome to modify this code under the condition that you do not remove any 
//  copyright notices and prominently state that you modified it, giving a relevant date.

package modules

// Module implementing Slonczewski spin transfer torque.
// Authors: Mykol Dvornik, Graham Rowlands, Arne Vansteenkiste

import (
	. "mumax/common"
	. "mumax/engine"
	"mumax/gpu"
	//"math"
)

// Register this module
func init() {
	RegisterModule("slonczewski", "Slonczewski spin transfer torque.", LoadSlonczewskiTorque)
}

func LoadSlonczewskiTorque(e *Engine) {
	e.LoadModule("llg") // needed for alpha, hfield, ...

	// ============ New Quantities =============
	e.AddNewQuant("t_fl", SCALAR, MASK, Unit(""), "Free layer thickness")

	labmda := e.AddNewQuant("lambda", SCALAR, MASK, Unit(""), "Scattering control parameter")
	labmda.SetValue([]float64{1.0})
	e.AddNewQuant("p", VECTOR, MASK, Unit(""), "Polarization Vector")

	pol := e.AddNewQuant("pol", SCALAR, MASK, Unit(""), "Polarization efficiency")
	pol.SetValue([]float64{1.0})
	epsilon_prime := e.AddNewQuant("epsilon_prime", SCALAR, MASK, Unit(""), "Field-like term")
	epsilon_prime.SetValue([]float64{0.0})
	LoadUserDefinedCurrentDensity(e)
	stt := e.AddNewQuant("stt", VECTOR, FIELD, Unit("/s"), "Slonczewski Spin Transfer Torque")

	// ============ Dependencies =============
	e.Depends("stt", "lambda", "p", "pol", "epsilon_prime", "j", "m", "gamma", "msat", "gamma", "alpha", "t_fl")

	// ============ Updating the torque =============
	stt.SetUpdater(&slonczewskiUpdater{stt: stt})

	// Add spin-torque to LLG torque
	AddTermToQuant(e.Quant("torque"), stt)
}

type slonczewskiUpdater struct {
	stt *Quant
}

func (u *slonczewskiUpdater) Update() {
	e := GetEngine()

	worldSize := e.WorldSize()

	stt := u.stt
	m := e.Quant("m")
	msat := e.Quant("msat")
	pol := e.Quant("pol")
	lambda := e.Quant("lambda")
	epsilon_prime := e.Quant("epsilon_prime")
	p := e.Quant("p")
	curr := e.Quant("j")
	alpha := e.Quant("alpha")
	gamma := e.Quant("gamma").Scalar()
	t_fl := e.Quant("t_fl")

	//njn := math.Sqrt(float64(curr.Multiplier()[0] * curr.Multiplier()[0]) + float64(curr.Multiplier()[1] * curr.Multiplier()[1]) + float64(curr.Multiplier()[2] * curr.Multiplier()[2]))

	nmsatn := msat.Multiplier()[0]

	beta := H_bar * gamma / (Mu0 * E * nmsatn)      // njn is missing
	beta_prime := pol.Multiplier()[0] * beta        // epsilon is missing, polMask is missing
	pre_fld := beta * epsilon_prime.Multiplier()[0] // epsilon_primeMsk is missing

	gpu.LLSlon(stt.Array(),
		m.Array(),
		msat.Array(),
		p.Array(),
		curr.Array(),
		alpha.Array(),
		t_fl.Array(),
		pol.Array(),
		lambda.Array(),
		epsilon_prime.Array(),
		p.Multiplier(),
		curr.Multiplier(),
		float32(beta_prime),
		float32(pre_fld),
		worldSize,
		alpha.Multiplier(),
		t_fl.Multiplier(),
		lambda.Multiplier())
}
