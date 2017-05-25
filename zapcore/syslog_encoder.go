package zapcore

import (
	"log/syslog"
	"os"
	"sync"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/internal/bufferpool"
)

// /usr/include/sys/syslog.h
const (
	severityMask   = 0x07
	facilityMask   = 0xf8
	syslogNilValue = "-"
	ceeCookie      = "@cee:"
)

var _ Encoder = &syslogEncoder{}

var _syslogPool = sync.Pool{New: func() interface{} {
	return &syslogEncoder{}
}}

type SyslogEncoderConfig struct {
	// CEECookie assists syslog parsers to know that the msg is structured data. By default, don't set the cee cookie.
	CEECookie bool
	Facility  syslog.Priority
	Hostname  string
	PID       int
	Tag       string
}

type syslogEncoder struct {
	cee      bool
	facility syslog.Priority
	hostname string // TODO make configurable
	je       *jsonEncoder
	pid      int
	tag      string // TODO make configurable
}

func getSyslogEncoder() *syslogEncoder {
	return _syslogPool.Get().(*syslogEncoder)
}

func putSyslogEncoder(enc *syslogEncoder) {
	enc.cee = false
	enc.facility = 0
	enc.hostname = ""
	enc.je = nil
	enc.pid = 0
	enc.tag = ""
	_jsonPool.Put(enc)
}

// NewSyslogEncoder creates a syslogEncoder that you should not use because I'm not done with it yet.
func NewSyslogEncoder(cfg SyslogEncoderConfig) *syslogEncoder {
	enc := getSyslogEncoder()
	enc.cee = cfg.CEECookie
	enc.facility = cfg.Facility
	enc.hostname = cfg.Hostname
	if len(enc.hostname) == 0 {
		hostname, _ := os.Hostname()
		enc.hostname = hostname
	}
	if len(enc.hostname) == 0 {
		enc.hostname = syslogNilValue
	}
	// TODO validate hostname has no whitespace
	enc.je = newJSONEncoder(
		EncoderConfig{
			TimeKey:        "",
			LevelKey:       "",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    LowercaseLevelEncoder,
			EncodeTime:     EpochTimeEncoder,
			EncodeDuration: SecondsDurationEncoder,
		},
		false,
	)
	enc.pid = cfg.PID
	if enc.pid == 0 {
		enc.pid = os.Getpid()
	}
	enc.tag = cfg.Tag
	if len(enc.tag) == 0 {
		enc.tag = os.Args[0]
	}
	return enc
}

func (enc *syslogEncoder) AddArray(key string, arr ArrayMarshaler) error {
	return enc.je.AddArray(key, arr)
}

func (enc *syslogEncoder) AddObject(key string, obj ObjectMarshaler) error {
	return enc.je.AddObject(key, obj)
}

func (enc *syslogEncoder) AddBinary(key string, val []byte) {
	enc.je.AddBinary(key, val)
}

func (enc *syslogEncoder) AddBool(key string, val bool) {
	enc.je.AddBool(key, val)
}

func (enc *syslogEncoder) AddComplex128(key string, val complex128) {
	enc.je.AddComplex128(key, val)
}

func (enc *syslogEncoder) AddDuration(key string, val time.Duration) {
	enc.je.AddDuration(key, val)
}

func (enc *syslogEncoder) AddFloat64(key string, val float64) {
	enc.je.AddFloat64(key, val)
}

func (enc *syslogEncoder) AddInt64(key string, val int64) {
	enc.je.AddInt64(key, val)
}

func (enc *syslogEncoder) AddReflected(key string, obj interface{}) error {
	return enc.je.AddReflected(key, obj)
}

func (enc *syslogEncoder) OpenNamespace(key string) {
	enc.je.OpenNamespace(key)
}

func (enc *syslogEncoder) AddString(key, val string) {
	enc.je.AddString(key, val)
}

func (enc *syslogEncoder) AddTime(key string, val time.Time) {
	enc.je.AddTime(key, val)
}

func (enc *syslogEncoder) AddUint64(key string, val uint64) {
	enc.je.AddUint64(key, val)
}

func (enc *syslogEncoder) AppendArray(arr ArrayMarshaler) error {
	return enc.je.AppendArray(arr)
}

func (enc *syslogEncoder) AppendObject(obj ObjectMarshaler) error {
	return enc.je.AppendObject(obj)
}

func (enc *syslogEncoder) AppendBool(val bool) {
	enc.je.AppendBool(val)
}

func (enc *syslogEncoder) AppendComplex128(val complex128) {
	enc.je.AppendComplex128(val)
}

func (enc *syslogEncoder) AppendDuration(val time.Duration) {
	enc.je.AppendDuration(val)
}

func (enc *syslogEncoder) AppendInt64(val int64) {
	enc.je.AppendInt64(val)
}

func (enc *syslogEncoder) AppendReflected(val interface{}) error {
	return enc.je.AppendReflected(val)
}

