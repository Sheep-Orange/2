//  This file is part of MuMax, a high-performance micromagnetic simulator.
//  Copyright 2011  Arne Vansteenkiste and Ben Van de Wiele.
//  Use of this source code is governed by the GNU General Public License version 3
//  (as published by the Free Software Foundation) that can be found in the license.txt file.
//  Note that you are welcome to modify this code under the condition that you do not remove any
//  copyright notices and prominently state that you modified it, giving a relevant date.

package modules

// Simple module for thermal fluctuations according to Brown.
// Author: Arne Vansteenkiste

import (
	"cuda/curand"
	. "mumax/common"
	. "mumax/engine"
	"mumax/gpu"
)

var inBA = map[string]string{
	"Therm_seed": "Therm_seed",
	"cutoff_dt":  "cutoff_dt",
}

var depsBA = map[string]string{
	"T":       LtempName,
	"mu":      "mu",
	"msat":    "msat",
	"msat0T0": "msat0T0",
}

var outBA = map[string]string{
	"H_therm": "H_therm",
}

// Register this module
func init() {
	args := Arguments{inBA, depsBA, outBA}
	RegisterModuleArgs("temperature/brown-anisotropic", "Anisotropic thermal fluctuating field according to Brown.", args, LoadAnizBrown)
}

func LoadAnizBrown(e *Engine, args ...Arguments) {

	// make it automatic !!!
	var arg Arguments

	if len(args) == 0 {
		arg = Arguments{inBA, depsBA, outBA}
	} else {
		arg = args[0]
	}
	//
	Debug(arg)

	LoadTemp(e, arg.Deps("T")) // load temperature

	Therm_seed := e.AddNewQuant(arg.Ins("Therm_seed"), SCALAR, VALUE, Unit(""), `Random seed for H\_therm`)
	Therm_seed.SetVerifier(Int)

	Htherm := e.AddNewQuant(arg.Outs("H_therm"), VECTOR, FIELD, Unit("A/m"), "Thermal fluctuating field")
	cutoff_dt := e.AddNewQuant(arg.Ins("cutoff_dt"), SCALAR, VALUE, "s", `Update thermal field at most once per cutoff\_dt. Works best with fixed time step equal to N*cutoff\_dt.`)

	// By declaring that H_therm depends on Step,
	// It will be automatically updated at each new time step
	// and remain constant during the stages of the step.

	T := e.Quant(arg.Deps("T"))
	mu := e.Quant(arg.Deps("mu"))
	msat := e.Quant(arg.Deps("msat"))
	msat0T0 := e.Quant(arg.Deps("msat0T0"))

	e.Depends(arg.Outs("H_therm"), arg.Deps("T"), arg.Deps("mu"), arg.Deps("msat"), arg.Deps("msat0T0"), arg.Ins("Therm_seed"), arg.Ins("cutoff_dt"), "Step", "dt", "γ_LL")
	Htherm.SetUpdater(NewAnizBrownUpdater(Htherm, Therm_seed, cutoff_dt, T, mu, msat, msat0T0))

	// Add thermal field to total field
	hfield := e.Quant("H_eff")
	sum := hfield.GetUpdater().(*SumUpdater)
	sum.AddParent(arg.Outs("H_therm"))
}

// Updates the thermal field
type AnizBrownUpdater struct {
	rng              []curand.Generator // Random number generator for each GPU
	htherm           *Quant             // The quantity I will update
	therm_seed       *Quant
	mu               *Quant
	msat             *Quant
	msat0T0          *Quant
	T                *Quant
	cutoff_dt        *Quant
	therm_seed_cache int64
	last_time        float64 // time of last htherm update
}

func NewAnizBrownUpdater(htherm, therm_seed, cutoff_dt, T, mu, msat, msat0T0 *Quant) Updater {
	u := new(AnizBrownUpdater)
	u.therm_seed = therm_seed
	u.therm_seed_cache = -1e10
	u.htherm = htherm
	u.cutoff_dt = cutoff_dt
	u.mu = mu
	u.msat = msat
	u.msat0T0 = msat0T0
	u.T = T
	u.rng = make([]curand.Generator, gpu.NDevice())
	for dev := range u.rng {
		gpu.SetDeviceForIndex(dev)
		u.rng[dev] = curand.CreateGenerator(curand.PSEUDO_DEFAULT)
	}
	return u
}

// Updates H_therm
func (u *AnizBrownUpdater) Update() {
	e := GetEngine()

	therm_seed := int64(u.therm_seed.Scalar())

	if therm_seed != u.therm_seed_cache {
		for dev := range u.rng {
			seed := therm_seed + int64(dev)
			u.rng[dev].SetSeed(seed)
		}
	}

	u.therm_seed_cache = therm_seed

	// Nothing to do for zero temperature
	temp := u.T
	tempMul := temp.Multiplier()[0]
	if tempMul == 0 {
		u.htherm.Array().Zero()
		return
	}

	// Update only if we went past the dt cutoff
	t := e.Quant("t").Scalar()
	dt := e.Quant("dt").Scalar()
	cutoff_dt := u.cutoff_dt.Scalar()
	if dt < cutoff_dt {
		dt = cutoff_dt
		if u.last_time != 0 && t < u.last_time+dt {
			return
		}
	}

	// Make standard normal noise
	noise := u.htherm.Array()
	devPointers := noise.Pointers()
	N := int64(noise.PartLen4D())
	// Fills H_therm with gaussian noise.
	// CURAND does not provide an out-of-the-box way to do this in parallel over the GPUs
	for dev := range u.rng {
		gpu.SetDeviceForIndex(dev)
		u.rng[dev].GenerateNormal(uintptr(devPointers[dev]), N, 0, 1)
	}

	// Scale the noise according to local parameters
	cellSize := e.CellSize()
	V := cellSize[X] * cellSize[Y] * cellSize[Z]
	mu := u.mu

	gamma := e.Quant("γ_LL").Scalar()
	mSat := u.msat
	msat0T0 := u.msat0T0
	msatMask := mSat.Array()
	mSatMul := mSat.Multiplier()[0]
	tempMask := temp.Array()
	KB2tempMul := Kb * 2.0 * tempMul
	mu0VgammaDtMsatMul := Mu0 * V * gamma * dt * mSatMul
	KB2tempMul_mu0VgammaDtMsatMul := KB2tempMul / mu0VgammaDtMsatMul

	gpu.ScaleNoiseAniz(noise,
		mu.Array(),
		tempMask,
		msatMask,
		msat0T0.Array(),
		mu.Multiplier(),
		KB2tempMul_mu0VgammaDtMsatMul)
	noise.Stream.Sync()

	u.last_time = t
}
