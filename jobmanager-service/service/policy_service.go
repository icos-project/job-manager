/*
  JOB-MANAGER
  Copyright © 2022-2024 EVIDEN

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

  This work has received funding from the European Union's HORIZON research
  and innovation programme under grant agreement No. 101070177.
*/

package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/repository"
	"icos/server/jobmanager-service/utils/logs"
	"net/http"

	"moul.io/http2curl"
)

type PolicyService interface {
	HandlePolicyIncompliance(incomplianceBody []byte) (*models.Incompliance, error)
	NotifyPolicyManager(manifest string, jobGroup *models.JobGroup, token string) error
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PolicyService struct implements the PolicyService interface
type policyService struct {
	policyRepository repository.PolicyRepository
	jobRepository    repository.JobRepository
	httpClient       HTTPClient
}

// NewPolicyService returns a new instance of policyService
func NewPolicyService(policyRepository repository.PolicyRepository, jobRepository repository.JobRepository, httpClient HTTPClient) PolicyService {
	return &policyService{policyRepository: policyRepository, jobRepository: jobRepository, httpClient: httpClient}
}

// HandlePolicyIncompliance processes incompliance and applies remediation
func (s *policyService) HandlePolicyIncompliance(incomplianceBody []byte) (*models.Incompliance, error) {
	incompliance := models.Incompliance{}
	err := json.Unmarshal(incomplianceBody, &incompliance)
	if err != nil {
		return nil, err
	}

	// Save incompliance to the database
	_, err = s.policyRepository.SaveIncompliance(&incompliance)
	if err != nil {
		return nil, err
	}

	// Retrieve the job related to the incompliance
	jobGotten, err := s.jobRepository.FindJobByResourceUUID(incompliance.Subject.ResourceID)
	if err != nil {
		return nil, err
	}
	logs.Logger.Println("job found " + jobGotten.ID)

	// Validate job for remediation
	if jobGotten.OwnerID == "" {
		return nil, errors.New("OwnerID cannot be nil")
	}
	if jobGotten.State != models.JobFinished {
		return nil, errors.New("job cannot be remediated")
	}

	jobGotten.State = models.JobCreated
	jobGotten.Type = models.UpdateDeployment

	// Set job subtype based on the remediation type
	switch incompliance.Remediation {
	case models.ScaleUp:
		jobGotten.SubType = models.ScaleUp
	case models.ScaleDown:
		jobGotten.SubType = models.ScaleDown
	case models.ScaleIn:
		jobGotten.SubType = models.ScaleIn
	case models.ScaleOut:
		jobGotten.SubType = models.ScaleOut
	case models.Reallocation:
		jobGotten.SubType = models.Reallocation
	}

	// Update the job
	_, err = s.jobRepository.UpdateJob(jobGotten)
	if err != nil {
		return nil, err
	}

	return &incompliance, nil
}

func (s *policyService) NotifyPolicyManager(manifest string, jobGroup *models.JobGroup, token string) error {
	notification := models.Notification{
		AppInstance: jobGroup.ID,
		Service:     "job-manager",
		CommonAction: models.Action{
			URI:                "/jobmanager/policies/incompliance/create",
			HTTPMethod:         "POST",
			IncludeAccessToken: true,
		},
	}
	bodyBytes, err := json.Marshal(notification)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", models.PolicyManagerBaseURL+"/polman/registry/api/v1/icos/", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}
	defer resp.Body.Close()

	command, _ := http2curl.GetCurlCommand(req)
	logs.Logger.Println("Request sent to Policy Manager Service: ")
	logs.Logger.Println(command)
	logs.Logger.Println("End Policy Manager Request.")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	} else {
		err := errors.New("Bad response from Policy Manager: status code - " + string(rune(resp.StatusCode)))
		return err
	}
}
