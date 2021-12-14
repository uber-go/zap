package zapcore

import (
	"encoding/json"
	"io"
)

// A ReflectedEncoder handles the serialization of Field which have FieldType = ReflectType and writes them to the underlying data stream
// The underlying data stream is provided by zap during initialization. See EncoderConfig.NewReflectedEncoder
type ReflectedEncoder interface {
	// Encode encodes and writes to the underlying data stream
	Encode(interface{}) error
}

func GetDefaultReflectedEncoder() func(writer io.Writer) ReflectedEncoder {
	return func(writer io.Writer) ReflectedEncoder {
		stdJsonEncoder := json.NewEncoder(writer)
		// For consistency with our custom JSON encoder.
		stdJsonEncoder.SetEscapeHTML(false)
		return &stdReflectedEncoder{
			encoder: stdJsonEncoder,
		}
	}
}

type stdReflectedEncoder struct {
	encoder *json.Encoder
}

func (enc *stdReflectedEncoder) Encode(obj interface{}) error {
	return enc.encoder.Encode(obj)
}
