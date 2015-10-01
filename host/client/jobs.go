package client

import (
	"encoding/json"
	"fmt"

	"github.com/pebblescape/pebblescape/host/types"
)

func ListJobs() ([]host.Job, error) {
	resp, body, errs := get("/job").EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	if resp.StatusCode != 200 {
		return nil, parseError(resp, body)
	}

	var jobs []host.Job
	err := json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func GetJob(id string) (*host.Job, error) {
	resp, body, errs := get("/job/" + id).EndBytes()
	if errs != nil {
		return nil, errs[0]
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Job not found")
	}
	if resp.StatusCode != 200 {
		return nil, parseError(resp, body)
	}

	var job host.Job
	err := json.Unmarshal(body, &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}
