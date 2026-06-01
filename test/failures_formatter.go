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

type stepEvent struct {
	kind     string
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
	stepEvents   []stepEvent
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

func (f *failuresFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, _ *formatters.StepDefinition) {
	f.recordStepEvent("skipped", pickle, step, nil)
}

func (f *failuresFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, _ *formatters.StepDefinition) {
	f.recordStepEvent("undefined", pickle, step, nil)
}

func (f *failuresFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, _ *formatters.StepDefinition) {
	f.recordStepEvent("pending", pickle, step, nil)
}

func (f *failuresFormatter) Ambiguous(pickle *messages.Pickle, step *messages.PickleStep, _ *formatters.StepDefinition, err error) {
	f.recordStepEvent("ambiguous", pickle, step, err)
}

func (f *failuresFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, _ *formatters.StepDefinition, err error) {
	f.recordStepEvent("failed", pickle, step, err)
}

func (f *failuresFormatter) recordStepEvent(kind string, pickle *messages.Pickle, step *messages.PickleStep, err error) {
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

	f.stepEvents = append(f.stepEvents, stepEvent{
		kind:     kind,
		uri:      uri,
		scenario: scenario,
		step:     stepText,
		err:      err,
	})
}

func (f *failuresFormatter) Summary() {
	f.lock.Lock()
	defer f.lock.Unlock()

	if len(f.stepEvents) == 0 {
		if f.passInterval > 0 && f.passedSteps > 0 && f.passedSteps%f.passInterval != 0 {
			fmt.Fprintf(f.out, "passed steps: %d\n", f.passedSteps)
		}
		return
	}

	fmt.Fprintln(f.out, "")
	fmt.Fprintln(f.out, "Non-passing steps:")
	for index, failure := range f.stepEvents {
		line := fmt.Sprintf("%d) %s: %s", index+1, failure.kind, failure.scenario)
		if failure.uri != "" {
			line = fmt.Sprintf("%s (%s)", line, failure.uri)
		}
		fmt.Fprintln(f.out, line)
		fmt.Fprintf(f.out, "   %s\n", failure.step)
		if failure.err != nil {
			fmt.Fprintf(f.out, "   error: %v\n", failure.err)
		}
	}
}
