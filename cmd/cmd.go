package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
)

func FlagPlatform(v *string) {
	d, err := ioutil.TempDir("", "platform")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "platform", d, "platform directory")
}

func FlagLayers(v *string) {
	d, err := ioutil.TempDir("", "layers")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "layers", d, "layers directory")
}

func FlagBuildpack(v *string) {
	d, err := ioutil.TempDir("", "buildpack")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "buildpack", d, "buildpack directory for this buildpack")
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
	color.NoColor = false
	scanner := bufio.NewScanner(strings.NewReader(err.Error()))
	for scanner.Scan() {
		fmt.Fprintln(os.Stderr, color.New(color.FgRed).Sprintf("! %s", scanner.Text()))
	}
	fmt.Fprintln(os.Stderr, color.New(color.FgRed).Sprint("!"))
	if err, ok := err.(*ErrorFail); ok {
		os.Exit(err.Code)
	}
	os.Exit(CodeFailed)
}