func (enc *syslogEncoder) AppendString(val string) {
	enc.je.AppendString(val)
}

func (enc *syslogEncoder) AppendTime(val time.Time) {
	enc.je.AppendTime(val)
}

func (enc *syslogEncoder) AppendUint64(val uint64) {
	enc.je.AppendUint64(val)
}

func (enc *syslogEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }
func (enc *syslogEncoder) AddFloat32(k string, v float32)     { enc.AddFloat64(k, float64(v)) }
func (enc *syslogEncoder) AddInt(k string, v int)             { enc.AddInt64(k, int64(v)) }
func (enc *syslogEncoder) AddInt32(k string, v int32)         { enc.AddInt64(k, int64(v)) }
func (enc *syslogEncoder) AddInt16(k string, v int16)         { enc.AddInt64(k, int64(v)) }
func (enc *syslogEncoder) AddInt8(k string, v int8)           { enc.AddInt64(k, int64(v)) }
func (enc *syslogEncoder) AddUint(k string, v uint)           { enc.AddUint64(k, uint64(v)) }
func (enc *syslogEncoder) AddUint32(k string, v uint32)       { enc.AddUint64(k, uint64(v)) }
func (enc *syslogEncoder) AddUint16(k string, v uint16)       { enc.AddUint64(k, uint64(v)) }
func (enc *syslogEncoder) AddUint8(k string, v uint8)         { enc.AddUint64(k, uint64(v)) }
func (enc *syslogEncoder) AddUintptr(k string, v uintptr)     { enc.AddUint64(k, uint64(v)) }
func (enc *syslogEncoder) AppendComplex64(v complex64)        { enc.AppendComplex128(complex128(v)) }
func (enc *syslogEncoder) AppendFloat64(v float64)            { enc.je.appendFloat(v, 64) }
func (enc *syslogEncoder) AppendFloat32(v float32)            { enc.je.appendFloat(float64(v), 32) }
func (enc *syslogEncoder) AppendInt(v int)                    { enc.AppendInt64(int64(v)) }
func (enc *syslogEncoder) AppendInt32(v int32)                { enc.AppendInt64(int64(v)) }
func (enc *syslogEncoder) AppendInt16(v int16)                { enc.AppendInt64(int64(v)) }
func (enc *syslogEncoder) AppendInt8(v int8)                  { enc.AppendInt64(int64(v)) }
func (enc *syslogEncoder) AppendUint(v uint)                  { enc.AppendUint64(uint64(v)) }
func (enc *syslogEncoder) AppendUint32(v uint32)              { enc.AppendUint64(uint64(v)) }
func (enc *syslogEncoder) AppendUint16(v uint16)              { enc.AppendUint64(uint64(v)) }
func (enc *syslogEncoder) AppendUint8(v uint8)                { enc.AppendUint64(uint64(v)) }
func (enc *syslogEncoder) AppendUintptr(v uintptr)            { enc.AppendUint64(uint64(v)) }

func (enc *syslogEncoder) Clone() Encoder {
	return enc.clone()
}

func (enc *syslogEncoder) clone() *syslogEncoder {
	clone := getSyslogEncoder()
	clone.facility = enc.facility
	clone.je = enc.je.clone()
	clone.pid = enc.pid
	clone.hostname = enc.hostname
	clone.tag = enc.tag
	return clone
}

func (enc *syslogEncoder) EncodeEntry(ent Entry, fields []Field) (*buffer.Buffer, error) {
	var p syslog.Priority
	switch ent.Level {
	case FatalLevel:
		p = syslog.LOG_EMERG
	case PanicLevel:
		p = syslog.LOG_CRIT
	case DPanicLevel:
		p = syslog.LOG_CRIT
	case ErrorLevel:
		p = syslog.LOG_ERR
	case WarnLevel:
		p = syslog.LOG_WARNING
	case InfoLevel:
		p = syslog.LOG_INFO
	case DebugLevel:
		p = syslog.LOG_DEBUG
	}
	pr := int64((enc.facility & facilityMask) | (p & severityMask))
	buf := bufferpool.Get()
	buf.AppendByte('<')
	buf.AppendInt(pr)
	buf.AppendByte('>')
	ts := ent.Time.UTC().Format(time.RFC3339Nano)
	if len(ts) == 0 {
		ts = syslogNilValue
	}
	buf.AppendString(ts)
	buf.AppendByte(' ')
	buf.AppendString(enc.hostname)
	buf.AppendByte(' ')
	buf.AppendString(enc.tag)
	if enc.pid != 0 {
		buf.AppendByte('[')
		buf.AppendInt(int64(enc.pid))
		buf.AppendByte(']')
	}
	buf.AppendByte(':')
	buf.AppendByte(' ')
	if enc.cee {
		buf.AppendString(ceeCookie)
	}
	json, err := enc.je.EncodeEntry(ent, fields)
	buf.AppendString(json.String())
	bufferpool.Put(json)
	return buf, err
}
