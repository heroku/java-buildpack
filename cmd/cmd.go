package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"io/ioutil"
)

const (
	DefaultGoals      = "clean install"
)

func FlagGoals(dir *string) {
	flag.StringVar(dir, "goals", DefaultGoals, "maven goals to run")
}

func FlagPlatform(dir *string) {
	d, err := ioutil.TempDir("", "platform")
	if err != nil {
		panic(err)
	}
	flag.StringVar(dir, "platform", d, "platform directory")
}

func FlagCache(dir *string) {
	d, err := ioutil.TempDir("", "cache")
	if err != nil {
		panic(err)
	}
	flag.StringVar(dir, "cache", d, "cache directory")
}

func FlagLaunch(dir *string) {
	d, err := ioutil.TempDir("", "launch")
	if err != nil {
		panic(err)
	}
	flag.StringVar(dir, "launch", d, "launch directory")
}

func FlagBuildpack(dir *string) {
	d, err := ioutil.TempDir("", "buildpack")
	if err != nil {
		panic(err)
	}
	flag.StringVar(dir, "buildpack", d, "buildpack directory for this buildpack")
}

const (
	CodeFailed      = 1
	CodeInvalidArgs = iota + 2
)

type ErrorFail struct {
	Err    error
	Code   int
	Action []string
}

func (e *ErrorFail) Error() string {
	message := "failed to " + strings.Join(e.Action, " ")
	if e.Err == nil {
		return message
	}
	return fmt.Sprintf("%s: %s", message, e.Err)
}

func FailCode(code int, action ...string) error {
	return FailErrCode(nil, code, action...)
}

func FailErr(err error, action ...string) error {
	code := CodeFailed
	if err, ok := err.(*ErrorFail); ok {
		code = err.Code
	}
	return FailErrCode(err, code, action...)
}

func FailErrCode(err error, code int, action ...string) error {
	return &ErrorFail{Err: err, Code: code, Action: action}
}

func Exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Printf("Error: %s\n", err)
	if err, ok := err.(*ErrorFail); ok {
		os.Exit(err.Code)
	}
	os.Exit(CodeFailed)
}