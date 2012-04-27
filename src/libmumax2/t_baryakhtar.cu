#include "t_baryakhtar.h"
#include "multigpu.h"
#include "gpu_conf.h"
#include "gpu_safe.h"
#include "stdio.h"
#include <cuda.h>

#ifdef __cplusplus
extern "C" {
#endif
	// mod
	int MMod(int a, int b){
		return (a%b+b)%b;
	}
	
	// dot product
	inline __host__ __device__ float dotf(float3 a, float3 b)
	{ 
		return a.x * b.x + a.y * b.y + a.z * b.z;
	}

	// cross product
	inline __host__ __device__ float3 crossf(float3 a, float3 b)
	{ 
		return make_float3(a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x); 
	}
	
  // ========================================
   
  
 __global__ void zhangli_delta2HKernMGPU(float* tx, float* ty, float* tz,
					 float* mx, float* my, float* mz,
					 float* hx, float* hy, float* hz,
					 float* lhx, float* lhy, float* lhz,
					 float* rhx, float* rhy, float* rhz,	 
					 float* msat,
					 float* Aex,
					 float pre,
					 int4 size,		
					 float3 mstep,
					 int3 pbc,
					 int i)
  {	
	
	int j = blockIdx.x * blockDim.x + threadIdx.x;
	int k = blockIdx.y * blockDim.y + threadIdx.y;			
	int x0 = i * size.w + j * size.z + k;
	
	float m_sat = (msat != NULL) ? msat[x0] : 1.0f;
	
    if (m_sat != 0.0f && j < size.y && k < size.z){ // 3D now:)
	   
	   m_sat = Aex[x0] / (m_sat*m_sat);
	  
	  
	  float3 m = make_float3(mx[x0], my[x0], mz[x0]);		
		 
      // Second-order derivative 5-points stencil
	   
	  int xb2 = i - 2;
	  int xb1 = i - 1;
	  int xf1 = i + 1;
	  int xf2 = i + 2;
	  
	  int yb2 = j - 2;
	  int yb1 = j - 1;
	  int yf1 = j + 1;
	  int yf2 = j + 2; 
	   
	  int zb2 = k - 2;
	  int zb1 = k - 1;
	  int zf1 = k + 1;
	  int zf2 = k + 2;
	  
	  int4 yi = make_int4(yb2, yb1, yf1, yf2);		  

	  xb2 = (pbc.x == 0 && xb2 < 0)? i : xb2; // backward coordinates are negative
	  xb1 = (pbc.x == 0 && xb1 < 0)? i : xb1;
	  xf1 = (pbc.x == 0 && xf1 >= size.x)? i : xf1;
	  xf2 = (pbc.x == 0 && xf2 >= size.x)? i : xf2;
	  	  
	  yb2 = (lhx == NULL && yb2 < 0)? j : yb2;
	  yb1 = (lhx == NULL && yb1 < 0)? j : yb1;
	  yf1 = (rhx == NULL && yf1 >= size.y)? j : yf1;
	  yf2 = (rhx == NULL && yf2 >= size.y)? j : yf2;
	 	  
	  zb2 = (pbc.z == 0 && zb2 < 0)? k : zb2;
	  zb1 = (pbc.z == 0 && zb1 < 0)? k : zb1;
	  zf1 = (pbc.z == 0 && zf1 >= size.z)? k : zf1;
	  zf2 = (pbc.z == 0 && zf2 >= size.z)? k : zf2;
	 	  
	  xb2 = (xb2 >= 0)? xb2 : size.x + xb2;
	  xb1 = (xb1 >= 0)? xb1 : size.x + xb1;
	  xf1 = (xf1 < size.x)? xf1 : xf1 - size.x;
	  xf2 = (xf2 < size.x)? xf2 : xf2 - size.x;
	  
	  yb2 = (yb2 >= 0)? yb2 : size.y + yb2;
	  yb1 = (yb1 >= 0)? yb1 : size.y + yb1;
	  yf1 = (yf1 < size.y)? yf1 : yf1 - size.y;
	  yf2 = (yf2 < size.y)? yf2 : yf2 - size.y;
	  
	  zb2 = (zb2 >= 0)? zb2 : size.z + zb2;
	  zb1 = (zb1 >= 0)? zb1 : size.z + zb1;
	  zf1 = (zf1 < size.z)? zf1 : zf1 - size.z;
	  zf2 = (zf2 < size.z)? zf2 : zf2 - size.z;
	  	  
	  int comm = j * size.z + k;	   
	  int4 xn = make_int4(xb2 * size.w + comm, 
						  xb1 * size.w + comm, 
						  xf1 * size.w + comm, 
						  xf2 * size.w + comm); 
						 

	  comm = i * size.w + k; 
	  int4 yn = make_int4(yb2 * size.z + comm, 
						  yb1 * size.z + comm, 
						  yf1 * size.z + comm, 
						  yf2 * size.z + comm);

	  
	  comm = i * size.w + j * size.z;
	  int4 zn = make_int4(zb2 + comm, 
						  zb1 + comm, 
						  zf1 + comm, 
						  zf2 + comm);

	  	
	  // Let's use 5-point stencil to avoid problems at the boundaries
	  // CUDA does not have vec3 operators like GLSL has, except of .xxx, 
	  // Perhaps for performance need to take into account special cases where j || to x, y or z  
	  
	  float4 HH;
  
	  HH.x = (yi.x >= 0 || lhx == NULL) ? hx[yn.x] : lhx[yn.x];
	  HH.y = (yi.y >= 0 || lhx == NULL) ? hx[yn.y] : lhx[yn.y];
	  HH.z = (yi.z < size.y || rhx == NULL) ? hx[yn.z] : rhx[yn.z];
	  HH.w = (yi.w < size.y || rhx == NULL) ? hx[yn.w] : rhx[yn.w];
	 	  	    
	  float3 dhdx2 = 	make_float3(mstep.x * (-hx[xn.x] + 16.0f * hx[xn.y] - 30.0f * hx[x0] + 16.0f * hx[xn.z] - hx[xn.w]),
									mstep.y * (-HH.x + 16.0f * HH.y - 30.0f * hx[x0] + 16.0f * HH.z - HH.w),
									mstep.z * (-hx[zn.x] + 16.0f * hx[zn.y] - 30.0f * hx[x0] + 16.0f * hx[zn.z] - hx[zn.w]));
									
      HH.x = (yi.x >= 0 || lhx == NULL) ? hy[yn.x] : lhy[yn.x];
	  HH.y = (yi.y >= 0 || lhx == NULL) ? hy[yn.y] : lhy[yn.y];
	  HH.z = (yi.z < size.y || rhx == NULL) ? hy[yn.z] : rhy[yn.z];
	  HH.w = (yi.w < size.y || rhx == NULL) ? hy[yn.w] : rhy[yn.w];
								      
	  float3 dhdy2 = 	make_float3(mstep.x * (-hy[xn.x] + 16.0f * hy[xn.y] - 30.0f * hy[x0] + 16.0f * hy[xn.z] - hy[xn.w]),
								    mstep.y * (-HH.x + 16.0f * HH.y - 30.0f * hy[x0] + 16.0f * HH.z - HH.w),
									mstep.z * (-hy[zn.x] + 16.0f * hy[zn.y] - 30.0f * hy[x0] + 16.0f * hy[zn.z] - hy[zn.w]));
									
	  HH.x = (yi.x >= 0 || lhx == NULL) ? hz[yn.x] : lhz[yn.x];
	  HH.y = (yi.y >= 0 || lhx == NULL) ? hz[yn.y] : lhz[yn.y];
	  HH.z = (yi.z < size.y || rhx == NULL) ? hz[yn.z] : rhz[yn.z];
	  HH.w = (yi.w < size.y || rhx == NULL) ? hz[yn.w] : rhz[yn.w]; 								
	  								
										
	  float3 dhdz2 = 	make_float3(mstep.x * (-hz[xn.x] + 16.0f * hz[xn.y] - 30.0f * hz[x0] + 16.0f * hz[xn.z] - hz[xn.w]),
									mstep.y * (-HH.x + 16.0f * HH.y - 30.0f * hz[x0] + 16.0f * HH.z - HH.w),
								    mstep.z * (-hz[zn.x] + 16.0f * hz[zn.y] - 30.0f * hz[x0] + 16.0f * hz[zn.z] - hz[zn.w])); 
		
	  // Don't see a point of such overkill, nevertheless:
	  
	  
	  //-------------------------------------------------//
	  		  
	  
	  float3 ddh = make_float3(dhdx2.x + dhdy2.x + dhdz2.x, dhdx2.y + dhdy2.y + dhdz2.y, dhdx2.z + dhdy2.z + dhdz2.z);
								    	  
      float3 ddhxm = crossf(ddh, m); // with minus in it
      
      float3 mxddhxm = crossf(m, ddhxm); // with minus from [ddh x m]
	  	  	  
	  tx[x0] = m_sat * (pre * mxddhxm.x);
      ty[x0] = m_sat * (pre * mxddhxm.y);
      tz[x0] = m_sat * (pre * mxddhxm.z);   
    } 
  }

  
#define BLOCKSIZE 16


  
__export__  void tbaryakhtar_async(float** tx, float** ty, float** tz, 
			 float** mx, float** my, float** mz, 
			 float** hx, float** hy, float** hz,
			 float** msat,
			 float** Aex,
			 const float pre,
			 const int sx, const int sy, const int sz,
			 const float csx, const float csy, const float csz,
			 const int pbc_x, const int pbc_y, const int pbc_z, 
			 CUstream* stream)
  {

	// 3D :)
	
	dim3 gridSize(divUp(sy, BLOCKSIZE), divUp(sz, BLOCKSIZE));
    dim3 blockSize(BLOCKSIZE, BLOCKSIZE, 1);
		
	// FUCKING THREADS PER BLOCK LIMITATION
	check3dconf(gridSize, blockSize);
		
	float i12csx = 1.0f / (12.0f * csx * csx);
	float i12csy = 1.0f / (12.0f * csy * csy);
	float i12csz = 1.0f / (12.0f * csz * csz);
	
	int syz = sy * sz;
	
		
	float3 mstep = make_float3(i12csx, i12csy, i12csz);	
	int4 size = make_int4(sx, sy, sz, syz);
	int3 pbc = make_int3(pbc_x, pbc_y, pbc_z);
	
    int nDev = nDevice();
		
	/*cudaEvent_t start,stop;
	float time;
	cudaEventCreate(&start);
	cudaEventCreate(&stop);
	cudaEventRecord(start,0);*/
		
	
	
    for (int dev = 0; dev < nDev; dev++) {
      gpu_safe(cudaSetDevice(deviceId(dev)));	 
	  		
		// calculate dev neighbours
		
		int ld = MMod(dev - 1, nDev);
		int rd = MMod(dev + 1, nDev);
				
		float* lhx = hx[ld]; 
		float* lhy = hy[ld];
		float* lhz = hz[ld];

		float* rhx = hx[rd]; 
		float* rhy = hy[rd];
		float* rhz = hz[rd];
		
		if(pbc_y == 0){             
			if(dev == 0){
				lhx = NULL;
				lhy = NULL;
				lhz = NULL;			
			}
			if(dev == nDev-1){
				rhx = NULL;
				rhy = NULL;
				rhz = NULL;
			}
		}
		
		// printf("Devices are: %d\t%d\t%d\n", ld, dev, rd);
		
		for (int i = 0; i < sx; i++) {
												   
			zhangli_delta2HKernMGPU<<<gridSize, blockSize, 0, cudaStream_t(stream[dev])>>> (tx[dev], ty[dev], tz[dev],  
												   mx[dev], my[dev], mz[dev],												   
												   hx[dev], hy[dev], hz[dev],
												   lhx, lhy, lhz,
												   rhx, rhy, rhz,
												   msat[dev],
												   Aex[dev],
												   pre,
												   size,
												   mstep,
												   pbc,
												   i);
		}

    } // end dev < nDev loop
	
	
	/*cudaEventRecord(stop,0);
	cudaEventSynchronize(stop);
	cudaEventElapsedTime(&time, start, stop);
	cudaEventDestroy(start);
	cudaEventDestroy(stop);
	printf("T-Baryakhtar kernel requires: %f ms\n",time);*/
	
  }
  
  // ========================================

#ifdef __cplusplus
}
#endif