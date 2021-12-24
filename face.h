#pragma once

#ifdef __cplusplus
extern "C" {
#endif


typedef struct facerec {
	void* cls;
	const char* err_str;
} facerec;

typedef struct faceret {
	int num_faces;
	long* rectangles;
	int num_shapes;
	long* shapes;
	const char* err_str;
} faceret;

facerec* facerec_init(const char* model_dir);
faceret* facerec_recognize(facerec* rec, const uint8_t* img_data, int len, int max_faces,int type);
void facerec_free(facerec* rec);

#ifdef __cplusplus
}
#endif
