package test

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
)

const (
	failuresFormatName  = "failures"
	defaultPassInterval = 50
	passIntervalEnvVar  = "GODOG_PASS_INTERVAL"
)

func init() {
	formatters.Format(failuresFormatName, "Shows only failed steps with periodic pass counts.", failuresFormatterFunc)
}

type failureDetail struct {
	uri      string
	scenario string
	step     string
	err      error
}

type failuresFormatter struct {
	out          io.Writer
	passInterval int
	lock         sync.Mutex
	passedSteps  int
	failedSteps  []failureDetail
}

func failuresFormatterFunc(_ string, out io.Writer) formatters.Formatter {
	return &failuresFormatter{
		out:          out,
		passInterval: readPassInterval(),
	}
}

func readPassInterval() int {
	if env := os.Getenv(passIntervalEnvVar); env != "" {
		if value, err := strconv.Atoi(env); err == nil && value > 0 {
			return value
		}
	}
	return defaultPassInterval
}

func (f *failuresFormatter) TestRunStarted() {}

func (f *failuresFormatter) Feature(*messages.GherkinDocument, string, []byte) {}

func (f *failuresFormatter) Pickle(*messages.Pickle) {}

func (f *failuresFormatter) Defined(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

func (f *failuresFormatter) Passed(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.passedSteps++
	if f.passInterval > 0 && f.passedSteps%f.passInterval == 0 {
		fmt.Fprintf(f.out, "passed steps: %d\n", f.passedSteps)
	}
}

func (f *failuresFormatter) Skipped(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

func (f *failuresFormatter) Undefined(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

func (f *failuresFormatter) Pending(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

func (f *failuresFormatter) Ambiguous(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition, error) {
}

func (f *failuresFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, _ *formatters.StepDefinition, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	scenario := "<unknown scenario>"
	uri := ""
	stepText := "<unknown step>"
	if pickle != nil {
		scenario = pickle.Name
		uri = pickle.Uri
	}
	if step != nil {
		stepText = step.Text
	}

	f.failedSteps = append(f.failedSteps, failureDetail{
		uri:      uri,
		scenario: scenario,
		step:     stepText,
		err:      err,
	})
}

func (f *failuresFormatter) Summary() {
	f.lock.Lock()
	defer f.lock.Unlock()

	if len(f.failedSteps) == 0 {
		if f.passedSteps > 0 && f.passedSteps%f.passInterval != 0 {
			fmt.Fprintf(f.out, "passed steps: %d\n", f.passedSteps)
		}
		return
	}

	fmt.Fprintln(f.out, "")
	fmt.Fprintln(f.out, "Failed steps:")
	for index, failure := range f.failedSteps {
		line := fmt.Sprintf("%d) %s", index+1, failure.scenario)
		if failure.uri != "" {
			line = fmt.Sprintf("%s (%s)", line, failure.uri)
		}
		fmt.Fprintln(f.out, line)
		fmt.Fprintf(f.out, "   %s\n", failure.step)
		fmt.Fprintf(f.out, "   error: %v\n", failure.err)
	}
}
