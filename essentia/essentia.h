
#ifdef __cplusplus
extern "C" {
#endif

void* NewAnalyzer();
void DestroyAnalyzer(void *ptr);
void AnalyzeFile(void *ptr, const char *path);

#ifdef __cplusplus
}
#endif