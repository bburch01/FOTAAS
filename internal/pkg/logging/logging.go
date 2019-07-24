package logging

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogMode int

const (
	Development LogMode = iota
	Production
)

func (lm LogMode) String() string {
	return [...]string{"Development", "Production"}[lm]
}

func LogModeForString(s string) (LogMode, error) {
	var lm LogMode
	var err error
	switch s {
	case Development.String():
		lm = Development
	case Production.String():
		lm = Production
	default:
		err = errors.New(fmt.Sprintf("invalid log mode: %v", s))
	}
	return lm, err
}

func NewLogger(logmode LogMode, logdir string, logfileName string) (*zap.Logger, error) {

	var cfg zap.Config
	logfilePathname := logdir + "/" + logfileName
	switch logmode {
	case Development:
		{
			cfg = zap.Config{
				Encoding:         "console",
				Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
				OutputPaths:      []string{"stderr", logfilePathname},
				ErrorOutputPaths: []string{"stderr"},
				EncoderConfig: zapcore.EncoderConfig{
					MessageKey: "message",

					LevelKey:    "level",
					EncodeLevel: zapcore.CapitalLevelEncoder,

					TimeKey:    "time",
					EncodeTime: zapcore.ISO8601TimeEncoder,

					CallerKey:    "caller",
					EncodeCaller: zapcore.ShortCallerEncoder,

					StacktraceKey: "stack",
				},
			}

		}
	case Production:
		{
			cfg = zap.Config{
				Encoding:         "json",
				Level:            zap.NewAtomicLevelAt(zapcore.WarnLevel),
				OutputPaths:      []string{"stderr", logfilePathname},
				ErrorOutputPaths: []string{"stderr"},
				EncoderConfig: zapcore.EncoderConfig{
					MessageKey: "message",

					LevelKey:    "level",
					EncodeLevel: zapcore.CapitalLevelEncoder,

					TimeKey:    "time",
					EncodeTime: zapcore.ISO8601TimeEncoder,

					CallerKey:    "caller",
					EncodeCaller: zapcore.ShortCallerEncoder,

					StacktraceKey: "stack",
				},
			}
		}
	}

	return cfg.Build()
}
