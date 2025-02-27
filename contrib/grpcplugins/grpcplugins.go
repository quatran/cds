package grpcplugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ovh/cds/sdk"
)

func GetRunResults(workerHTTPPort int32) ([]sdk.WorkflowRunResult, error) {
	if workerHTTPPort == 0 {
		return nil, nil
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/run-result", workerHTTPPort), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request to get run result: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot get run result /run-result: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body on get run result /run-result: %v", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot get run result /run-result: HTTP %d", resp.StatusCode)
	}

	var results []sdk.WorkflowRunResult
	if err := sdk.JSONUnmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response: %v", err)
	}
	return results, nil
}

// SendVulnerabilityReport call worker to send vulnerabiliry report to API
func SendVulnerabilityReport(workerHTTPPort int32, report sdk.VulnerabilityWorkerReport) error {
	if workerHTTPPort == 0 {
		return nil
	}

	data, errD := json.Marshal(report)
	if errD != nil {
		return fmt.Errorf("unable to marshal report: %v", errD)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/vulnerability", workerHTTPPort), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send report to worker /vulnerability: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send report to worker /vulnerability: %v", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("cannot send report to worker /vulnerability: HTTP %d", resp.StatusCode)
	}

	return nil
}
