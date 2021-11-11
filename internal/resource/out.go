package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tbe/resource-framework/log"
)

type Body struct {
	Inventory int `json:"inventory,omitempty"`
	ExtraVars string `json:"extra_vars,omitempty"`
}

type LaunchResponse struct {
	Job int `json:"job"`
}

type InvResult struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type InvResponse struct {
	Count   int         `json:"count"`
	Results []InvResult `json:"results"`
}

type JobResponse struct {
	Finished *string `json:"finished"`
	Failed   bool    `json:"failed"`
}
type AWXRange struct {
	Start       int `json:"start"`
	End         int `json:"end"`
	AbsoluteEnd int `json:"absolute_end"`
}

type StdoutResponse struct {
	Range   AWXRange `json:"range"`
	Content string   `json:"content"`
}

func (a *AWXResource) Out(_ string) (version interface{}, metadata []interface{}, err error) {

	logger := log.NewDefaultLogger()
	logger.SetLevel(0)
	// get inventory id
	var jsonStr []byte
	bdy := Body{}
	if a.params.Inventory != "" {
		inventoryId, err := a.getInventoryId(a.params.Inventory)
		if err != nil {
			logger.Warn("error getting inventory id: %v", err)
			return nil, nil, err
		}
		bdy.Inventory = inventoryId
	}
	if a.params.ExtraVars != "" {
		bdy.ExtraVars = a.params.ExtraVars
	}

	jsonStr, err = json.Marshal(bdy)
	if err != nil {
		logger.Warn("error marshalling json body: %v", err)
		return nil, nil, err
	}
	// launch job from template
	req, err := http.NewRequest("POST", a.source.Endpoint+"/api/v2/job_templates/"+strconv.Itoa(a.params.TemplateId)+"/launch/", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.source.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logger.Warn("error requesting job template: %v", err)
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode != 201 {
		logger.Warn("Unexpected Status Code: %d\n", res.StatusCode)
		logger.Warn("Body: \n%v\n", string(body))
		err = fmt.Errorf("unexpected http status code: %d", res.StatusCode)
		return nil, nil, err
	}
	obj := LaunchResponse{}
	err = json.Unmarshal(body, &obj)
	if err != nil {
		return nil, nil, err
	}
	logger.Info("Job Id: %d\n", obj.Job)

	// wait for the job to complete and stream output
	finished := false
	for !finished {
		job, err := a.getJobStatus(obj.Job)
		if err != nil {
			logger.Warn("error getting job status: %v", err)
			return nil, nil, err
		}
		if job.Finished != nil {
			finished = true
			if job.Failed {
				logger.Warn("job failed\n")
				content, err := a.getJobStdout(obj.Job)
				if err != nil {
					logger.Warn("error getting job stdout: %v\n", err)
					return nil, nil, err
				}
				_, _ = fmt.Fprintf(os.Stderr, "%s", content)
				os.Exit(1)
			}
		}
		time.Sleep(10 * time.Second)
	}

	// the job is finished get the output
	content, err := a.getJobStdout(obj.Job)
	if err != nil {
		logger.Warn("error getting job stdout: %v", err)
		return nil, nil, err
	}

	// did not use logger here due to colors
	_, _ = fmt.Fprintf(os.Stderr, "%s", content)

	// return nothing else
	return AWXVersion{}, []interface{}{&AWXMetadata{}}, nil
}

func (a *AWXResource) getInventoryId(inventory string) (int, error) {
	logger := log.NewDefaultLogger()
	logger.SetLevel(0)

	invReq, err := http.NewRequest("GET", a.source.Endpoint+"/api/v2/inventories/?search="+inventory, nil)
	if err != nil {
		logger.Warn("error creating inventory request: %v\n", err)
		return 0, err
	}
	invReq.Header.Set("Authorization", "Bearer "+a.source.AuthToken)
	invReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	invResponse, err := client.Do(invReq)
	if err != nil {
		logger.Warn("error requesting inventory: %v\n", err)
		return 0, err
	}
	body, err := ioutil.ReadAll(invResponse.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(invResponse.Body)
	if err != nil {
		logger.Warn("error reading inventory response: %v\n", err)
		return 0, err
	}

	invR := InvResponse{}
	err = json.Unmarshal(body, &invR)
	if err != nil {
		logger.Warn("error unmarshalling inventory response: %v\n", err)
		return 0, err
	}
	if invR.Count != 1 {
		logger.Warn("found wrong number of inventories: %d\n", invR.Count)
		return 0, fmt.Errorf("wrong number of inventories: %d", invR.Count)
	}

	return invR.Results[0].Id, nil
}

func (a *AWXResource) getJobStatus(jobId int) (JobResponse, error) {
	logger := log.NewDefaultLogger()
	logger.SetLevel(0)

	jobReq, err := http.NewRequest("GET", a.source.Endpoint+"/api/v2/jobs/"+strconv.Itoa(jobId)+"/", nil)
	if err != nil {
		logger.Warn("Error creating job request: %v", err)
		return JobResponse{}, err
	}
	jobReq.Header.Set("Authorization", "Bearer "+a.source.AuthToken)
	jobReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	jobRes, err := client.Do(jobReq)
	if err != nil {
		logger.Warn("error requesting job status: %v", err)
		return JobResponse{}, err
	}
	body, err := ioutil.ReadAll(jobRes.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(jobRes.Body)
	if err != nil {
		logger.Warn("error reading job response: %v", err)
		return JobResponse{}, err
	}
	job := JobResponse{}
	err = json.Unmarshal(body, &job)
	if err != nil {
		logger.Warn("error unmarshalling job response: %v", err)
		return JobResponse{}, err
	}
	return job, nil
}

func (a *AWXResource) getJobStdout(jobId int) (string, error) {
	logger := log.NewDefaultLogger()
	logger.SetLevel(0)

	stdoutReq, err := http.NewRequest("GET", a.source.Endpoint+"/api/v2/jobs/"+strconv.Itoa(jobId)+"/stdout/?format=json", nil)
	if err != nil {
		logger.Warn("error creating stdout request: %v", err)
		return "", err
	}
	stdoutReq.Header.Set("Authorization", "Bearer "+a.source.AuthToken)
	stdoutReq.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	stdoutRes, err := client.Do(stdoutReq)
	if err != nil {
		logger.Warn("error getting stdout: %v", err)
		return "", err
	}
	body, err := ioutil.ReadAll(stdoutRes.Body)
	if err != nil {
		logger.Warn("error reading stdout response: %v", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(stdoutRes.Body)

	stdout := StdoutResponse{}
	err = json.Unmarshal(body, &stdout)
	if err != nil {
		logger.Warn("error unmarshalling stdout response: %v", err)
		return "", err
	}
	return stdout.Content, nil
}
