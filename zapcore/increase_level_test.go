// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zapcore_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestIncreaseLevel(t *testing.T) {
	tests := []struct {
		coreLevel     Level
		increaseLevel Level
		wantErr       bool
	}{
		{
			coreLevel:     InfoLevel,
			increaseLevel: DebugLevel,
			wantErr:       true,
		},
		{
			coreLevel:     InfoLevel,
			increaseLevel: InfoLevel,
		},
		{
			coreLevel:     InfoLevel,
			increaseLevel: ErrorLevel,
		},
		{
			coreLevel:     ErrorLevel,
			increaseLevel: DebugLevel,
			wantErr:       true,
		},
		{
			coreLevel:     ErrorLevel,
			increaseLevel: InfoLevel,
			wantErr:       true,
		},
		{
			coreLevel:     ErrorLevel,
			increaseLevel: WarnLevel,
			wantErr:       true,
		},
		{
			coreLevel:     ErrorLevel,
			increaseLevel: PanicLevel,
		},
	}

	for _, tt := range tests {
		msg := fmt.Sprintf("increase %v to %v", tt.coreLevel, tt.increaseLevel)
		t.Run(msg, func(t *testing.T) {
			logger, _ := observer.New(tt.coreLevel)

			filteredLogger, err := NewIncreaseLevelCore(logger, tt.increaseLevel)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid increase level")
				return
			}

			require.NoError(t, err)

			for l := DebugLevel; l <= FatalLevel; l++ {
				enabled := filteredLogger.Enabled(l)
				ce := filteredLogger.Check(Entry{Level: l}, nil)

				if l >= tt.increaseLevel {
					assert.True(t, enabled, "expect %v to be enabled", l)
					assert.NotNil(t, ce, "expect non-nil Check")
				} else {
					assert.False(t, enabled, "expect %v to be disabled", l)
					assert.Nil(t, ce, "expect nil Check")
				}
			}
		})
	}
}
