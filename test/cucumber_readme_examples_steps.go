package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	readmeDataWeaveCommandPattern = regexp.MustCompile(`(?m)^dw run (-i=payload=payload\.json) "([^"]+)"$`)
	readmeServerBodyPattern       = regexp.MustCompile(`(?m)^  -d '([^']+)'$`)
)

func (tc *testContext) theREADMEDataWeaveExampleShouldUseANamedFileInputAndBeExecutable() error {
	readme, err := os.ReadFile("../README.md")
	if err != nil {
		return fmt.Errorf("read README: %w", err)
	}

	match := readmeDataWeaveCommandPattern.FindSubmatch(readme)
	if match == nil {
		return errors.New("README should contain a DataWeave command using -i=payload=payload.json")
	}
	if _, err := exec.LookPath("dw"); err != nil {
		return nil
	}

	payloadFile, err := os.CreateTemp("../tmp", "readme-dataweave-*.json")
	if err != nil {
		return fmt.Errorf("create DataWeave example payload: %w", err)
	}
	payloadPath := payloadFile.Name()
	defer os.Remove(payloadPath)
	if _, err := payloadFile.WriteString(`{"message":"hello"}`); err != nil {
		_ = payloadFile.Close()
		return fmt.Errorf("write DataWeave example payload: %w", err)
	}
	if err := payloadFile.Close(); err != nil {
		return fmt.Errorf("close DataWeave example payload: %w", err)
	}

	inputArg := strings.Replace(string(match[1]), "payload.json", payloadPath, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "dw", "run", inputArg, string(match[2]))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return errors.New("README DataWeave example timed out")
		}
		return fmt.Errorf("README DataWeave example failed: %w, output: %s", err, output)
	}
	if !strings.Contains(string(output), "<message>hello</message>") {
		return fmt.Errorf("README DataWeave example returned unexpected output: %s", output)
	}
	return nil
}

func (tc *testContext) iPostTheREADMEServerExample() error {
	if tc.serverURL == "" {
		return errors.New("server is not running")
	}
	readme, err := os.ReadFile("../README.md")
	if err != nil {
		return fmt.Errorf("read README: %w", err)
	}

	bodyMatch := readmeServerBodyPattern.FindSubmatch(readme)
	if bodyMatch == nil {
		return errors.New("README should contain a single-quoted curl -d body")
	}
	if !bytes.Contains(readme, []byte("curl -X POST http://localhost:8080/run")) ||
		!bytes.Contains(readme, []byte("-H 'Content-Type: application/json'")) ||
		!bytes.Contains(readme, []byte("-H 'X-API-Key: your-secret'")) {
		return errors.New("README server example should document the /run URL, JSON content type, and API key")
	}

	req, err := http.NewRequest(http.MethodPost, tc.serverURL+"/run", bytes.NewReader(bodyMatch[1]))
	if err != nil {
		return fmt.Errorf("create README server request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "your-secret")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("post README server example: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read README server response: %w", err)
	}
	tc.lastHTTPStatus = resp.StatusCode
	tc.lastOutput = string(responseBody)
	return nil
}
