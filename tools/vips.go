package tools

/*
#cgo pkg-config: vips
#include "vips.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/h2non/bimg"
	log "github.com/sirupsen/logrus"
)

// CreatePackshot adds the packshot effect on a picture
func CreatePackshot(image []byte) ([]byte, error) {
	jpgImage, err := ensureJPEG(image)
	if err != nil {
		return nil, err
	}

	var out *C.VipsImage

	prod, err := toVipsImage(&jpgImage)
	if err != nil {
		return nil, err
	}

	error := C.wof_generate_packshot(&out, prod)
	defer C.g_object_unref(C.gpointer(prod))
	if error == 1 {
		return nil, catchVipsError()
	}

	buf, err := toBimgImage(out, 85)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// -----------------------------------------

func ensureJPEG(image []byte) ([]byte, error) {
	format := bimg.DetermineImageType(image)

	if format == bimg.JPEG {
		// nn case it as already a JPEG
		return image, nil
	}

	bimage := bimg.NewImage(image)

	opt := bimg.Options{
		Background: bimg.Color{R: 255, G: 255, B: 255},
		Type:       bimg.JPEG,
		Quality:    100,
	}

	jpgBuf, err := bimage.Process(opt)
	if err != nil {
		log.WithError(err).Warn("Cannot convert the image to JPEG")
		return nil, err
	}

	return jpgBuf, nil
}

func toVipsImage(buf *[]byte) (*C.VipsImage, error) {
	if len(*buf) == 0 {
		return nil, errors.New("Image buffer is empty")
	}

	image, err := vipsRead(*buf)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func toBimgImage(image *C.VipsImage, quality int) ([]byte, error) {
	tmpImage, err := vipsPreSave(image)
	if err != nil {
		return nil, err
	}

	// When an image has an unsupported color space, vipsPreSave
	// returns the pointer of the image passed to it unmodified.
	// When this occurs, we must take care to not dereference the
	// original image a second time; we may otherwise erroneously
	// free the object twice.
	if tmpImage != image {
		defer C.g_object_unref(C.gpointer(tmpImage))
	}

	length := C.size_t(0)
	saveErr := C.int(0)
	interlace := C.int(0)
	qual := C.int(quality)
	strip := C.int(1)

	var ptr unsafe.Pointer
	defer C.free(ptr)
	saveErr = C.vips_jpegsave_bridge1(tmpImage, &ptr, &length, strip, qual, interlace)
	if int(saveErr) != 0 {
		return nil, catchVipsError()
	}

	buf := C.GoBytes(ptr, C.int(length))

	// Clean up
	C.g_free(C.gpointer(ptr))
	C.vips_error_clear()

	return buf, nil
}

func catchVipsError() error {
	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()
	C.vips_thread_shutdown()
	return errors.New(s)
}

func vipsRead(buf []byte) (*C.VipsImage, error) {
	var image *C.VipsImage
	imageType := bimg.DetermineImageType(buf)

	if imageType == bimg.UNKNOWN {
		return nil, errors.New("Unsupported image format")
	}

	length := C.size_t(len(buf))
	imageBuf := unsafe.Pointer(&buf[0])

	err := C.vips_init_image1(imageBuf, length, C.int(imageType), &image)
	if err != 0 {
		return nil, catchVipsError()
	}

	return image, nil
}

func vipsPreSave(image *C.VipsImage) (*C.VipsImage, error) {
	var outImage *C.VipsImage

	// Apply the proper colour space
	if vipsColourspaceIsSupported(image) {
		err := C.vips_colourspace_bridge1(image, &outImage, C.VIPS_INTERPRETATION_sRGB)
		if int(err) != 0 {
			return nil, catchVipsError()
		}
		defer C.g_object_unref(C.gpointer(image))
		image = outImage
	}

	return image, nil
}

func vipsColourspaceIsSupported(image *C.VipsImage) bool {
	return int(C.vips_colourspace_issupported_bridge1(image)) == 1
}
