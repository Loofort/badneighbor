
#include "stdint.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct ResultArr {
    const char *Err;
    void *Res;
    uint32_t Cnt;
} ResultArr;

void* NewAnalyzer(int frameSize);
void DestroyAnalyzer(void *ptr);
ResultArr AnalyzeFile(void *ptr, const char *path);
float FrameEnergy(void *ptr, const float *gobuf);

#ifdef __cplusplus
}
#endif