package gromwell

import (
	"net/http"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"io"
	"os"
	"fmt"
	"bytes"
	"mime/multipart"
)

const engineEndpoint = "/api/engine/v1"
const versionEndpoint = "/version"

const workflowEndpoint = "/api/workflows/v1"
const statusEndpoint = "/status"
const outputsEndpoint = "/outputs"
const metadataEndpoint = "/metadata"
const abortEndpoint = "/abort"

func (client CromwellClient) AbortWorkflow(workflowId string) (WorkflowStatus, error) {
	var status WorkflowStatus
	body, err := getResponseBody(http.Post(client.makeWorkflowEndpointPath(workflowId, abortEndpoint), "text/plain", strings.NewReader("")))
	if (err != nil) {
		return status, err
	}
	err = json.Unmarshal(body, &status)
	if (err != nil) {
		return status, err
	}
	return status, err
}

func (client CromwellClient) Version() (string, error) {
	var version string
	body, err := getResponseBody(http.Get(client.makeEngineEndpointPath(versionEndpoint)))
	if (err != nil) {
		return version, err
	}
	version = string(body)
	return version, err
}

func (client CromwellClient) SubmitWorkflow(command SubmitCommand) (WorkflowStatus, error) {
	var status WorkflowStatus
	
	body, err := getResponseBody(client.submit(command))
	if (err != nil) {
		return status, err
	}
	err = json.Unmarshal(body, &status)
	if (err != nil) {
		return status, err
	}
	return status, err
}

func (client CromwellClient) GetWorkflowStatus(workflowId string) (WorkflowStatus, error) {
	var status WorkflowStatus
	body, err := getResponseBody(http.Get(client.makeWorkflowEndpointPath(workflowId, statusEndpoint)))
	if (err != nil) {
		return status, err
	}
	err = json.Unmarshal(body, &status)
	if (err != nil) {
		return status, err
	}
	return status, err
}

func (client CromwellClient) GetWorkflowOutputs(workflowId string) (WorkflowOutputs, error) {
	var outputs WorkflowOutputs

	// Get Body
	body, err := getResponseBody(http.Get(client.makeWorkflowEndpointPath(workflowId, outputsEndpoint)))
	if (err != nil) { return outputs, err }

	// Construct response
	outputs = WorkflowOutputs {
		JsonResponse: &JsonResponse {
			Id: workflowId,
			JsonValue: body,
		},
	}

	return outputs, err
}

func (client CromwellClient) GetWorkflowMetadata(workflowId string) (WorkflowMetadata, error) {
	var metadata WorkflowMetadata

	// Get Body
	body, err := getResponseBody(http.Get(client.makeWorkflowEndpointPath(workflowId, metadataEndpoint)))
	if (err != nil) { return metadata, err }

	// Construct response
	metadata = WorkflowMetadata {
		JsonResponse: &JsonResponse {
			Id: workflowId,
			JsonValue: body,
		},
	}

	return metadata, err
}

// Utility method to extract the body from the response
func getResponseBody(resp *http.Response, err error) ([]byte, error) {
	var body []byte
	if (err != nil) {
		return body, err
	}
	if (resp.StatusCode >= 300) {
		return body, errors.New("Request failed: " + resp.Status)
	}
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}


func (client CromwellClient) makeWorkflowEndpointPath(workflowId string, endpoint string) string {
	return fmt.Sprintf("%s%s/%s%s", client.CromwellUrl.String(), workflowEndpoint, workflowId, endpoint)
}

func (client CromwellClient) makeEngineEndpointPath(endpoint string) string {
	return fmt.Sprintf("%s%s%s", client.CromwellUrl.String(), engineEndpoint, endpoint)
}

func (client CromwellClient) submitEndpoint() string {
	return client.CromwellUrl.String() + workflowEndpoint
}

func (cromwellClient CromwellClient) submit(command SubmitCommand) (*http.Response, error) {
	wdlSource := command.WdlSource
	workflowInputs := command.WorkflowInputs
	workflowOptions := command.WorkflowOptions

	// buffer that'll contain the form content
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Open WDL file
	wdlFile, err := os.Open(wdlSource)
	if err != nil { return nil, err }
	defer wdlFile.Close()

	fw, err := w.CreateFormFile("wdlSource", wdlSource)
	if err != nil { return nil, err	}
	if _, err = io.Copy(fw, wdlFile); err != nil { return nil, err }

	if (workflowInputs != "") {
		inputsFile, err := os.Open(workflowInputs)
		if err != nil { return nil, err }
		defer inputsFile.Close()

		fw, err := w.CreateFormFile("workflowInputs", workflowInputs)
		if err != nil { return nil, err	}
		if _, err = io.Copy(fw, inputsFile); err != nil { return nil, err }
	}

	if (workflowOptions != "") {
		optionsFile, err := os.Open(workflowOptions)
		if err != nil { return nil, err }
		defer optionsFile.Close()

		fw, err := w.CreateFormFile("workflowOptions", workflowOptions)
		if err != nil { return nil, err	}
		if _, err = io.Copy(fw, optionsFile); err != nil { return nil, err }
	}

	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", cromwellClient.submitEndpoint(), &b)
	if err != nil {	return nil, err	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)

	return res, err
}
