
#include "stdint.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct ResultArr {
    const char *Err;
    void *Res;
    uint32_t Cnt;
} ResultArr;

typedef struct Resul {
    const char *Err;
    void *Signal;
    uint32_t SignalCnt;
    float Energy;
    float Loudness;
    float ReplayGain;
    float InstantPower;
    float RMS;
    int Intensity;
} Result;


void* NewAnalyzer(int frameSize);
void DestroyAnalyzer(void *ptr);
ResultArr AnalyzeFile(void *ptr, const char *path);
Result AnalyzeFrame(void *ptr, const float *gobuf);

#ifdef __cplusplus
}
#endif