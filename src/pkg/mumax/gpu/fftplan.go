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
	"cuda/cufft"
	"fmt"
)

type FFTPlan struct {
	dataSize [3]int         // Size of the (non-zero) input data block
	fftSize  [3]int         // Transform size including zero-padding. >= dataSize
	padZ     Array          // Buffer for Z-zeropadding and +2 elements for R2C
	planZ    []cufft.Handle // In-place transform of padZ parts, 1/GPU /// ... from outer space
	transp1  Array          // Buffer for partial transpose per GPU
	chunks   []Array        // A chunk (single-GPU part of these arrays) is copied from GPU to GPU
	transp2  Array          // Buffer for full YZ inter device transpose + zero padding in Z' and X
	planYX   []cufft.Handle // In-place transform of transp2 parts. Is just a Y transform for 2D.
	Stream                  //
}

func (fft *FFTPlan) Init(dataSize, fftSize []int) {
	Assert(len(dataSize) == 3)
	Assert(len(fftSize) == 3)
	NDev := NDevice()
	const nComp = 1

	// init size
	for i := range fft.dataSize {
		fft.dataSize[i] = dataSize[i]
		fft.fftSize[i] = fftSize[i]
	}

	// init stream
	fft.Stream = NewStream()

	// init padZ
	padZN0 := fft.dataSize[0]
	padZN1 := fft.dataSize[1]
	padZN2 := fft.fftSize[2] + 2
	fft.padZ.Init(nComp, []int{padZN0, padZN1, padZN2}, DO_ALLOC)

	// init planZ
	fft.planZ = make([]cufft.Handle, NDev)
	for dev := range _useDevice {
		setDevice(_useDevice[dev])
		Assert((nComp*padZN0*padZN1)%NDev == 0)
		fft.planZ[dev] = cufft.Plan1d(fft.fftSize[2], cufft.R2C, (nComp*padZN0*padZN1)/NDev)
		fft.planZ[dev].SetStream(uintptr(fft.Stream[dev])) // TODO: change
	}

	// init transp1
	fft.transp1.Init(nComp, fft.padZ.size3D, DO_ALLOC)

	// init chunks
	chunkN0 := dataSize[0]
	Assert((fftSize[2]/2)%NDev == 0)
	chunkN1 := ((fftSize[2]/2)/NDev + 1) * NDev // (complex numbers)
	Assert(dataSize[1]%NDev == 0)
	chunkN2 := (dataSize[1] / NDev) * 2 // (complex numbers)
	fft.chunks = make([]Array, NDev)
	for dev := range _useDevice {
		fft.chunks[dev].Init(nComp, []int{chunkN0, chunkN1, chunkN2}, DO_ALLOC)
	}

	// init transp2
	transp2N0 := dataSize[0] // make this fftSize[0] when copyblock can handle it
	Assert((fftSize[2]+2*NDev)%2 == 0)
	transp2N1 := (fftSize[2] + 2*NDev) / 2
	transp2N2 := fftSize[1] * 2
	fft.transp2.Init(nComp, []int{transp2N0, transp2N1, transp2N2}, DO_ALLOC)

	// init planYX
	fft.planYX = make([]cufft.Handle, NDev)
	for dev := range _useDevice {
		setDevice(_useDevice[dev])
		if fft.fftSize[0] == 1 { // 2D
			// ... fft.planYX[dev] = cufft.Plan1d(fft.fftSize[2], cufft.R2C, (nComp*padZN0*padZN1)/NDev)
		} else { //3D

		}
		fft.planYX[dev].SetStream(uintptr(fft.Stream[dev])) // TODO: change 
	}
}

func NewFFTPlan(dataSize, fftSize []int) *FFTPlan {
	fft := new(FFTPlan)
	fft.Init(dataSize, fftSize)
	return fft
}

func (fft *FFTPlan) Free() {
	for i := range fft.dataSize {
		fft.dataSize[i] = 0
		fft.fftSize[i] = 0
	}
	(&(fft.padZ)).Free()

	// TODO destroy
}

func (fft *FFTPlan) Forward(in, out *Array) {
	// shorthand
	padZ := &(fft.padZ)
	transp1 := &(fft.transp1)
	dataSize := fft.dataSize
	fftSize := fft.fftSize
	NDev := NDevice()
	chunks := fft.chunks // not sure if chunks[0] copies the struct...
	transp2 := &(fft.transp2)

	fmt.Println("in:", in.LocalCopy().Array)

	CopyPadZ(padZ, in)
	fmt.Println("padZ:", padZ.LocalCopy().Array)

	for dev := range _useDevice {
		fft.planZ[dev].ExecR2C(uintptr(padZ.pointer[dev]), uintptr(padZ.pointer[dev])) // is this really async?
	}
	fft.Sync()
	fmt.Println("fftZ:", padZ.LocalCopy().Array)

	TransposeComplexYZPart(transp1, padZ) // fftZ!
	//(&transp1).CopyFromDevice(&padZ)
	fmt.Println("transp1:", transp1.LocalCopy().Array)

	// copy chunks, cross-device
	//chunkBytes := int64(chunks[0].partLen4D) * SIZEOF_FLOAT // entire chunk  	
	chunkPlaneBytes := int64(chunks[0].partSize[1]*chunks[0].partSize[2]) * SIZEOF_FLOAT // one plane 

	Assert(dataSize[1]%NDev == 0)
	Assert(fftSize[2]%NDev == 0)
	for dev := range _useDevice { // source device
		for c := range chunks { // source chunk
			// source device = dev
			// target device = chunk

			for i := 0; i < dataSize[0]; i++ { // only memcpys in this loop
				srcPlaneN := transp1.partSize[1] * transp1.partSize[2] //fmt.Println("srcPlaneN:", srcPlaneN)//seems OK
				srcOffset := i*srcPlaneN + c*((dataSize[1]/NDev)*(fftSize[2]/NDev))
				src := cu.DevicePtr(ArrayOffset(uintptr(transp1.pointer[dev]), srcOffset))

				dstPlaneN := chunks[0].partSize[1] * chunks[0].partSize[2] //fmt.Println("dstPlaneN:", dstPlaneN)//seems OK
				dstOffset := i * dstPlaneN
				dst := cu.DevicePtr(ArrayOffset(uintptr(chunks[dev].pointer[c]), dstOffset))
				// must be done plane by plane
				cu.MemcpyDtoD(dst, src, chunkPlaneBytes) // chunkPlaneBytes for plane-by-plane
			}
		}
	}

	transp2.Zero()

	for c := range chunks {
		CopyBlockZ(transp2, &(chunks[c]), c) // no need to offset planes here.
	}

	fmt.Println("transp2:", transp2.LocalCopy().Array)
}
