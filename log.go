// package github.com/HPInc/krypton-iot-authorizer
// Author: Mahesh Unnikrishnan
// Component: Krypton AWS IoT Authorizer Lambda
// (C) HP Development Company, LP
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	iotLogger *zap.Logger
	logLevel  zap.AtomicLevel
)

const (
	defaultLogLevel = "Debug"
)

func initLogger() {
	// Log to the console by default.
	logLevel = zap.NewAtomicLevel()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout),
		logLevel)

	iotLogger = zap.New(core, zap.AddCaller())
	setLogLevel(defaultLogLevel)
}

func shutdownLogger() {
	_ = iotLogger.Sync()
}

func setLogLevel(level string) {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		// Fallback to logging at the info level.
		fmt.Printf("Falling back to the info log level. You specified: %s.\n",
			level)
		logLevel.SetLevel(zapcore.InfoLevel)
	} else {
		logLevel.SetLevel(parsedLevel)
	}
}
