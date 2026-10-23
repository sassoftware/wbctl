// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package logutil

import (
	"log/slog"
	"os"
	"strings"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/spf13/viper"
)

// Add a lower Silent level and higher Trace level to the standard slog levels.
const (
	LevelTrace  slog.Level = -8
	LevelDebug             = slog.LevelDebug
	LevelInfo              = slog.LevelInfo
	LevelWarn              = slog.LevelWarn
	LevelError             = slog.LevelError
	LevelSilent slog.Level = 12
)

var _levelVar = slog.LevelVar{}

// Do not convert to a TextHandler without hiding the ClientSecret from
// auth.GetAccessTokenParams in the logs.
var _logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	Level: &_levelVar,
}))

// UpdateLevel updates the logging level based on the provided string. It is
// used as a callback in pflags BoolFuncP which is why it returns an error even
// though the returned value is always nil.
func UpdateLevel(level string) error {
	switch level {
	case "true":
		incrementLevel()
	case "false":
		_levelVar.Set(LevelSilent)
	default:
		setLevel(level)
	}
	viper.Set(constants.VerboseFlag, _levelVar.Level())
	// Return an error to satisfy pflag interface requirement.
	return nil
}

// GetLevel returns the current logging level.
func GetLevel() slog.Level {
	return _levelVar.Level()
}

// SetLevel sets the logging level.
func SetLevel(level any) {
	_levelVar.Set(AsLevel(level))
}

// GetLogger returns the current logger.
func GetLogger() *slog.Logger {
	return _logger
}

func incrementLevel() {
	switch _levelVar.Level() {
	case LevelDebug:
		_levelVar.Set(LevelTrace)
	case LevelInfo:
		_levelVar.Set(LevelDebug)
	case LevelWarn:
		_levelVar.Set(LevelInfo)
	case LevelError:
		_levelVar.Set(LevelInfo)
	case LevelSilent:
		_levelVar.Set(LevelInfo)
	}
}

func setLevel(level string) {
	_levelVar.Set(stringAsLevel(level))
}

func stringAsLevel(level string) slog.Level {
	level = strings.ToLower(level)
	switch level {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "silent":
		return LevelSilent
	default:
		return LevelSilent
	}
}

// AsLevel parses the given level value and returns the corresponding
// slog.Level. It supports slog.Level, int, and string types. If the input is
// not recognized, it defaults to LevelSilent.
func AsLevel(level any) slog.Level {
	switch level := level.(type) {
	case slog.Level:
		return level
	case int:
		return slog.Level(level)
	case string:
		return stringAsLevel(level)
	default:
		return LevelSilent
	}
}

// nolint:gochecknoinits
func init() {
	_levelVar.Set(LevelSilent)
}
