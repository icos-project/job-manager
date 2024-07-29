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
	"fmt"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/repository"
	"icos/server/jobmanager-service/utils/logs"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// JobGroupService interface defines the methods for job group operations
type JobGroupService interface {
	CreateJobGroup(bodyBytes []byte, header http.Header) (*models.JobGroup, error)
	UpdateJobGroup(bodyJob []byte) (*models.JobGroup, error)
	FindJobGroupByUUID(string) (*models.JobGroup, error)
	FindAllJobGroups() (*[]models.JobGroup, error)
	DeleteJobGroupByID(id string) (*models.JobGroup, error)
	StopJobGroupByID(stringID string) (*models.JobGroup, error)
}

// jobGroupService struct implements the JobGroupService interface
type jobGroupService struct {
	repo repository.JobGroupRepository
}

// NewJobGroupService returns a new instance of jobGroupService
func NewJobGroupService(repo repository.JobGroupRepository) JobGroupService {
	return &jobGroupService{repo: repo}
}

// SaveJobGroup saves a new job group
func (s *jobGroupService) CreateJobGroup(bodyBytes []byte, header http.Header) (*models.JobGroup, error) {
	bodyString := string(bodyBytes)
	bodyStringTrimmed := strings.Trim(bodyString, "\r\n")
	logs.Logger.Println("Trimmed body: " + bodyStringTrimmed)

	applicationDescriptor := models.JobGroupHeader{}
	err := yaml.Unmarshal([]byte(bodyStringTrimmed), &applicationDescriptor)
	if err != nil {
		return nil, err
	}

	logs.Logger.Printf("Application descriptor: %#v", applicationDescriptor)
	// // Prepare request for MM
	req, err := http.NewRequest("POST", models.MatchmakerBaseURL+"/matchmake", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return nil, err
	}

	// add content type
	req.Header.Set("Content-Type", "application/x-yaml")
	// forward the authorization token
	req.Header.Add("Authorization", header.Get("Authorization"))

	// do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Logger.Printf("ERROR executing request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	//Mocking MM response for development
	// mMResponseJson := MockMatchmakerResponse()
	// bodyMM := []byte(mMResponseJson)
	//end of mock

	bodyMM, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Logger.Printf("ERROR reading response body: %v", err)
		return nil, err
	}

	// Log and format MM response
	dst := &bytes.Buffer{}
	if err := json.Indent(dst, bodyMM, "", "  "); err != nil {
		logs.Logger.Printf("ERROR formatting JSON response: %v", err)
		return nil, err
	}
	logs.Logger.Println("MM response is: " + dst.String())

	_ = json.Indent(dst, bodyMM, "", "  ")
	logs.Logger.Println("MM response is: " + dst.String())
	err = json.Unmarshal(bodyMM, &applicationDescriptor)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return nil, err
	}
	logs.Logger.Printf("Matchmaking response details: %#v", applicationDescriptor)

	conditions := []models.Condition{
		{
			Type:               "Created",
			Status:             "True",
			ObservedGeneration: 1,
			LastTransitionTime: time.Now(),
			Reason:             "AwaitingForTarget",
			Message:            "Waiting for the Target",
		},
		{
			Type:               "Created",
			Status:             "True",
			ObservedGeneration: 1,
			LastTransitionTime: time.Now(),
			Reason:             "AwaitingForExecution",
			Message:            "Waiting an Orchestrator to take the Job",
		},
	}

	jobGroup := models.JobGroup{
		AppName:        applicationDescriptor.Name,
		AppDescription: applicationDescriptor.Description,
	}

	if jobGroup.AppName == "" {
		jobGroup.AppName = uuid.New().String()
	}

	for _, comp := range applicationDescriptor.Components {
		job := models.Job{
			Type:  models.CreateDeployment,
			State: models.JobCreated,
			//Targets:      comp.Targets,
			JobGroupName: jobGroup.AppName,
			//Orchestrator: comp.Targets.Orchestrator,
			Namespace: applicationDescriptor.Name,
			Resource: &models.Resource{
				ResourceName: comp.Name,
				Conditions:   conditions,
			},
		}

		/* Given that it is possible that MM returns an empty target, we need to handle this case.
		The switch statement below checks the type of `targets` and handles each case appropriately:
		- If `targets` is a `map[string]interface{}`, it represents a single target object.
		  We marshal this map to JSON and then unmarshal it into a `Target` struct.
		- If `targets` is an empty array (`[]interface{}` with length 0), we assign an empty
		  `Target` struct to indicate that there are no targets.
		- If `targets` is any other type, it's an unexpected scenario and we log an error.
		*/

		switch targets := comp.Targets.(type) {
		case map[string]interface{}:
			targetBytes, err := json.Marshal(targets)
			if err != nil {
				logs.Logger.Println("ERROR " + err.Error())
				return nil, err
			}

			var targetStruct models.Target
			err = json.Unmarshal(targetBytes, &targetStruct)
			if err != nil {
				logs.Logger.Println("ERROR " + err.Error())
				return nil, err
			}

			job.Targets = targetStruct
			job.Orchestrator = targetStruct.Orchestrator

		case []interface{}:
			if len(targets) == 0 {
				job.Targets = models.Target{} // Empty Target
			} else {
				logs.Logger.Fatalf("Unexpected non-empty array for targets")
			}

		default:
			logs.Logger.Fatalf("Unexpected type for targets: %T", comp.Targets)
		}

		for _, manifest := range applicationDescriptor.Manifests {
			manifestMap, ok := manifest["metadata"].(map[interface{}]interface{})
			if !ok {
				logs.Logger.Println("ERROR: Invalid manifest structure")
				continue
			}

			manifestName, ok := manifestMap["name"].(string)
			if !ok {
				logs.Logger.Println("ERROR: Invalid manifest name")
				continue
			}

			for _, manifestRef := range comp.Manifests {
				logs.Logger.Printf("ManifestRef name: %s", manifestRef.Name)
				logs.Logger.Print("Manifest name: " + manifestName)
				if manifestRef.Name == manifestName {
					logs.Logger.Println("Marshalling manifest to YAML for manifestRef: " + manifestRef.Name)
					manifestYAML, err := yaml.Marshal(manifest)
					if err != nil {
						logs.Logger.Println("ERROR during k8s manifest marshalling: " + err.Error())
						continue
					}

					logs.Logger.Println("YAML Marshalled: " + string(manifestYAML))
					_, err = decodeYAMLToObject(string(manifestYAML))
					if err != nil {
						logs.Logger.Println("ERROR during k8s manifest validation: " + err.Error())
						continue
					}

					logs.Logger.Println("Successfully decoded YAML to k8s object for manifestRef: " + manifestRef.Name)
					logs.Logger.Println("Manifest to be populated: " + manifestRef.Name)
					job.Manifests = append(job.Manifests, models.PlainManifest{
						YamlString: string(manifestYAML),
					})
				}
			}

		}

		jobGroup.Jobs = append(jobGroup.Jobs, job)
		logs.Logger.Println("New Job appended to JobGroup: " + job.JobGroupID)
	}

	_, err = s.repo.SaveJobGroup(&jobGroup)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return nil, err
	}

	return &jobGroup, nil
}

