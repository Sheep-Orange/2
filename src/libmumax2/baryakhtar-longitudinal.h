/**
  * @file
  * This file implements perpendicular Baryakhtar's relaxation
  * See: unpublished W Wang, ..., MD, VVK, MF, HFG (2012)
  *
  * @author Mykola Dvornik
  */

#ifndef _BARYAKHTAR_LONGITUDINAL_H_
#define _BARYAKHTAR_LONGITUDINAL_H_

#include <cuda.h>
#include "cross_platform.h"


#ifdef __cplusplus
extern "C" {
#endif

DLLEXPORT  void baryakhtar_longitudinal_async(float** tx, float**  ty, float**  tz, 
			 float**  hx, float**  hy, float**  hz,
			 
			 float** msat0T0,
			 
			 float** lambda_xx,
			 float** lambda_yy,
			 float** lambda_zz,
			 float** lambda_yz,
			 float** lambda_xz,
			 float** lambda_xy,
			 
			 const float lambdaMul_xx,
			 const float lambdaMul_yy,
			 const float lambdaMul_zz,
			 const float lambdaMul_yz,
			 const float lambdaMul_xz,
			 const float lambdaMul_xy,
			 
			 CUstream* stream,
			 int Npart);
			 
#ifdef __cplusplus
}
#endif
#endif
