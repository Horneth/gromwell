package gromwell

import (
	"net/url"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

type CromwellClient struct {
	CromwellUrl *url.URL
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

func (jsonResponse *JsonResponse) ToFile(filePath string) error {
	return ioutil.WriteFile(filePath, jsonResponse.JsonValue, 0644)
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
