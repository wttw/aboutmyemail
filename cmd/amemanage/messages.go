package main

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

func printError(format string, args ...any) {
	red := color.New(color.FgHiRed).SprintFunc()
	_, _ = fmt.Fprintf(color.Output, "%s: %s\n", red("ERROR"), fmt.Sprintf(format, args...))
}

func printWarning(format string, args ...any) {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	_, _ = fmt.Fprintf(color.Output, "%s: %s\n", yellow("WARN"), fmt.Sprintf(format, args...))
}

func printSuccess(globals *Globals, msg string, args ...any) {
	if !globals.Quiet {
		green := color.New(color.FgHiGreen).SprintFunc()
		_, _ = fmt.Fprintf(color.Output, "%s\n", green(fmt.Sprintf(msg, args...)))
	}
}

func fatal(format string, args ...any) {
	printError(format, args...)
	os.Exit(1)
}
