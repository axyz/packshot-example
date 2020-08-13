#include <stdlib.h>
#include <string.h>
#include <vips/vips.h>
#include <vips/foreign.h>
#include <vips/vips7compat.h>

#define INT_TO_GBOOLEAN(bool) (bool > 0 ? TRUE : FALSE)

// ------------------------------------------------------------------


static double PACKSHOT_GRAY_LEVEL = 246;
static double PACKSHOT_SCALE = 0.86;
static double PACKSHOT_WHITE_THRESHOLD = 250;
static double PACKSHOT_FEATHER_AMOUNT = 1;
static double PACKSHOT_OVERLAY[3] = {246.0, 246.0, 246.0};

static int
wof_generate_packshot(VipsImage **out, VipsImage *product) {

  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 6);
  VipsArrayDouble *overlayColor = vips_array_double_new(PACKSHOT_OVERLAY, 3);

  int width = product->Xsize;
  int height = product->Ysize;

  if (
    // resize product image in t[0]
    vips_resize( product, &t[0], PACKSHOT_SCALE, NULL ) ||

    // find pixels over PACKSHOT_WHITE_THRESHOLD and feather slightly ... JPG compression arifacts
    // will typically be in the 250-255 range
    //
    // bandand() ANDs all the tests together, ie. we find pixels where r AND g
    // AND b are all > PACKSHOT_WHITE_THRESHOLD
    vips_more_const1( t[0], &t[1],
                      PACKSHOT_WHITE_THRESHOLD,
                      NULL ) ||
    vips_bandand( t[1], &t[2], NULL ) ||
    vips_gaussblur( t[2], &t[3],
                    PACKSHOT_FEATHER_AMOUNT,
                    "precision", VIPS_PRECISION_INTEGER,
                    NULL ) ||

    // Make the new background, and blend with our product shot.
    !(t[4] = vips_image_new_from_image1( t[0], PACKSHOT_GRAY_LEVEL )) ||
    vips_ifthenelse( t[3], t[4], t[0], &t[5],
                     "blend", TRUE,
                     NULL ) ||

    // Expand outwards with the new background colour.
    vips_gravity( t[5], out,
                  VIPS_COMPASS_DIRECTION_CENTRE,
                  width,
                  height,
                  "extend", VIPS_EXTEND_BACKGROUND,
                  "background", overlayColor,
                  NULL )

  ) {
    g_object_unref(base);
    vips_area_unref((VipsArea *) overlayColor);
    return 1;
  }

  g_object_unref(base);
  vips_area_unref((VipsArea *) overlayColor);

  return 0;
}


// ------------------------------------------------------------------

enum types {
	UNKNOWN = 0,
	JPEG,
	WEBP,
	PNG,
	TIFF,
	GIF,
	PDF,
	SVG,
	MAGICK
};

int
vips_init_image1(void *buf, size_t len, int imageType, VipsImage **out) {
  int code = 1;

  	if (imageType == JPEG) {
  		code = vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  	} else if (imageType == PNG) {
  		code = vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  	} else if (imageType == WEBP) {
  		code = vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  	} else if (imageType == TIFF) {
  		code = vips_tiffload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  #if (VIPS_MAJOR_VERSION >= 8)
  #if (VIPS_MINOR_VERSION >= 3)
  	} else if (imageType == GIF) {
  		code = vips_gifload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  	} else if (imageType == PDF) {
  		code = vips_pdfload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  	} else if (imageType == SVG) {
  		code = vips_svgload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  #endif
  	} else if (imageType == MAGICK) {
  		code = vips_magickload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
  #endif
  	}

  	return code;
}

int
vips_colourspace_bridge1(VipsImage *in, VipsImage **out, VipsInterpretation space) {
	return vips_colourspace(in, out, space, NULL);
}

int
vips_colourspace_issupported_bridge1(VipsImage *in) {
	return vips_colourspace_issupported(in) ? 1 : 0;
}

int
vips_jpegsave_bridge1(VipsImage *in, void **buf, size_t *len, int strip, int quality, int
interlace) {
	return vips_jpegsave_buffer(in, buf, len,
		"strip", INT_TO_GBOOLEAN(strip),
		"Q", quality,
		"optimize_coding", TRUE,
		"interlace", INT_TO_GBOOLEAN(interlace),
		NULL
	);
}
