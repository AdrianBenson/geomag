//nolint //cgo generates code that doesn't pass linting
package mseed

//#cgo CFLAGS: -I${SRCDIR}
//#cgo LDFLAGS: ${SRCDIR}/libmseed.a
//#include <libmseed.h>
//typedef struct buffer_s {
//	size_t	buflen;
//	void	*buffer;
//}
//buffer;
//
//static void msr_repack_handler(char *record, int reclen, void *handlerdata) {
//	buffer *b = (buffer *)(handlerdata);
//	b->buffer = (b->buflen > 0) ? realloc(b->buffer, b->buflen + (size_t) reclen) : malloc((size_t) reclen);
//	memcpy (b->buffer + b->buflen, record, reclen);
//	b->buflen += reclen;
//}
//
//static int msr_repack(MSRecord *msr, void **records, flag flush, flag verbose) {
//
//	buffer buf;
//
//	int recordcnt = 0;
//
//	buf.buflen = 0;
//	buf.buffer = NULL;
//
//	recordcnt = msr_pack (msr, msr_repack_handler, &buf, NULL, flush, verbose);
//
//	if (records != NULL && buf.buflen > 0) {
//		(*records) = malloc(buf.buflen);
//		memcpy (*records, buf.buffer, buf.buflen);
//	}
//
//	if (buf.buffer != NULL) {
//		free(buf.buffer);
//	}
//
//	return(recordcnt);
//}
import "C"

import (
	"errors"
	"unsafe"
)

func (m *MSRecord) Repack(samples []int32, flush int, verbose int) ([]byte, error) {
	buffer := C.CBytes([]byte{})

	ptr := (*[1 << 30](C.int))(unsafe.Pointer(m.datasamples))
	for i := 0; i < int(m.numsamples) && i < len(samples); i++ {
		ptr[i] = C.int(samples[i])
	}

	cErr := (int)(C.msr_repack((*C.struct_MSRecord_s)((unsafe.Pointer)(m)), &buffer, C.flag(flush), C.flag(verbose)))
	if cErr < 0 {
		return nil, errors.New("msr_repack: error")
	}

	return C.GoBytes(buffer, C.int(cErr*512)), nil
}

func (m *MSRecord) DataSamplesFloat64() ([]float64, error) {
	if m.sampletype == 'a' {
		return nil, errors.New("not a numerical formatted record")
	}
	samples := make([]float64, m.numsamples)

	switch {
	case m.sampletype == 'i':
		ptr := (*[1 << 30](C.int))(unsafe.Pointer(m.datasamples))
		for i := 0; i < int(m.numsamples); i++ {
			samples[i] = (float64)(ptr[i])
		}
	case m.sampletype == 'f':
		ptr := (*[1 << 30](C.float))(unsafe.Pointer(m.datasamples))
		for i := 0; i < int(m.numsamples); i++ {
			samples[i] = (float64)(ptr[i])
		}
	case m.sampletype == 'd':
		ptr := (*[1 << 30](C.double))(unsafe.Pointer(m.datasamples))
		for i := 0; i < int(m.numsamples); i++ {
			samples[i] = (float64)(ptr[i])
		}
	default:
		return nil, errors.New("format not coded")
	}

	return samples, nil
}

func (t *MSTrace) DataSamplesFloat64() ([]float64, error) {
	if t.sampletype == 'a' {
		return nil, errors.New("not a numerical formatted record")
	}
	samples := make([]float64, t.numsamples)

	switch {
	case t.sampletype == 'i':
		ptr := (*[1 << 30](C.int))(unsafe.Pointer(t.datasamples))
		for i := 0; i < int(t.numsamples); i++ {
			samples[i] = (float64)(ptr[i])
		}
	case t.sampletype == 'f':
		ptr := (*[1 << 30](C.float))(unsafe.Pointer(t.datasamples))
		for i := 0; i < int(t.numsamples); i++ {
			samples[i] = (float64)(ptr[i])
		}
	case t.sampletype == 'd':
		ptr := (*[1 << 30](C.double))(unsafe.Pointer(t.datasamples))
		for i := 0; i < int(t.numsamples); i++ {
			samples[i] = (float64)(ptr[i])
		}
	default:
		return nil, errors.New("format not coded")
	}

	return samples, nil
}
