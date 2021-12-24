package face23

// #cgo CXXFLAGS: -std=c++1z -Wall -O3 -DNDEBUG -march=native
// #cgo LDFLAGS: -ldlib -ljpeg -lpng /lib/libz.so  /usr/lib/libopenblas.so
// #include <stdlib.h>
// #include <stdint.h>
// #include <face.h>
import "C"

import (
	"errors"
	"image"
	"io/ioutil"
	"os"
	"unsafe"
)

const (
	rectLen  = 4
	shapeLen = 2

	// maxElements = maxFaceLimit * (rectLen + 68 * shapeLen)
	maxElements  = 7000
	maxFaceLimit = 50
)

type Recognizer struct {
	ptr *C.facerec
}

type Face struct {
	Rectangle image.Rectangle
	Shapes    []image.Point
}

func NewRecognizer(modelDir string) (rec *Recognizer, err error) {
	cModelDir := C.CString(modelDir)
	defer C.free(unsafe.Pointer(cModelDir))
	ptr := C.facerec_init(cModelDir)

	if ptr.err_str != nil {
		defer C.facerec_free(ptr)
		defer C.free(unsafe.Pointer(ptr.err_str))
		err = errors.New(C.GoString(ptr.err_str))
		return
	}
	rec = &Recognizer{ptr}
	return
}

// type never use
func (rec *Recognizer) recognize(type_ int, imgData []byte, maxFaces int) (faces []Face, err error) {
	if len(imgData) == 0 {
		err = errors.New("empty image")
		return
	}
	if maxFaces > maxFaceLimit {
		maxFaces = maxFaceLimit
	}
	cImgData := (*C.uint8_t)(&imgData[0])
	cLen := C.int(len(imgData))
	cMaxFaces := C.int(maxFaces)
	cType := C.int(type_)

	ret := C.facerec_recognize(rec.ptr, cImgData, cLen, cMaxFaces, cType)
	defer C.free(unsafe.Pointer(ret))

	if ret.err_str != nil {
		defer C.free(unsafe.Pointer(ret.err_str))
		err = errors.New(C.GoString(ret.err_str))
		return
	}

	numFaces := int(ret.num_faces)
	if numFaces == 0 {
		return
	}
	numShapes := int(ret.num_shapes)

	// Copy faces data to Go structure.
	defer C.free(unsafe.Pointer(ret.shapes))
	defer C.free(unsafe.Pointer(ret.rectangles))

	rDataLen := numFaces * rectLen
	rDataPtr := unsafe.Pointer(ret.rectangles)
	rData := (*[maxElements]C.long)(rDataPtr)[:rDataLen:rDataLen]

	sDataLen := numFaces * numShapes * shapeLen
	sDataPtr := unsafe.Pointer(ret.shapes)
	sData := (*[maxElements]C.long)(sDataPtr)[:sDataLen:sDataLen]

	for i := 0; i < numFaces; i++ {
		face := Face{}
		x0 := int(rData[i*rectLen])
		y0 := int(rData[i*rectLen+1])
		x1 := int(rData[i*rectLen+2])
		y1 := int(rData[i*rectLen+3])
		face.Rectangle = image.Rect(x0, y0, x1, y1)
		for j := 0; j < numShapes; j++ {
			shapeX := int(sData[(i*numShapes+j)*shapeLen])
			shapeY := int(sData[(i*numShapes+j)*shapeLen+1])
			face.Shapes = append(face.Shapes, image.Point{X: shapeX, Y: shapeY})
		}
		faces = append(faces, face)
	}
	return
}

func (rec *Recognizer) RecognizeFile(imgPath string) (faces []Face, err error) {
	fd, err := os.Open(imgPath)
	if err != nil {
		return
	}
	defer fd.Close()
	imgData, err := ioutil.ReadAll(fd)
	if err != nil {
		return
	}
	return rec.recognize(0, imgData, 0)
}

func (rec *Recognizer) Recognize(imgData []byte) (faces []Face, err error) {
	return rec.recognize(0, imgData, 0)
}

func (rec *Recognizer) Close() {
	C.facerec_free(rec.ptr)
	rec.ptr = nil
}
