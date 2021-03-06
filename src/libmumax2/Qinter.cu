#include "Qinter.h"
#include "multigpu.h"
#include "gpu_conf.h"
#include "gpu_safe.h"
#include "stdio.h"
#include <cuda.h>
#include "common_func.h"

#ifdef __cplusplus
extern "C" {
#endif

///@internal
__global__ void QinterKern(float* __restrict__ Qi,
                           const float* __restrict__ Ti, const float* __restrict__ Tj,
                           const float* __restrict__ GijMsk,
                           const float GijMul,
                           int Npart)
{

    int i = threadindex;
    if (i < Npart)
    {
        float Tii = Ti[i];
        float Tjj = Tj[i];
        float Gij = (GijMsk == NULL) ? GijMul : GijMul * GijMsk[i];
        Qi[i] = Gij * (Tjj - Tii);
    }
}

__export__ void QinterAsync(float** Qi,
                            float** Ti, float** Tj,
                            float** Gij,
                            float GijMul,
                            int Npart,
                            CUstream* stream)
{
    dim3 gridSize, blockSize;
    make1dconf(Npart, &gridSize, &blockSize);
    for (int dev = 0; dev < nDevice(); dev++)
    {
        assert(Qi[dev] != NULL);
        assert(Ti[dev] != NULL);
        assert(Tj[dev] != NULL);
        gpu_safe(cudaSetDevice(deviceId(dev)));
        QinterKern <<< gridSize, blockSize, 0, cudaStream_t(stream[dev])>>> (Qi[dev],
                Ti[dev],
                Tj[dev],
                Gij[dev],
                GijMul,
                Npart);
    }
}

#ifdef __cplusplus
}
#endif

