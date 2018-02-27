
#include "stdint.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct ResultArr {
    const char *Err;
    void *Res;
    uint32_t Cnt;
} ResultArr;

void* NewAnalyzer();
void DestroyAnalyzer(void *ptr);
ResultArr AnalyzeFile(void *ptr, const char *path);

#ifdef __cplusplus
}
#endif