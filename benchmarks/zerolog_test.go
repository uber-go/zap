package benchmarks

import (
	"io/ioutil"
	"time"

	"github.com/rs/zerolog"
)

func newDisabledZerolog() zerolog.Logger {
	return zerolog.New(ioutil.Discard).Level(zerolog.Disabled)
}

func newZerolog() zerolog.Logger {
	return zerolog.New(ioutil.Discard)
}

func newSampledZerolog() zerolog.Logger {
	return zerolog.New(ioutil.Discard).Sample(zerolog.Often)
}

func fakeZerologFields(e zerolog.Event) zerolog.Event {
	return e.
		Int("int", 1).
		Int64("int64", 2).
		Float64("float", 3.0).
		Str("string", "four!").
		Bool("bool", true).
		Time("time", time.Unix(0, 0)).
		Err(errExample).
		Int("duration", int(time.Second)).
		Str("obj", "not supported yet").
		Str("another string", "done!")
}

func fakeZerologContext(c zerolog.Context) zerolog.Context {
	return c.
		Int("int", 1).
		Int64("int64", 2).
		Float64("float", 3.0).
		Str("string", "four!").
		Bool("bool", true).
		Time("time", time.Unix(0, 0)).
		Err(errExample).
		Str("duration", time.Second.String()).
		Str("obj", "not supported yet").
		Str("another string", "done!")
}
