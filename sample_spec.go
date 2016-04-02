package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

const (
	DEFAULT_SAMPLE_RATE  = 44100
	DEFAULT_NUM_CHANNELS = 2
)

type SampleFormat int

const (
	FormatInvalid        SampleFormat = C.PA_SAMPLE_INVALID   // An invalid value.
	FormatPcmU8                       = C.PA_SAMPLE_U8        // Unsigned 8 Bit PCM.
	FormatALaw8                       = C.PA_SAMPLE_ALAW      // 8 Bit a-Law
	FormatMuLaw8                      = C.PA_SAMPLE_ULAW      // 8 Bit mu-Law
	FormatPcmS16LE                    = C.PA_SAMPLE_S16LE     // Signed 16 Bit PCM, little endian (PC)
	FormatPcmS16BE                    = C.PA_SAMPLE_S16BE     // Signed 16 Bit PCM, big endian.
	FormatIEEEFloat32LE               = C.PA_SAMPLE_FLOAT32LE // 32 Bit IEEE floating point, little endian (PC), range -1.0 to 1.0
	FormatIEEEFloat32BE               = C.PA_SAMPLE_FLOAT32BE // 32 Bit IEEE floating point, big endian, range -1.0 to 1.0
	FormatPcmS32LE                    = C.PA_SAMPLE_S32LE     // Signed 32 Bit PCM, little endian (PC)
	FormatPcmS32BE                    = C.PA_SAMPLE_S32BE     // Signed 32 Bit PCM, big endian.
	FormatPcmS24PackedLE              = C.PA_SAMPLE_S24LE     // Signed 24 Bit PCM packed, little endian (PC).
	FormatPcmS24PackedBE              = C.PA_SAMPLE_S24BE     // Signed 24 Bit PCM packed, big endian. (Since 0.9.15)
	FormatPcmS24Lsb32LE               = C.PA_SAMPLE_S24_32LE  // Signed 24 Bit PCM in LSB of 32 Bit words, little endian (PC). (Since 0.9.15)
	FormatPcmS24Lsb32BE               = C.PA_SAMPLE_S24_32BE  // Signed 24 Bit PCM in LSB of 32 Bit words, big endian. (Since 0.9.15)
	FormatMax                         = C.PA_SAMPLE_MAX       // Upper limit of valid sample types. (Since 0.9.15)
)

// A SampleSpec describes an audio sampling format
//
type SampleSpec struct {
	Format      SampleFormat
	SampleRate  uint32
	NumChannels int
}

func (self *SampleSpec) toNative() *C.pa_sample_spec {
	rv := C.pulse_new_sample_spec((C.pa_sample_format_t)(self.Format), C.uint32_t(self.SampleRate), C.uint8_t(self.NumChannels))
	return &rv
}

func DefaultSampleSpec() SampleSpec {
	return SampleSpec{
		Format:      FormatPcmS16LE,
		SampleRate:  DEFAULT_SAMPLE_RATE,
		NumChannels: DEFAULT_NUM_CHANNELS,
	}
}

func GetSampleFormat(cvalue int) SampleFormat {
	switch cvalue {
	case int(C.PA_SAMPLE_U8):
		return FormatPcmU8

	case int(C.PA_SAMPLE_ALAW):
		return FormatALaw8

	case int(C.PA_SAMPLE_ULAW):
		return FormatMuLaw8

	case int(C.PA_SAMPLE_S16LE):
		return FormatPcmS16LE

	case int(C.PA_SAMPLE_S16BE):
		return FormatPcmS16BE

	case int(C.PA_SAMPLE_FLOAT32LE):
		return FormatIEEEFloat32LE

	case int(C.PA_SAMPLE_FLOAT32BE):
		return FormatIEEEFloat32BE

	case int(C.PA_SAMPLE_S32LE):
		return FormatPcmS32LE

	case int(C.PA_SAMPLE_S32BE):
		return FormatPcmS32BE

	case int(C.PA_SAMPLE_S24LE):
		return FormatPcmS24PackedLE

	case int(C.PA_SAMPLE_S24BE):
		return FormatPcmS24PackedBE

	case int(C.PA_SAMPLE_S24_32LE):
		return FormatPcmS24Lsb32LE

	case int(C.PA_SAMPLE_S24_32BE):
		return FormatPcmS24Lsb32BE

	case int(C.PA_SAMPLE_MAX):
		return FormatMax

	default:
		return FormatInvalid
	}
}
