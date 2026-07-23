package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"infomunge/internal/cli"
	"infomunge/internal/evaluator"
	resultoutput "infomunge/internal/output"
	standaloneplayground "infomunge/internal/playground"
	"infomunge/internal/runner"
	"infomunge/internal/runtimeio"
	"infomunge/pkg/formats"
)

type testContext struct {
	lastOutput      string
	lastStdout      string
	lastStderr      string
	lastExitCode    int
	lastRunnerMime  string
	inputContent    string
	stdinContent    string
	payloadContent  string
	payloadMime     string
	scriptContent   string
	serverURL       string
	serverClose     func()
	serverAPIKey    string
	serverRunInputs []map[string]string
	lastHTTPStatus  int
	workDir         string
	timeout         time.Duration // Timeout for script execution to prevent infinite loops

	runCLIOverride          func(args ...string) error
	runCLIWithStdinOverride func(stdinContent string, args ...string) error
}

var (
	godogBuildOnce sync.Once
	godogBinary    string
	godogBuildErr  error
)

func TestFeatures(t *testing.T) {
	if os.Getenv("INFOMUNGE_SKIP_GODOG") == "1" {
		t.Skip("skipping Godog in nested go test invocation")
	}

	// Support filtering via environment variables:
	//   GODOG_PATHS - comma-separated feature file paths (default: "features")
	//   GODOG_TAGS  - tag expression to filter scenarios (e.g., "@fast", "@slow and not @wip")
	//   GODOG_FORMAT - output format (default: "failures", options: any registered Godog formatter)
	//
	// Examples:
	//   GODOG_PATHS=features/date_time_functions.feature go test -v
	//   GODOG_TAGS=@fast go test -v
	//   GODOG_PATHS=features/string_functions.feature,features/array_functions.feature go test -v

	paths := []string{"features"}
	if envPaths := os.Getenv("GODOG_PATHS"); envPaths != "" {
		paths = strings.Split(envPaths, ",")
		for i := range paths {
			paths[i] = strings.TrimSpace(paths[i])
		}
	}

	opts := &godog.Options{
		Format:      resolveGodogFormat(),
		Paths:       paths,
		Output:      os.Stdout, // Direct output to stdout for immediate feedback
		Concurrency: resolveGodogConcurrency(),
	}

	if tags := os.Getenv("GODOG_TAGS"); tags != "" {
		opts.Tags = tags
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func resolveGodogFormat() string {
	if envFormat := os.Getenv("GODOG_FORMAT"); envFormat != "" {
		return envFormat
	}
	// failures formatter suppresses per-step names while still printing periodic pass counts.
	return failuresFormatName
}

func resolveGodogConcurrency() int {
	if envConcurrency := os.Getenv("GODOG_CONCURRENCY"); envConcurrency != "" {
		if parsed, err := strconv.Atoi(envConcurrency); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 4
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := &testContext{
		timeout: 5 * time.Second, // Default 5-second timeout for all script executions
	}
	quotedStepArgument := `((?:[^"\\]|\\.)*)`

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if tc.serverClose != nil {
			tc.serverClose()
			tc.serverClose = nil
			tc.serverURL = ""
		}
		if tc.workDir != "" {
			_ = os.RemoveAll(tc.workDir)
			tc.workDir = ""
		}
		return ctx, nil
	})

	// Steps for read_file.feature
	ctx.Step(`^a file named "`+quotedStepArgument+`" with content "`+quotedStepArgument+`"$`, tc.aFileNamedWithContent)
	ctx.Step(`^a file named "([^"]*)" with content:$`, tc.aFileNamedWithDocstringContent)
	ctx.Step(`^a file named "([^"]*)" with "([^"]*)" repeated (\d+) times$`, tc.aFileNamedWithRepeatedContent)
	ctx.Step(`^I run the application with "([^"]*)"$`, tc.iRunTheApplicationWith)
	ctx.Step(`^I run the application with "([^"]*)" and it fails$`, tc.iRunTheApplicationWithAndItFails)
	ctx.Step(`^I run the application with arguments "([^"]*)" and it fails$`, tc.iRunTheApplicationWithArgumentsAndItFails)
	ctx.Step(`^I inspect the application help$`, tc.iInspectTheApplicationHelp)
	ctx.Step(`^the server listen default should be "([^"]*)"$`, tc.theServerListenDefaultShouldBe)
	ctx.Step(`^I run repo-wide package discovery from the repo root$`, tc.iRunRepoWidePackageDiscoveryFromRepoRoot)
	ctx.Step(`^the output should contain "`+quotedStepArgument+`"$`, tc.theOutputShouldContain)
	ctx.Step(`^the Makefile cucumber targets should use a 5 minute Go test timeout$`, tc.theMakefileCucumberTargetsShouldUseAFiveMinuteGoTestTimeout)
	ctx.Step(`^the repo-wide go test regression step should use a 5 minute Go test timeout$`, tc.theRepoWideGoTestRegressionStepShouldUseAFiveMinuteGoTestTimeout)
	ctx.Step(`^the testing docs should show cucumber commands with a 5 minute Go test timeout$`, tc.theTestingDocsShouldShowCucumberCommandsWithAFiveMinuteGoTestTimeout)
	ctx.Step(`^the testing docs should show bounded repo-wide package test commands$`, tc.theTestingDocsShouldShowBoundedRepoWidePackageTestCommands)
	ctx.Step(`^the default mutation corpus test should be skipped outside soak mode$`, tc.theDefaultMutationCorpusTestShouldBeSkippedOutsideSoakMode)
	ctx.Step(`^the README DataWeave example should use a named file input and be executable$`, tc.theREADMEDataWeaveExampleShouldUseANamedFileInputAndBeExecutable)
	ctx.Step(`^I post the README server example$`, tc.iPostTheREADMEServerExample)
	ctx.Step(`^every repository path referenced by the extension guide should exist$`, tc.everyRepositoryPathReferencedByTheExtensionGuideShouldExist)

	// Steps for docstring_input.feature
	ctx.Step(`^the following input content:$`, tc.theFollowingInputContent)
	ctx.Step(`^the following stdin content:$`, tc.theFollowingStdinContent)
	ctx.Step(`^the stdin content is "([^"]*)" repeated (\d+) times$`, tc.theStdinContentIsRepeatedTimes)
	ctx.Step(`^a script with (\d+) repeated lines$`, tc.aScriptWithRepeatedLines)
	ctx.Step(`^I run the application with this content$`, tc.iRunTheApplicationWithThisContent)
	ctx.Step(`^I run the application with this content using --lazy and it fails$`, tc.iRunTheApplicationWithThisContentUsingLazyAndItFails)
	ctx.Step(`^I run the application with this content and inputs and it fails:$`, tc.iRunTheApplicationWithThisContentAndInputsAndItFails)
	ctx.Step(`^I run the application with this content and stdin-backed inputs and it fails:$`, tc.iRunTheApplicationWithThisContentAndStdinBackedInputsAndItFails)
	ctx.Step(`^I run the application and it fails$`, tc.iRunTheApplicationAndItFails)
	ctx.Step(`^the output should be:$`, tc.theOutputShouldBe)
	ctx.Step(`^the error should contain "`+quotedStepArgument+`"$`, tc.theOutputShouldContain)
	ctx.Step(`^the application should fail with error containing "`+quotedStepArgument+`"$`, tc.theApplicationShouldFailWithErrorContaining)

	// Steps for typed input data (xr3 refactoring)
	ctx.Step(`^the following XML input:$`, tc.theFollowingXMLInput)
	ctx.Step(`^the following JSON input:$`, tc.theFollowingJSONInput)
	ctx.Step(`^the following CSV input:$`, tc.theFollowingCSVInput)
	ctx.Step(`^the following YAML input:$`, tc.theFollowingYAMLInput)
	ctx.Step(`^the following properties input:$`, tc.theFollowingPropertiesInput)
	ctx.Step(`^the following NDJSON input:$`, tc.theFollowingNDJSONInput)
	ctx.Step(`^the following URL-encoded input:$`, tc.theFollowingURLEncodedInput)
	ctx.Step(`^the following multipart input:$`, tc.theFollowingMultipartInput)
	ctx.Step(`^the following script:$`, tc.theFollowingScript)
	ctx.Step(`^I run the script$`, tc.iRunTheScript)
	ctx.Step(`^I execute the script through the structured runner API$`, tc.iExecuteTheScriptThroughStructuredRunnerAPI)
	ctx.Step(`^the runner output MIME type should be "([^"]*)"$`, tc.theRunnerOutputMIMETypeShouldBe)
	ctx.Step(`^I run the script through the runner output path$`, tc.iRunTheScriptThroughRunnerOutputPath)
	ctx.Step(`^running the script through the runner output path should fail with error containing "`+quotedStepArgument+`"$`, tc.runningTheScriptThroughRunnerOutputPathShouldFailWithErrorContaining)
	ctx.Step(`^I run the script with a canceled evaluation context$`, tc.iRunTheScriptWithCanceledEvaluationContext)
	ctx.Step(`^I run the script with an expired evaluation deadline$`, tc.iRunTheScriptWithExpiredEvaluationDeadline)
	ctx.Step(`^I run the script with URL IO disabled$`, tc.iRunTheScriptWithURLIODisabled)
	ctx.Step(`^I run the script with readUrl redirecting to "([^"]*)"$`, tc.iRunTheScriptWithReadURLRedirectingTo)
	ctx.Step(`^running the script should fail with error containing "`+quotedStepArgument+`"$`, tc.runningTheScriptShouldFailWithErrorContaining)

	// Additional step for JSON input with script from input content
	ctx.Step(`^I run the application with this JSON input:$`, tc.iRunTheApplicationWithThisJSONInput)

	// Steps for lambda expressions
	ctx.Step(`^the output should contain a lambda function$`, tc.theOutputShouldContainLambda)

	// Steps for optional "run" subcommand
	ctx.Step(`^I run the application with the run subcommand$`, tc.iRunTheApplicationWithRunSubcommand)

	// Steps for filter operator
	ctx.Step(`^the output should be valid JSON$`, tc.theOutputShouldBeValidJSON)
	ctx.Step(`^the output should be valid JSON with array length of (\d+)$`, tc.theOutputShouldBeValidJSONWithArrayLength)
	ctx.Step(`^the output should be valid JSON with (\d+) keys$`, tc.theOutputShouldBeValidJSONWithKeys)
	ctx.Step(`^the output should not contain "((?:[^"\\]|\\.)*)"$`, tc.theOutputShouldNotContain)
	ctx.Step(`^the output should contain an error$`, tc.theOutputShouldContainError)
	ctx.Step(`^the output should be valid JSON with number: (-?\d+)$`, tc.theOutputShouldBeValidJSONWithNumber)
	ctx.Step(`^the output should be valid JSON with number close to (-?\d+(?:\.\d+)?)$`, tc.theOutputShouldBeValidJSONWithNumberCloseTo)
	ctx.Step(`^the output should be valid JSON boolean$`, tc.theOutputShouldBeValidJSONBoolean)
	ctx.Step(`^the output should be true$`, tc.theOutputShouldBeTrue)
	ctx.Step(`^the output should be false$`, tc.theOutputShouldBeFalse)
	ctx.Step(`^the output should be null$`, tc.theOutputShouldBeNull)
	ctx.Step(`^the output should be "([^"]*)"$`, tc.theOutputShouldBeString)
	ctx.Step(`^the output should be a valid RFC3339 timestamp$`, tc.theOutputShouldBeValidRFC3339Timestamp)
	ctx.Step(`^the output should have UTC offset minutes (-?\d+)$`, tc.theOutputShouldHaveUTCOffsetMinutes)
	ctx.Step(`^the stdout should be valid JSON equal to:$`, tc.theStdoutShouldBeValidJSONEqualTo)
	ctx.Step(`^the stderr should contain "`+quotedStepArgument+`"$`, tc.theStderrShouldContain)
	ctx.Step(`^the exit status should be (\d+)$`, tc.theExitStatusShouldBe)

	// Steps for namespace support
	ctx.Step(`^the output should contain:$`, tc.theOutputShouldContainDocstring)

	// Step for setting timeout
	ctx.Step(`^I set the execution timeout to (\d+) seconds$`, tc.iSetTheExecutionTimeoutToSeconds)

	// Steps for quick evaluation (date_formatting.feature)
	ctx.Step(`^the input "([^"]*)"$`, tc.theInputString)
	ctx.Step(`^the input (-?\d+(?:\.\d+)?)$`, tc.theInputNumber)
	ctx.Step(`^I evaluate "([^"]*)"$`, tc.iEvaluate)
	ctx.Step(`^the result should be "([^"]*)"$`, tc.theOutputShouldBeString)
	ctx.Step(`^the result should be (-?\d+)$`, tc.theOutputShouldBeValidJSONWithNumber)
	ctx.Step(`^the result should match "([^"]*)"$`, tc.theOutputShouldMatchRegex)
	ctx.Step(`^the script:$`, tc.theFollowingScript)
	ctx.Step(`^I execute the script$`, tc.iRunTheScript)

	// Steps for server playground
	ctx.Step(`^the server is running$`, tc.theServerIsRunning)
	ctx.Step(`^the server is running with API key "([^"]*)"$`, tc.theServerIsRunningWithAPIKey)
	ctx.Step(`^I request the playground page$`, tc.iRequestThePlaygroundPage)
	ctx.Step(`^I read the standalone playground page$`, tc.iReadTheStandalonePlaygroundPage)
	ctx.Step(`^the standalone input-card builder should match the server playground$`, tc.theStandaloneInputCardBuilderShouldMatchTheServerPlayground)
	ctx.Step(`^the standalone playground is served over HTTP$`, tc.theStandalonePlaygroundIsServedOverHTTP)
	ctx.Step(`^I load the standalone playground and its WebAssembly assets$`, tc.iLoadTheStandalonePlaygroundAndItsWebAssemblyAssets)
	ctx.Step(`^the response status should be (\d+)$`, tc.theResponseStatusShouldBe)
	ctx.Step(`^I run the server script without specifying output$`, tc.iRunTheServerScriptWithoutOutput)
	ctx.Step(`^I run the server script with output "([^"]*)"$`, tc.iRunTheServerScriptWithOutput)
	ctx.Step(`^the server run inputs are:$`, tc.theServerRunInputsAre)
	ctx.Step(`^I configure (\d+) server JSON inputs$`, tc.iConfigureServerJSONInputs)
	ctx.Step(`^I run the server script with configured inputs$`, tc.iRunTheServerScriptWithConfiguredInputs)
	ctx.Step(`^I run the server script without an API key$`, tc.iRunTheServerScriptWithoutAPIKey)
	ctx.Step(`^I run the server script with API key "([^"]*)"$`, tc.iRunTheServerScriptWithAPIKey)
	ctx.Step(`^I run the server script with bearer token "([^"]*)"$`, tc.iRunTheServerScriptWithBearerToken)
	ctx.Step(`^I run the server script with a canceled request context$`, tc.iRunTheServerScriptWithCanceledContext)
	ctx.Step(`^I run the server script with an oversized request body$`, tc.iRunTheServerScriptWithOversizedBody)
}

// Existing steps
func (tc *testContext) aFileNamedWithContent(fileName, content string) error {
	fileName = unescapeStepArgument(fileName)
	content = unescapeStepArgument(content)
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	filePath := filepath.Join(tc.workDir, fileName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

func (tc *testContext) aFileNamedWithDocstringContent(fileName string, content *godog.DocString) error {
	return tc.aFileNamedWithContent(fileName, content.Content)
}

func (tc *testContext) aFileNamedWithRepeatedContent(fileName, fragment string, count int) error {
	if count < 0 {
		return fmt.Errorf("repeat count must be non-negative, got %d", count)
	}
	return tc.aFileNamedWithContent(fileName, strings.Repeat(fragment, count))
}

func unescapeStepArgument(value string) string {
	if !strings.Contains(value, `\`) {
		return value
	}
	unquoted, err := strconv.Unquote(`"` + value + `"`)
	if err != nil {
		return value
	}
	return unquoted
}

func (tc *testContext) iRunTheApplicationWith(arg string) error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	err := tc.runCLI("-f", arg)
	if err != nil {
		return fmt.Errorf("app failed: %v, output: %s", err, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) iRunTheApplicationWithAndItFails(arg string) error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	return tc.expectCLIFailure(tc.runCLI("-f", arg))
}

func (tc *testContext) iRunTheApplicationWithArgumentsAndItFails(argsLine string) error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	return tc.expectCLIFailure(tc.runCLI(strings.Fields(argsLine)...))
}

func (tc *testContext) iInspectTheApplicationHelp() error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	_ = tc.runCLI("--help")
	if tc.lastOutput == "" {
		return errors.New("application help produced no output")
	}
	return nil
}

func (tc *testContext) theServerListenDefaultShouldBe(expected string) error {
	defaultText := "-listen string"
	valueText := fmt.Sprintf(`(default %q)`, expected)
	if !strings.Contains(tc.lastOutput, defaultText) || !strings.Contains(tc.lastOutput, valueText) {
		return fmt.Errorf("expected server listen help to contain %q and %q, got %q", defaultText, valueText, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) iRunRepoWidePackageDiscoveryFromRepoRoot() error {
	ctx, cancel := context.WithTimeout(context.Background(), repoWideGoTestTimeout)
	defer cancel()

	cmd := repoWideGoTestCommand(ctx)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	tc.lastStdout = stdout.String()
	tc.lastStderr = stderr.String()
	tc.lastOutput = tc.lastStdout + tc.lastStderr
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("repo-wide package discovery timed out after %s, output: %s", repoWideGoTestTimeout, tc.lastOutput)
		}
		return fmt.Errorf("repo-wide package discovery failed: %v, output: %s", err, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theOutputShouldContain(expected string) error {
	expected = unescapeStepArgument(expected)

	if !strings.Contains(tc.lastOutput, expected) {
		return fmt.Errorf("expected output to contain %q, but got %q", expected, tc.lastOutput)
	}
	return nil
}

// New steps for docstrings
func (tc *testContext) theFollowingInputContent(content *godog.DocString) error {
	tc.inputContent = content.Content
	return nil
}

func (tc *testContext) theFollowingStdinContent(content *godog.DocString) error {
	tc.stdinContent = content.Content
	return nil
}

func (tc *testContext) theStdinContentIsRepeatedTimes(fragment string, count int) error {
	if count < 0 {
		return fmt.Errorf("repeat count must be non-negative, got %d", count)
	}
	tc.stdinContent = strings.Repeat(fragment, count)
	return nil
}

func (tc *testContext) aScriptWithRepeatedLines(count int) error {
	if count <= 0 {
		return fmt.Errorf("line count must be positive, got %d", count)
	}
	builder := strings.Builder{}
	builder.WriteString("%im 0.1\n")
	builder.WriteString("output application/json\n")
	builder.WriteString("---\n")
	for i := 0; i < count; i++ {
		builder.WriteString("1 + 1\n")
	}
	tc.inputContent = builder.String()
	return nil
}

func (tc *testContext) iRunTheApplicationWithThisContent() error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	// Write inputContent to a temp file
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}
	return tc.iRunTheApplicationWith("input.txt")
}

func (tc *testContext) iRunTheApplicationWithThisContentUsingLazyAndItFails() error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}
	return tc.expectCLIFailure(tc.runCLI("--lazy", "-f", "input.txt"))
}

func (tc *testContext) iRunTheApplicationWithThisContentAndInputsAndItFails(content *godog.DocString) error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}

	args := []string{"-f", "input.txt"}
	lines := strings.Split(content.Content, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		args = append(args, "-i", line)
	}

	return tc.expectCLIFailure(tc.runCLI(args...))
}

func (tc *testContext) iRunTheApplicationWithThisContentAndStdinBackedInputsAndItFails(content *godog.DocString) error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}

	args := []string{"-f", "input.txt"}
	lines := strings.Split(content.Content, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		args = append(args, "-i", line)
	}

	return tc.expectCLIFailure(tc.runCLIWithStdin(tc.stdinContent, args...))
}

func (tc *testContext) iRunTheApplicationAndItFails() error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}
	return tc.expectCLIFailure(tc.runCLI("-f", "input.txt"))
}

func (tc *testContext) iRunTheApplicationWithRunSubcommand() error {
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}
	err := tc.runCLI("run", "-f", "input.txt")
	if err != nil {
		return fmt.Errorf("app failed: %v, output: %s", err, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theOutputShouldBe(expected *godog.DocString) error {
	if tc.lastOutput != expected.Content {
		return fmt.Errorf("expected output:\n%s\n\nbut got:\n%s", expected.Content, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theApplicationShouldFailWithErrorContaining(expected string) error {
	expected = unescapeStepArgument(expected)
	if err := tc.ensureWorkspace(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tc.workDir, "input.txt"), []byte(tc.inputContent), 0644); err != nil {
		return err
	}
	err := tc.runCLI("-f", "input.txt")
	if err := tc.expectCLIFailure(err); err != nil {
		return err
	}
	if !strings.Contains(tc.lastOutput, expected) {
		return fmt.Errorf("expected error to contain %q, but got: %s", expected, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) expectCLIFailure(err error) error {
	if err == nil {
		return fmt.Errorf("expected application to fail, but it succeeded with exit status %d (stdout: %s, stderr: %s)", tc.lastExitCode, tc.lastStdout, tc.lastStderr)
	}
	if tc.lastExitCode == 0 {
		return fmt.Errorf("expected application failure to set a non-zero exit status, but got 0 (error: %v, stdout: %s, stderr: %s)", err, tc.lastStdout, tc.lastStderr)
	}
	return nil
}

func (tc *testContext) runCLI(args ...string) error {
	if tc.runCLIOverride != nil {
		return tc.runCLIOverride(args...)
	}

	binPath, err := ensureGodogBinary()
	if err != nil {
		return err
	}

	cmd := exec.Command(binPath, args...)
	cmd.Dir = tc.workDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	tc.lastStdout = stdout.String()
	tc.lastStderr = stderr.String()
	tc.lastOutput = tc.lastStdout + tc.lastStderr
	tc.lastExitCode = 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			tc.lastExitCode = exitErr.ExitCode()
		} else {
			tc.lastExitCode = -1
		}
	}
	return err
}

func (tc *testContext) runCLIWithStdin(stdinContent string, args ...string) error {
	if tc.runCLIWithStdinOverride != nil {
		return tc.runCLIWithStdinOverride(stdinContent, args...)
	}

	binPath, err := ensureGodogBinary()
	if err != nil {
		return err
	}

	cmd := exec.Command(binPath, args...)
	cmd.Dir = tc.workDir
	cmd.Stdin = strings.NewReader(stdinContent)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	tc.lastStdout = stdout.String()
	tc.lastStderr = stderr.String()
	tc.lastOutput = tc.lastStdout + tc.lastStderr
	tc.lastExitCode = 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			tc.lastExitCode = exitErr.ExitCode()
		} else {
			tc.lastExitCode = -1
		}
	}
	return err
}

// Typed input step definitions for xr3 refactoring
func (tc *testContext) theFollowingXMLInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "application/xml"
	return nil
}

func (tc *testContext) theFollowingJSONInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "application/json"
	return nil
}

func (tc *testContext) theFollowingCSVInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "application/csv"
	return nil
}

func (tc *testContext) theFollowingYAMLInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "application/yaml"
	return nil
}

func (tc *testContext) theFollowingPropertiesInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "text/x-java-properties"
	return nil
}

func (tc *testContext) theFollowingNDJSONInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "application/x-ndjson"
	return nil
}

func (tc *testContext) theFollowingURLEncodedInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "application/x-www-form-urlencoded"
	return nil
}

func (tc *testContext) theFollowingMultipartInput(content *godog.DocString) error {
	tc.payloadContent = content.Content
	tc.payloadMime = "multipart/form-data"
	return nil
}

func (tc *testContext) theFollowingScript(content *godog.DocString) error {
	tc.scriptContent = content.Content
	return nil
}

func (tc *testContext) iRunTheScript() error {
	// Build context with payload
	ctx := make(evaluator.Context)

	// Parse and inject payload if mime type was set (input step was called)
	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	// Run the script with timeout protection
	return tc.runScriptWithTimeout(tc.scriptContent, ctx)
}

func (tc *testContext) iExecuteTheScriptThroughStructuredRunnerAPI() error {
	ctx := make(evaluator.Context)
	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	result, err := runner.ExecuteString(context.Background(), tc.scriptContent, ctx, runner.RunnerOptions{})
	if err != nil {
		return err
	}
	tc.lastRunnerMime = result.OutputMimeType

	formatted, err := runner.FormatExecutionResult(result)
	if err != nil {
		return err
	}
	tc.lastOutput = formatted
	return nil
}

func (tc *testContext) theRunnerOutputMIMETypeShouldBe(expected string) error {
	if tc.lastRunnerMime != expected {
		return fmt.Errorf("expected runner output MIME type %q, got %q", expected, tc.lastRunnerMime)
	}
	return nil
}

// iRunTheScriptThroughRunnerOutputPath exercises the output adapter contract
// without depending on the deprecated stdout-printing runner wrappers.
func (tc *testContext) iRunTheScriptThroughRunnerOutputPath() error {
	ctx := make(evaluator.Context)
	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	formatted, runErr := formatRunnerOutputPath(tc.scriptContent, ctx)
	if runErr != nil {
		return runErr
	}
	tc.lastOutput = formatted
	return nil
}

func (tc *testContext) runningTheScriptThroughRunnerOutputPathShouldFailWithErrorContaining(expected string) error {
	expected = unescapeStepArgument(expected)
	ctx := make(evaluator.Context)
	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			if strings.Contains(err.Error(), expected) {
				return nil
			}
			return fmt.Errorf("payload parsing failed with unexpected error: %v", err)
		}
		ctx["payload"] = payload
	}

	_, runErr := formatRunnerOutputPath(tc.scriptContent, ctx)
	if runErr == nil {
		return fmt.Errorf("expected script to fail, but it succeeded")
	}
	if !strings.Contains(runErr.Error(), expected) {
		return fmt.Errorf("expected error to contain %q, but got: %v", expected, runErr)
	}
	return nil
}

func formatRunnerOutputPath(script string, ctx evaluator.Context) (string, error) {
	if err := runner.RequireScriptHeader(script); err != nil {
		return "", err
	}
	result, err := runner.ExecuteString(context.Background(), script, ctx, runner.RunnerOptions{})
	if err != nil {
		return "", err
	}
	return runner.FormatExecutionResult(result)
}

func executeRunnerValue(goCtx context.Context, script string, ctx evaluator.Context, opts runner.RunnerOptions) (evaluator.Value, error) {
	result, err := runner.ExecuteString(goCtx, script, ctx, opts)
	if err != nil {
		return nil, err
	}
	resolved, err := result.Resolved()
	if err != nil {
		return nil, err
	}
	return resolved.Value, nil
}

func (tc *testContext) iRunTheScriptWithCanceledEvaluationContext() error {
	ctx := make(evaluator.Context)

	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	goCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executeRunnerValue(goCtx, tc.scriptContent, ctx, runner.RunnerOptions{})
	if err == nil {
		return fmt.Errorf("expected script to fail, but it succeeded")
	}

	tc.lastOutput = err.Error()
	return nil
}

func (tc *testContext) iRunTheScriptWithExpiredEvaluationDeadline() error {
	ctx := make(evaluator.Context)

	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	goCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := executeRunnerValue(goCtx, tc.scriptContent, ctx, runner.RunnerOptions{})
	if err == nil {
		return fmt.Errorf("expected script to fail, but it succeeded")
	}

	tc.lastOutput = err.Error()
	return nil
}

func (tc *testContext) iRunTheScriptWithURLIODisabled() error {
	ctx := make(evaluator.Context)

	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	_, err := executeRunnerValue(
		context.Background(),
		tc.scriptContent,
		ctx,
		runner.RunnerOptions{DisableURLReadService: true},
	)
	if err == nil {
		return fmt.Errorf("expected script to fail, but it succeeded")
	}

	tc.lastOutput = err.Error()
	return nil
}

func (tc *testContext) iRunTheScriptWithReadURLRedirectingTo(target string) error {
	ctx := make(evaluator.Context)

	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			return fmt.Errorf("failed to parse payload: %v", err)
		}
		ctx["payload"] = payload
	}

	redirectClient := &http.Client{
		Transport: redirectRoundTripper{target: target},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return runtimeio.ValidateURL(req.URL)
		},
	}
	service := &runtimeio.URLReadService{
		Client:        redirectClient,
		FormatService: runtimeio.FormatService{},
	}

	result, err := runner.ExecuteString(
		context.Background(),
		tc.scriptContent,
		ctx,
		runner.RunnerOptions{URLReadService: service},
	)
	if err != nil {
		tc.lastOutput = err.Error()
		return nil
	}

	formatted, err := formatTestExecutionResult(result)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}
	tc.lastOutput = formatted
	return nil
}

type redirectRoundTripper struct {
	target string
}

func (rt redirectRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusFound,
		Status:     "302 Found",
		Header:     http.Header{"Location": []string{rt.target}},
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    req,
	}, nil
}

func (tc *testContext) runningTheScriptShouldFailWithErrorContaining(expected string) error {
	expected = unescapeStepArgument(expected)
	// Build context with payload
	ctx := make(evaluator.Context)

	// Parse and inject payload if mime type was set (input step was called)
	if tc.payloadMime != "" {
		payload, err := formats.Read(tc.payloadContent, tc.payloadMime)
		if err != nil {
			// If parsing fails with expected error, that's the error we're looking for
			if strings.Contains(err.Error(), expected) {
				return nil
			}
			return fmt.Errorf("payload parsing failed with unexpected error: %v", err)
		}
		ctx["payload"] = payload
	}

	// Run the script with injected context
	goCtx, cancel := context.WithTimeout(context.Background(), tc.timeout)
	defer cancel()
	_, err := executeRunnerValue(goCtx, tc.scriptContent, ctx, runner.RunnerOptions{})
	if err == nil {
		return fmt.Errorf("expected script to fail, but it succeeded")
	}

	if !strings.Contains(err.Error(), expected) {
		return fmt.Errorf("expected error to contain %q, but got: %v", expected, err)
	}

	return nil
}

func (tc *testContext) iRunTheApplicationWithThisJSONInput(jsonInput *godog.DocString) error {
	// Build context with payload
	ctx := make(evaluator.Context)

	// Parse JSON input as payload
	payload, err := formats.Read(jsonInput.Content, "application/json")
	if err != nil {
		return fmt.Errorf("failed to parse JSON input: %v", err)
	}
	ctx["payload"] = payload

	// Run the script with timeout protection
	return tc.runScriptWithTimeout(tc.inputContent, ctx)
}

func (tc *testContext) theOutputShouldContainLambda() error {
	trimmed := strings.TrimSpace(tc.lastOutput)
	if trimmed == "" {
		return fmt.Errorf("expected lambda expression output, but got empty output")
	}
	if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
		return fmt.Errorf("expected lambda expression to evaluate without error, but got: %s", tc.lastOutput)
	}

	// Lambda functions are often serialized as a JSON string like: "(lambda: [x] => x + 1)"
	decoded := trimmed
	if strings.HasPrefix(decoded, "\"") && strings.HasSuffix(decoded, "\"") {
		if unquoted, err := strconv.Unquote(decoded); err == nil {
			decoded = unquoted
		}
	}

	if strings.Contains(decoded, "(lambda:") && strings.Contains(decoded, "=>") {
		return nil
	}

	// Some outputs fall back to Go's default struct formatting for lambdas.
	if strings.HasPrefix(trimmed, "&{[") && strings.Contains(trimmed, "map[") {
		return nil
	}

	if !strings.Contains(decoded, "(lambda:") || !strings.Contains(decoded, "=>") {
		return fmt.Errorf("expected output to contain lambda representation, but got: %s", tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeValidJSON() error {
	// Parse output as JSON to verify it's valid
	var result evaluator.Value
	if err := json.Unmarshal([]byte(tc.lastOutput), &result); err != nil {
		return fmt.Errorf("expected valid JSON, but got invalid JSON: %v, output: %s", err, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeValidJSONWithArrayLength(lengthStr string) error {
	// Parse output as JSON to verify it's valid and is an array of expected length
	var arr evaluator.Array
	if err := json.Unmarshal([]byte(tc.lastOutput), &arr); err != nil {
		return fmt.Errorf("expected valid JSON array, but got invalid JSON: %v, output: %s", err, tc.lastOutput)
	}

	expectedLen, err := strconv.Atoi(lengthStr)
	if err != nil {
		return fmt.Errorf("invalid expected length: %v", err)
	}

	if len(arr) != expectedLen {
		return fmt.Errorf("expected array of length %d, but got %d", expectedLen, len(arr))
	}
	return nil
}

func (tc *testContext) theOutputShouldNotContain(notExpected string) error {
	// Unquote if the string contains escaped quotes
	if strings.Contains(notExpected, `\"`) {
		unquoted, err := strconv.Unquote(`"` + notExpected + `"`)
		if err == nil {
			notExpected = unquoted
		}
	}

	if strings.Contains(tc.lastOutput, notExpected) {
		return fmt.Errorf("expected output to not contain %q, but it did. Output: %s", notExpected, tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeValidJSONWithKeys(keysStr string) error {
	var obj evaluator.Context
	if err := json.Unmarshal([]byte(tc.lastOutput), &obj); err != nil {
		return fmt.Errorf("expected valid JSON object, but got invalid JSON: %v, output: %s", err, tc.lastOutput)
	}

	expectedKeys, err := strconv.Atoi(keysStr)
	if err != nil {
		return fmt.Errorf("invalid expected key count: %v", err)
	}

	if len(obj) != expectedKeys {
		return fmt.Errorf("expected object with %d keys, but got %d", expectedKeys, len(obj))
	}
	return nil
}

func (tc *testContext) theOutputShouldContainError() error {
	// Check if the output contains an error message (when the app fails)
	// The application should have failed and produced an error
	if !strings.Contains(tc.lastOutput, "error") && !strings.Contains(tc.lastOutput, "Error") {
		return fmt.Errorf("expected output to contain an error, but got: %s", tc.lastOutput)
	}
	return nil
}

func (tc *testContext) theOutputShouldContainDocstring(content *godog.DocString) error {
	if !strings.Contains(tc.lastOutput, content.Content) {
		return fmt.Errorf("expected output to contain:\n%s\n\nbut got:\n%s", content.Content, tc.lastOutput)
	}
	return nil
}

// iSetTheExecutionTimeoutToSeconds sets the timeout for script execution
func (tc *testContext) iSetTheExecutionTimeoutToSeconds(seconds string) error {
	secInt, err := strconv.Atoi(seconds)
	if err != nil {
		return fmt.Errorf("invalid timeout value: %v", err)
	}
	tc.timeout = time.Duration(secInt) * time.Second
	return nil
}

// theOutputShouldBeValidJSONWithNumber validates that the output is a specific JSON number
func (tc *testContext) theOutputShouldBeValidJSONWithNumber(numStr string) error {
	var num json.Number
	if err := json.Unmarshal([]byte(tc.lastOutput), &num); err != nil {
		return fmt.Errorf("expected valid JSON number, but got invalid JSON: %v, output: %s", err, tc.lastOutput)
	}

	expectedNum, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expected number: %v", err)
	}

	actual, ok := new(big.Rat).SetString(num.String())
	if !ok {
		return fmt.Errorf("expected valid JSON number, but got unparsable number: %s", num.String())
	}
	expected := new(big.Rat).SetInt64(expectedNum)
	if actual.Cmp(expected) != 0 {
		return fmt.Errorf("expected number %d, but got %s", expectedNum, num.String())
	}
	return nil
}

func (tc *testContext) theOutputShouldBeValidJSONWithNumberCloseTo(numStr string) error {
	var num evaluator.Value
	if err := json.Unmarshal([]byte(tc.lastOutput), &num); err != nil {
		return fmt.Errorf("expected valid JSON number, but got invalid JSON: %v, output: %s", err, tc.lastOutput)
	}

	expectedNum, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return fmt.Errorf("invalid expected number: %v", err)
	}

	switch v := num.(type) {
	case float64:
		const tol = 1e-9
		diff := math.Abs(v - expectedNum)
		if diff > tol && diff > math.Abs(expectedNum)*tol {
			return fmt.Errorf("expected number close to %v, but got %v", expectedNum, v)
		}
		return nil
	default:
		return fmt.Errorf("expected JSON number, but got %T: %v", num, num)
	}
}

func (tc *testContext) theOutputShouldBeValidJSONBoolean() error {
	var value evaluator.Value
	if err := json.Unmarshal([]byte(tc.lastOutput), &value); err != nil {
		return fmt.Errorf("expected valid JSON boolean, but got invalid JSON: %v, output: %s", err, tc.lastOutput)
	}

	if _, ok := value.(bool); !ok {
		return fmt.Errorf("expected JSON boolean, but got %T: %v", value, value)
	}
	return nil
}

// runScriptWithTimeout runs a script with the configured timeout
func (tc *testContext) runScriptWithTimeout(scriptContent string, additionalContext evaluator.Context) error {
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	if additionalContext == nil {
		additionalContext = make(evaluator.Context)
	}

	// Run script in a goroutine
	go func() {
		goCtx, cancel := context.WithTimeout(context.Background(), tc.timeout)
		defer cancel()
		result, err := runner.ExecuteString(goCtx, scriptContent, additionalContext, runner.RunnerOptions{})
		if err != nil {
			errChan <- err
		} else {
			formatted, err := formatTestExecutionResult(result)
			if err != nil {
				errChan <- fmt.Errorf("failed to format output: %v", err)
				return
			}
			resultChan <- formatted
		}
	}()

	// Wait for result with timeout
	select {
	case formatted := <-resultChan:
		tc.lastOutput = formatted
		return nil
	case err := <-errChan:
		return err
	case <-time.After(tc.timeout):
		return fmt.Errorf("script execution timed out after %v (possible infinite loop)", tc.timeout)
	}
}

func formatTestExecutionResult(result runner.ExecutionResult) (string, error) {
	if result.HasHeader {
		return runner.FormatExecutionResult(result)
	}
	resolved, err := result.Resolved()
	if err != nil {
		return "", err
	}
	return resultoutput.FormatResultWithMetadata(resolved.Value, resolved.OutputMimeType, resolved.Context, resolved.OutputMetadata)
}

func (tc *testContext) theOutputShouldBeTrue() error {
	trimmed := strings.TrimSpace(tc.lastOutput)
	if trimmed != "true" {
		return fmt.Errorf("expected output to be 'true', but got '%s'", trimmed)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeFalse() error {
	trimmed := strings.TrimSpace(tc.lastOutput)
	if trimmed != "false" {
		return fmt.Errorf("expected output to be 'false', but got '%s'", trimmed)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeNull() error {
	trimmed := strings.TrimSpace(tc.lastOutput)
	if trimmed != "null" {
		return fmt.Errorf("expected output to be 'null', but got '%s'", trimmed)
	}
	return nil
}

func (tc *testContext) theOutputShouldBeString(expected string) error {
	trimmed := strings.TrimSpace(tc.lastOutput)
	// The output might be JSON-quoted, so check both raw and quoted forms
	if trimmed == expected || trimmed == fmt.Sprintf("%q", expected) {
		return nil
	}
	return fmt.Errorf("expected output to be %q, but got '%s'", expected, trimmed)
}

func (tc *testContext) theStdoutShouldBeValidJSONEqualTo(expected *godog.DocString) error {
	var got evaluator.Value
	if err := json.Unmarshal([]byte(tc.lastStdout), &got); err != nil {
		return fmt.Errorf("expected stdout to be valid JSON, but got invalid JSON: %v, stdout: %s, stderr: %s", err, tc.lastStdout, tc.lastStderr)
	}

	var want evaluator.Value
	if err := json.Unmarshal([]byte(expected.Content), &want); err != nil {
		return fmt.Errorf("expected JSON literal in step is invalid: %v, content: %s", err, expected.Content)
	}

	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("expected stdout JSON %v, but got %v (stdout: %s, stderr: %s)", want, got, tc.lastStdout, tc.lastStderr)
	}
	return nil
}

func (tc *testContext) theStderrShouldContain(expected string) error {
	expected = unescapeStepArgument(expected)

	if !strings.Contains(tc.lastStderr, expected) {
		return fmt.Errorf("expected stderr to contain %q, but got %q", expected, tc.lastStderr)
	}
	return nil
}

func (tc *testContext) theExitStatusShouldBe(expected int) error {
	if tc.lastExitCode != expected {
		return fmt.Errorf("expected exit status %d, but got %d (stdout: %s, stderr: %s)", expected, tc.lastExitCode, tc.lastStdout, tc.lastStderr)
	}
	return nil
}

func (tc *testContext) theInputString(input string) error {
	tc.payloadContent = input
	tc.payloadMime = "text/plain"
	return nil
}

func (tc *testContext) theInputNumber(input string) error {
	tc.payloadContent = input
	tc.payloadMime = "application/json" // Numbers are valid JSON
	return nil
}

func (tc *testContext) iEvaluate(expression string) error {
	tc.scriptContent = fmt.Sprintf("%%dw 2.0\noutput application/json\n---\n%s", expression)
	return tc.iRunTheScript()
}

func (tc *testContext) theOutputShouldMatchRegex(pattern string) error {
	trimmed := strings.TrimSpace(tc.lastOutput)
	// Unquote if it's a JSON string
	if strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
		if unquoted, err := strconv.Unquote(trimmed); err == nil {
			trimmed = unquoted
		}
	}

	matched, err := regexp.MatchString(pattern, trimmed)
	if err != nil {
		return fmt.Errorf("invalid regex pattern %q: %v", pattern, err)
	}
	if !matched {
		return fmt.Errorf("expected output %q to match pattern %q", trimmed, pattern)
	}
	return nil
}

func (tc *testContext) outputAsString() (string, error) {
	trimmed := strings.TrimSpace(tc.lastOutput)
	if trimmed == "" {
		return "", fmt.Errorf("expected non-empty output")
	}
	if strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
		unquoted, err := strconv.Unquote(trimmed)
		if err != nil {
			return "", fmt.Errorf("expected valid JSON string output, got %q: %v", trimmed, err)
		}
		return unquoted, nil
	}
	return trimmed, nil
}

func (tc *testContext) theOutputShouldBeValidRFC3339Timestamp() error {
	value, err := tc.outputAsString()
	if err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return fmt.Errorf("expected RFC3339 timestamp, got %q: %v", value, err)
	}
	return nil
}

func (tc *testContext) theOutputShouldHaveUTCOffsetMinutes(expectedMinutes int) error {
	value, err := tc.outputAsString()
	if err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return fmt.Errorf("expected RFC3339 timestamp before checking offset, got %q: %v", value, err)
	}
	_, offsetSeconds := parsed.Zone()
	actualMinutes := offsetSeconds / 60
	if actualMinutes != expectedMinutes {
		return fmt.Errorf("expected UTC offset %d minutes, got %d minutes in %q", expectedMinutes, actualMinutes, value)
	}
	return nil
}

func (tc *testContext) theServerIsRunning() error {
	if tc.serverClose != nil {
		return nil
	}
	app := cli.NewApp()
	config := &cli.Config{
		Lazy: false,
	}
	server := httptest.NewServer(app.ServerHandler(config))
	tc.serverURL = server.URL
	tc.serverClose = server.Close
	tc.serverAPIKey = ""
	return nil
}

func (tc *testContext) theServerIsRunningWithAPIKey(apiKey string) error {
	if tc.serverClose != nil {
		tc.serverClose()
		tc.serverClose = nil
		tc.serverURL = ""
	}
	app := cli.NewApp()
	config := &cli.Config{
		Lazy:         false,
		ServerAPIKey: apiKey,
	}
	server := httptest.NewServer(app.ServerHandler(config))
	tc.serverURL = server.URL
	tc.serverClose = server.Close
	tc.serverAPIKey = apiKey
	return nil
}

func (tc *testContext) iRunTheServerScriptWithoutOutput() error {
	return tc.runServerScript(nil)
}

func (tc *testContext) iRunTheServerScriptWithOutput(output string) error {
	return tc.runServerScript(&output)
}

func (tc *testContext) runServerScript(output *string) error {
	return tc.runServerScriptWithAPIKey(output, nil)
}

func (tc *testContext) iRunTheServerScriptWithoutAPIKey() error {
	return tc.runServerScriptWithAPIKey(nil, nil)
}

func (tc *testContext) iRunTheServerScriptWithAPIKey(apiKey string) error {
	return tc.runServerScriptWithAPIKey(nil, &apiKey)
}

func (tc *testContext) iRunTheServerScriptWithBearerToken(apiKey string) error {
	return tc.runServerScriptWithBearerToken(nil, apiKey)
}

func (tc *testContext) runServerScriptWithAPIKey(output *string, apiKey *string) error {
	if tc.serverURL == "" {
		return fmt.Errorf("server is not running")
	}
	if strings.TrimSpace(tc.scriptContent) == "" {
		return fmt.Errorf("script is not set")
	}

	payload := evaluator.Context{
		"script": tc.scriptContent,
	}
	if output != nil {
		payload["output"] = *output
	}
	if len(tc.serverRunInputs) > 0 {
		payload["inputs"] = tc.serverRunInputs
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, tc.serverURL+"/run", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != nil {
		req.Header.Set("X-API-Key", *apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request /run: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	tc.lastHTTPStatus = resp.StatusCode
	tc.lastOutput = string(responseBody)
	return nil
}

func (tc *testContext) runServerScriptWithBearerToken(output *string, apiKey string) error {
	if tc.serverURL == "" {
		return fmt.Errorf("server is not running")
	}
	if strings.TrimSpace(tc.scriptContent) == "" {
		return fmt.Errorf("script is not set")
	}

	payload := evaluator.Context{
		"script": tc.scriptContent,
	}
	if output != nil {
		payload["output"] = *output
	}
	if len(tc.serverRunInputs) > 0 {
		payload["inputs"] = tc.serverRunInputs
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, tc.serverURL+"/run", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request /run: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	tc.lastHTTPStatus = resp.StatusCode
	tc.lastOutput = string(responseBody)
	return nil
}

func (tc *testContext) iRunTheServerScriptWithCanceledContext() error {
	if strings.TrimSpace(tc.scriptContent) == "" {
		return fmt.Errorf("script is not set")
	}

	payload := evaluator.Context{
		"script": tc.scriptContent,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	canceledCtx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(canceledCtx)

	rec := httptest.NewRecorder()
	app := cli.NewApp()
	config := &cli.Config{Lazy: false}
	app.ServerHandler(config).ServeHTTP(rec, req)

	tc.lastHTTPStatus = rec.Code
	tc.lastOutput = rec.Body.String()
	return nil
}

func (tc *testContext) iRunTheServerScriptWithOversizedBody() error {
	if tc.serverURL == "" {
		return fmt.Errorf("server is not running")
	}
	if strings.TrimSpace(tc.scriptContent) == "" {
		return fmt.Errorf("script is not set")
	}

	payload := evaluator.Context{
		"script": tc.scriptContent,
		"inputs": []map[string]string{
			{
				"name":    "payload",
				"format":  "text/plain",
				"content": strings.Repeat("x", 2*1024*1024),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(tc.serverURL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to request /run: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	tc.lastHTTPStatus = resp.StatusCode
	tc.lastOutput = string(responseBody)
	return nil
}

func (tc *testContext) theServerRunInputsAre(content *godog.DocString) error {
	var inputs []map[string]string
	if err := json.Unmarshal([]byte(content.Content), &inputs); err != nil {
		return fmt.Errorf("invalid server run inputs JSON: %v", err)
	}
	tc.serverRunInputs = inputs
	return nil
}

func (tc *testContext) iConfigureServerJSONInputs(countStr string) error {
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return fmt.Errorf("invalid input count %q: %v", countStr, err)
	}
	if count < 0 {
		return fmt.Errorf("input count must be non-negative")
	}

	inputs := make([]map[string]string, 0, count)
	for i := 0; i < count; i++ {
		inputs = append(inputs, map[string]string{
			"name":    fmt.Sprintf("input%d", i),
			"format":  "json",
			"content": fmt.Sprintf("%d", i),
		})
	}
	tc.serverRunInputs = inputs
	return nil
}

func (tc *testContext) iRunTheServerScriptWithConfiguredInputs() error {
	if strings.TrimSpace(tc.scriptContent) == "" {
		return fmt.Errorf("script is not set")
	}

	payload := evaluator.Context{
		"script": tc.scriptContent,
	}
	if len(tc.serverRunInputs) > 0 {
		payload["inputs"] = tc.serverRunInputs
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	app := cli.NewApp()
	config := &cli.Config{Lazy: false}
	app.ServerHandler(config).ServeHTTP(rec, req)

	tc.lastHTTPStatus = rec.Code
	tc.lastOutput = rec.Body.String()
	return nil
}

func (tc *testContext) iRequestThePlaygroundPage() error {
	if tc.serverURL == "" {
		return fmt.Errorf("server is not running")
	}
	resp, err := http.Get(tc.serverURL + "/")
	if err != nil {
		return fmt.Errorf("failed to request playground page: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	tc.lastHTTPStatus = resp.StatusCode
	tc.lastOutput = string(body)
	return nil
}

func (tc *testContext) iReadTheStandalonePlaygroundPage() error {
	body, err := os.ReadFile("../docs/playground/index.html")
	if err != nil {
		tc.lastHTTPStatus = 500
		return fmt.Errorf("failed to read standalone playground page: %v", err)
	}
	tc.lastHTTPStatus = 200
	tc.lastOutput = string(body)
	return nil
}

func (tc *testContext) theStandaloneInputCardBuilderShouldMatchTheServerPlayground() error {
	serverPage, err := os.ReadFile("../internal/cli/playground.html")
	if err != nil {
		return fmt.Errorf("failed to read server playground page: %v", err)
	}
	standalonePage, err := os.ReadFile("../docs/playground/index.html")
	if err != nil {
		return fmt.Errorf("failed to read standalone playground page: %v", err)
	}

	serverBuilder, err := extractInputCardBuilder(string(serverPage))
	if err != nil {
		return fmt.Errorf("server playground: %v", err)
	}
	standaloneBuilder, err := extractInputCardBuilder(string(standalonePage))
	if err != nil {
		return fmt.Errorf("standalone playground: %v", err)
	}
	if standaloneBuilder != serverBuilder {
		return fmt.Errorf("standalone input-card builder differs from server playground")
	}
	return nil
}

func extractInputCardBuilder(page string) (string, error) {
	const startMarker = "function createInputCard(index, preset) {"
	const endMarker = "function addInputCard(preset) {"
	start := strings.Index(page, startMarker)
	if start < 0 {
		return "", fmt.Errorf("input-card builder start marker not found")
	}
	end := strings.Index(page[start:], endMarker)
	if end < 0 {
		return "", fmt.Errorf("input-card builder end marker not found")
	}
	return strings.TrimSpace(page[start : start+end]), nil
}

func (tc *testContext) theStandalonePlaygroundIsServedOverHTTP() error {
	if tc.serverClose != nil {
		return nil
	}
	server := httptest.NewServer(standaloneplayground.Handler("../docs/playground"))
	tc.serverURL = server.URL
	tc.serverClose = server.Close
	return nil
}

func (tc *testContext) iLoadTheStandalonePlaygroundAndItsWebAssemblyAssets() error {
	if tc.serverURL == "" {
		return fmt.Errorf("standalone playground server is not running")
	}

	assets := []struct {
		path        string
		contentType string
	}{
		{path: "/", contentType: "text/html"},
		{path: "/wasm_exec.js", contentType: "text/javascript"},
		{path: "/infomunge.wasm", contentType: "application/wasm"},
	}

	var loaded strings.Builder
	for _, asset := range assets {
		resp, err := http.Get(tc.serverURL + asset.path)
		if err != nil {
			return fmt.Errorf("failed to load standalone playground asset %q: %v", asset.path, err)
		}
		_, readErr := io.Copy(io.Discard, resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("failed to read standalone playground asset %q: %v", asset.path, readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close standalone playground asset %q: %v", asset.path, closeErr)
		}
		if resp.StatusCode != http.StatusOK {
			tc.lastHTTPStatus = resp.StatusCode
			return fmt.Errorf("standalone playground asset %q returned status %d", asset.path, resp.StatusCode)
		}
		actualContentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(actualContentType, asset.contentType) {
			return fmt.Errorf("standalone playground asset %q has content type %q, want %q", asset.path, actualContentType, asset.contentType)
		}
		name := strings.TrimPrefix(asset.path, "/")
		if name == "" {
			name = "index.html"
		}
		fmt.Fprintf(&loaded, "%s: %s\n", name, asset.contentType)
	}

	tc.lastHTTPStatus = http.StatusOK
	tc.lastOutput = loaded.String()
	return nil
}

func (tc *testContext) theResponseStatusShouldBe(statusCode string) error {
	expected, err := strconv.Atoi(statusCode)
	if err != nil {
		return fmt.Errorf("invalid status code: %v", err)
	}
	if tc.lastHTTPStatus != expected {
		return fmt.Errorf("expected status %d, got %d", expected, tc.lastHTTPStatus)
	}
	return nil
}

func (tc *testContext) ensureWorkspace() error {
	if tc.workDir != "" {
		return nil
	}
	if err := os.MkdirAll("../tmp", 0755); err != nil {
		return fmt.Errorf("failed to create tmp directory: %v", err)
	}
	workDir, err := os.MkdirTemp("../tmp", "godog-scenario-")
	if err != nil {
		return fmt.Errorf("failed to create scenario workspace: %v", err)
	}
	tc.workDir = workDir

	// Set up modules directory so module imports resolve correctly.
	// Create a real modules/ dir, symlink project module subdirectories (dw/),
	// and write test-only module files.
	if err := setupWorkspaceModules(workDir); err != nil {
		return fmt.Errorf("failed to set up workspace modules: %v", err)
	}
	return nil
}

func setupWorkspaceModules(workDir string) error {
	modulesDir := filepath.Join(workDir, "modules")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return err
	}

	// Symlink project module subdirectories (e.g. dw/) into the workspace
	projectModules, err := filepath.Abs("../modules")
	if err != nil {
		return nil // non-fatal: skip if can't resolve
	}
	entries, err := os.ReadDir(projectModules)
	if err != nil {
		return nil // non-fatal: skip if modules dir doesn't exist
	}
	for _, entry := range entries {
		if entry.IsDir() {
			src := filepath.Join(projectModules, entry.Name())
			dst := filepath.Join(modulesDir, entry.Name())
			_ = os.Symlink(src, dst)
		}
	}

	// Write test-only module files used by import_directive.feature
	testModules := map[string]string{
		"MathUtils.im":   "%im 0.1\nvar offset = 5\nvar scale = 10\nvar settings = {\n  scale: 10,\n  offset: 5\n}\nfun double(x) = x * 2\nfun triple(x) = x * 3\nfun addOffset(x) = x + offset\nfun scaleAndAdd(x) = x * scale + offset\nfun scaleAndAddSettings(x) = x * settings.scale + settings.offset\n",
		"StringUtils.im": "%im 0.1\nfun greet(name) = \"Hello, \" ++ name\nfun shout(s) = upper(s)\n",
		"DoBlock.im":     "%im 0.1\nfun addOne(x) = do {\n  var result = x + 1\n  ---\n  result\n}\n",
	}
	for name, content := range testModules {
		if err := os.WriteFile(filepath.Join(modulesDir, name), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func ensureGodogBinary() (string, error) {
	godogBuildOnce.Do(func() {
		if err := os.MkdirAll("../tmp", 0755); err != nil {
			godogBuildErr = fmt.Errorf("failed to create tmp directory: %v", err)
			return
		}
		binaryPath := filepath.Join("..", "tmp", fmt.Sprintf("infomunge-godog-%d", os.Getpid()))
		absBinaryPath, err := filepath.Abs(binaryPath)
		if err != nil {
			godogBuildErr = fmt.Errorf("failed to resolve absolute binary path: %v", err)
			return
		}
		godogBinary = absBinaryPath
		buildCmd := exec.Command("go", "build", "-o", godogBinary, "../cmd/infomunge/main.go")
		output, err := buildCmd.CombinedOutput()
		if err != nil {
			godogBuildErr = fmt.Errorf("failed to build app: %v, output: %s", err, strings.TrimSpace(string(output)))
		}
	})
	return godogBinary, godogBuildErr
}
