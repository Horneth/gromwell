package cromwell_api

import (
	"net/url"
	"fmt"
	"os"
	"io"
	"net/http"
	"bytes"
	"mime/multipart"
	"encoding/json"
)

type CromwellClient struct {
	CromwellUrl *url.URL
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

		fw, err := w.CreateFormFile("workflowInputs", workflowOptions)
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

// Requests

type SubmitCommand struct {
	WdlSource       string
	WorkflowInputs  string
	WorkflowOptions string
}

// Responses

type WorkflowStatus struct {
	Id string
	Status string
}

type WorkflowOutputs struct {
	Id string
	Outputs map[string]interface{}
}

type WorkflowMetadata struct {
	Id string
	Metadata []byte
}

func (outputs WorkflowOutputs) String() string {
	prettyJson, err := prettyPrint(outputs.Outputs)
	if (err != nil) { return fmt.Sprintf("%s", err) }
	return fmt.Sprintf("outputs for %s:\n %s\n", outputs.Id, prettyJson)
}

func (metadata WorkflowMetadata) String() string {
	var rawMetadata map[string]interface{}
	err := json.Unmarshal(metadata.Metadata, &rawMetadata)
	if (err != nil) { panic(err) }
	prettyJson, err := prettyPrint(rawMetadata)
	if (err != nil) { return fmt.Sprintf("%s", err) }
	return fmt.Sprintf("metadata for %s:\n %s\n", metadata.Id, prettyJson)
}

func (status WorkflowStatus) String() string {
	return fmt.Sprintf("%s: %s\n", status.Id, status.Status)
}

func prettyPrint(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	return string(b), err
}