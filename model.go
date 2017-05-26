package gromwell

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

type JsonResponse struct {
	Id string
	JsonValue []byte
}

type WorkflowOutputs struct {
	*JsonResponse
}

type WorkflowMetadata struct {
	*JsonResponse
}

func (outputs WorkflowOutputs) String() string {
	outputsMetadata, err := outputs.toJson()
	if (err != nil) { return err.Error() }
	return fmt.Sprintf("outputs for %s:\n %s\n", outputs.Id, prettyPrint(outputsMetadata))
}

func (metadata WorkflowMetadata) String() string {
	jsonMetadata, err := metadata.toJson()
	if (err != nil) { return err.Error() }
	return fmt.Sprintf("metadata for %s:\n %s\n", metadata.Id, prettyPrint(jsonMetadata))
}

func (status WorkflowStatus) String() string {
	return fmt.Sprintf("%s: %s\n", status.Id, status.Status)
}

func (jsonResponse *JsonResponse) toJson() (map[string]interface{}, error) {
	return parseJson(jsonResponse.JsonValue)
}

func prettyPrint(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if (err != nil) { return fmt.Sprintf("%s", err) }
	return string(b)
}

func parseJson(bytes []byte) (map[string]interface{}, error) {
	var jsonObject map[string]interface{}
	err := json.Unmarshal(bytes, &jsonObject)
	return jsonObject, err
}
