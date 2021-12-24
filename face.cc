#include <dlib/image_processing/frontal_face_detector.h>
#include <dlib/image_processing/shape_predictor.h>
#include <dlib/image_loader/jpeg_loader.h>
#include "face.h"

using namespace dlib;

static const size_t RECT_LEN = 4;
static const size_t SHAPE_LEN = 2;
static const size_t RECT_SIZE = RECT_LEN * sizeof(long);
static const size_t SHAPE_SIZE = SHAPE_LEN * sizeof(long);


class FaceRec{
private:
    std::mutex detector_mutex_;

	frontal_face_detector detector_;
	shape_predictor sp_;

public:
	FaceRec(const char* model_dir) {
		detector_ = get_frontal_face_detector();

		std::string dir = model_dir;
	    std::string shape_predictor_path = dir + "/shape_predictor_68_face_landmarks.dat";

    	deserialize(shape_predictor_path) >> sp_;
	}


	std::tuple<std::vector<rectangle>, std::vector<full_object_detection>>
	Recognize(const matrix<rgb_pixel>& img, int max_faces, int type){
	    std::vector<rectangle> rects;
	    std::vector<full_object_detection> shapes;

        std::lock_guard<std::mutex> lock(detector_mutex_);
        rects = detector_(img);

        if (rects.size() ==0 || (max_faces  > 0 && rects.size() > (size_t)max_faces ))
            return {std::move(rects), std::move(shapes)};

        std::sort(rects.begin(), rects.end());

        for (const auto& rect : rects){
            auto shape = sp_(img, rect);
            shapes.push_back(shape);
        }
        return {std::move(rects), std::move(shapes)};
	}

};


facerec* facerec_init(const char* model_dir){
	facerec* rec = (facerec*)calloc(1, sizeof(facerec));
	try {
        FaceRec* cls = new FaceRec(model_dir);
	    rec->cls = (void*) cls;
	} catch (std::exception& e) {
     	rec->err_str = strdup(e.what());
    }
	return rec;
}


faceret* facerec_recognize(facerec* rec, const uint8_t* img_data, int len, int max_faces, int type){
	faceret* ret = (faceret*) calloc(1, sizeof(faceret));
	FaceRec* cls = (FaceRec*)(rec->cls);
    matrix<rgb_pixel> img;
    std::vector<rectangle> rects;
    std::vector<full_object_detection> shapes;

    try{
        load_jpeg(img, img_data, len);
		std::tie(rects, shapes) = cls->Recognize(img, max_faces,type);
    } catch(std::exception& e){
        ret->err_str = strdup(e.what());
        return ret;
    }
    ret->num_faces = rects.size();

	if (ret->num_faces ==0)
		return ret;
	ret->rectangles = (long*)malloc(ret->num_faces*RECT_SIZE);

	for (int i = 0; i < ret->num_faces; i++) {
		long* dst = ret->rectangles + i * RECT_LEN;
		dst[0] = rects[i].left();
		dst[1] = rects[i].top();
		dst[2] = rects[i].right();
		dst[3] = rects[i].bottom();
	}

	ret->num_shapes = shapes[0].num_parts();
    ret->shapes = (long*)malloc(ret->num_faces * ret->num_shapes * SHAPE_SIZE);
    for (int i = 0; i < ret->num_faces; i++) {
        long* dst = ret->shapes + i * ret->num_shapes * SHAPE_LEN;
        const auto& shape = shapes[i];
        for (int j = 0; j < ret->num_shapes; j++) {
            dst[j*SHAPE_LEN] = shape.part(j).x();
            dst[j*SHAPE_LEN+1] = shape.part(j).y();
        }
    }
    return ret;
}


void facerec_free(facerec* rec){
	if (rec) {
		if (rec->cls){
			FaceRec* cls = (FaceRec*)(rec->cls);
			delete cls;
			rec->cls = NULL;
		}
		free(rec);
	}
}
