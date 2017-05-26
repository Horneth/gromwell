package cromwell_api

import (
	"net/http"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
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
	body, err := getResponseBody(http.Get(client.makeWorkflowEndpointPath(workflowId, outputsEndpoint)))
	if (err != nil) {
		return outputs, err
	}
	err = json.Unmarshal(body, &outputs)
	if (err != nil) {
		return outputs, err
	}
	return outputs, err
}

func (client CromwellClient) GetWorkflowMetadata(workflowId string) (WorkflowMetadata, error) {
	var metadata WorkflowMetadata
	var rawMetadata map[string]interface{}
	body, err := getResponseBody(http.Get(client.makeWorkflowEndpointPath(workflowId, metadataEndpoint)))
	if (err != nil) {
		return metadata, err
	}
	err = json.Unmarshal(body, &rawMetadata)
	if (err != nil) {
		return metadata, err
	}
	metadata = WorkflowMetadata {
		Id: rawMetadata["id"].(string),
		Metadata: body,
	}
	return metadata, err
}

// Utility method to extract the body from the response
func getResponseBody(resp *http.Response, err error) ([]byte, error) {
	var body []byte
	if (err != nil) {
		return body, err
	}
	if (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated) {
		return body, errors.New("Status query failed: " + resp.Status)
	}
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}