func MockMatchmakerResponse() string {
	mMResponseJson := `
	{
		"components": [
			{
				"name": "producer",
				"type": "manifest",
				"manifests": [
					{
						"name": "producer"
					},
					{
						"name": "producer-service"
					}
				],
				"targets": {
					"cluster_name": "raspis",
					"node_name": "k3s-master",
					"orchestrator": "ocm"
				}
			},
			{
				"name": "consumer",
				"type": "manifest",
				"manifests": [
					{
						"name": "consumer-statefulset"
					},
					{
						"name": "consumer-service"
					}
				],
				"targets": {
					"cluster_name": "raspis",
					"node_name": "k3s-master",
					"orchestrator": "ocm"
				}
			},
			{
				"name": "player",
				"type": "manifest",
				"manifests": [
					{
						"name": "player"
					},
					{
						"name": "player-service"
					}
				],
				"targets": {
					"cluster_name": "raspis",
					"node_name": "k3s-master",
					"orchestrator": "ocm"
				}
			}
		]
	}`
	return mMResponseJson
}

func decodeYAMLToObject(yamlString string) (runtime.Object, error) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add apps/v1 to scheme: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add core/v1 to scheme: %w", err)
	}

	codecFactory := serializer.NewCodecFactory(scheme)
	decoder := codecFactory.UniversalDeserializer()

	obj, _, err := decoder.Decode([]byte(yamlString), nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// UpdateJobGroup updates an existing job group
func (s *jobGroupService) UpdateJobGroup(bodyJob []byte) (*models.JobGroup, error) {
	var jobGroupUpdate models.JobGroup
	if err := json.Unmarshal(bodyJob, &jobGroupUpdate); err != nil {
		logs.Logger.Println("Error unmarshaling request body:", err)
		return nil, err
	}

	logs.Logger.Println("Updating job group with ID:", jobGroupUpdate.ID)
	existingJobGroup, err := s.FindJobGroupByUUID(jobGroupUpdate.ID)
	if err != nil {
		logs.Logger.Println("Error finding job group by UUID:", err)
		return nil, err
	}

	existingJobGroup.AppName = jobGroupUpdate.AppName
	existingJobGroup.AppDescription = jobGroupUpdate.AppDescription

	if len(jobGroupUpdate.Jobs) > 0 {
		jobMap := make(map[string]models.Job)
		for _, updatedJob := range jobGroupUpdate.Jobs {
			jobMap[updatedJob.ID] = updatedJob
		}

		for i := range existingJobGroup.Jobs {
			if updatedJob, ok := jobMap[existingJobGroup.Jobs[i].ID]; ok {
				existingJobGroup.Jobs[i] = updatedJob
			}
		}
	}

	for i := range existingJobGroup.Jobs {
		job := &existingJobGroup.Jobs[i]
		job.State = models.JobCreated
		if job.OwnerID != "" {
			job.Type = models.ReplaceDeployment
		} else {
			job.Type = models.CreateDeployment
		}
	}

	jobGroupUpdated, err := s.repo.UpdateJobGroup(existingJobGroup)
	if err != nil {
		logs.Logger.Println("Error updating job group:", err)
		return nil, err
	}

	return jobGroupUpdated, nil
}

// DeleteJobGroup deletes a job group
func (s *jobGroupService) DeleteJobGroupByID(id string) (*models.JobGroup, error) {
	if id == "" {
		err := errors.New("ID Cannot be empty")
		logs.Logger.Println("JobGroup's ID is empty!")
		return nil, err
	}

	// Validate job group can be deleted
	jobGroupGotten, err := s.repo.FindJobGroupByUUID(id)
	if err != nil {
		return nil, err
	}

	for _, job := range jobGroupGotten.Jobs {
		logs.Logger.Printf("Checking job with ID: %s, Type: %d, State: %d\n", job.ID, job.Type, job.State)
		// Job can only be deleted if it is of type DeleteDeployment
		if job.Type == models.DeleteDeployment {
			// Job has resources and it is not in JobFinished or JobCreated state
			if job.State != models.JobFinished && job.State != models.JobCreated {
				err := errors.New("JobGroup cannot be deleted, one or more jobs are not in JobFinished state")
				logs.Logger.Println("JobGroup cannot be deleted, one or more jobs are not in JobFinished state")
				return nil, err
			}
		} else {
			err := errors.New("JobGroup cannot be deleted, one or more jobs are not of type DeleteDeployment")
			logs.Logger.Println("JobGroup cannot be deleted, one or more jobs are not of type DeleteDeployment")
			return nil, err
		}
	}

	// Delete the job group
	_, err = s.repo.DeleteJobGroup(id)
	if err != nil {
		return nil, err
	}

	return jobGroupGotten, nil
}

func (s *jobGroupService) StopJobGroupByID(stringID string) (*models.JobGroup, error) {
	if stringID == "" {
		return nil, errors.New("ID Cannot be empty")
	}

	jobGroupGotten, err := s.FindJobGroupByUUID(stringID)
	if err != nil {
		return nil, errors.New("JobGroup not found")
	}

	for i := range jobGroupGotten.Jobs {
		// TODO: Add comment explaining new executable jobs
		job := &jobGroupGotten.Jobs[i]
		switch job.State {
		case models.JobCreated:
			job.State = models.JobFinished
			job.OwnerID = ""
		default:
			job.State = models.JobCreated
			job.Type = models.DeleteDeployment
		}
	}

	updatedJobGroup, err := s.repo.UpdateJobGroup(jobGroupGotten)
	if err != nil {
		return nil, errors.New("error updating JobGroup")
	}

	return updatedJobGroup, nil
}

// FindJobGroupByUUID finds a job group by its UUID
func (s *jobGroupService) FindJobGroupByUUID(id string) (*models.JobGroup, error) {
	return s.repo.FindJobGroupByUUID(id)
}

// FindAllJobGroups finds all job groups
func (s *jobGroupService) FindAllJobGroups() (*[]models.JobGroup, error) {
	// Implementation for finding all job groups
	return s.repo.FindAllJobGroups()
}
