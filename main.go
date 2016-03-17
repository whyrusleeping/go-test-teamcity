package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"
)

const (
	TEAMCITY_TIMESTAMP_FORMAT = "2006-01-02T15:04:05.000"
)

var (
	input  = os.Stdin
	output = os.Stdout

	additionalTestName = ""

	run  = regexp.MustCompile("=== RUN\\s+(\\w+)")
	pass = regexp.MustCompile("--- PASS:\\s+(\\w+) \\(([\\.\\d]+)s\\)")
	skip = regexp.MustCompile("--- SKIP:\\s+(\\w+)\\s+\\(([\\.\\d]+)s\\)")
	fail = regexp.MustCompile("--- FAIL:\\s+(\\w+)\\s+\\(([\\.\\d]+)s\\)")
)

func init() {
	flag.StringVar(&additionalTestName, "name", "", "Add prefix to test name")
}

func main() {
	flag.Parse()

	if len(additionalTestName) > 0 {
		additionalTestName += " "
	}

	scan := bufio.NewScanner(input)

	tests := make(map[string]time.Time)

	for scan.Scan() {
		line := scan.Text()

		tnow := time.Now()
		now := tnow.Format(TEAMCITY_TIMESTAMP_FORMAT)

		runOut := run.FindStringSubmatch(line)
		if runOut != nil {
			tests[additionalTestName+runOut[1]] = time.Now()
			fmt.Fprintf(output, "##teamcity[testStarted name='%s']\n", additionalTestName+runOut[1])
			continue
		}

		passOut := pass.FindStringSubmatch(line)
		if passOut != nil {
			msec := tnow.Sub(tests[additionalTestName+passOut[1]])
			fmt.Fprintf(output, "##teamcity[testFinished name='%s' duration='%d']\n",
				additionalTestName+passOut[1], int(msec.Seconds()*1000))
			continue
		}

		skipOut := skip.FindStringSubmatch(line)
		if skipOut != nil {
			fmt.Fprintf(output, "##teamcity[testIgnored timestamp='%s' name='%s']\n", now,
				additionalTestName+skipOut[1])
			continue
		}

		failOut := fail.FindStringSubmatch(line)
		if failOut != nil {
			msec := tnow.Sub(tests[additionalTestName+passOut[1]])
			fmt.Fprintf(output, "##teamcity[testFailed name='%s' duration='%d']\n",
				additionalTestName+failOut[1], int(msec.Seconds()*1000))
			continue
		}

		fmt.Fprintln(output, line)
	}
}
