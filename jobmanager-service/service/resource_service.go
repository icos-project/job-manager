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
	"encoding/json"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/repository"
	"icos/server/jobmanager-service/utils/logs"
)

type ResourceService interface {
	SaveResource(*models.Resource) (*models.Resource, error)
	UpdateAResource(*models.Resource) (*models.Resource, error)
	AddCondition(*models.Resource, *models.Condition) (*models.Resource, error)
	RemoveConditions(*models.Resource) (*models.Resource, error)
	FindResourceByJobUUID(string) (*models.Resource, error)
	UpdateResourceState([]byte) (*models.Resource, error)
}

// ResourceService struct implements the ResourceService interface
type resourceService struct {
	resourceRepository repository.ResourceRepository
	jobRepository      repository.JobRepository
}

// NewResourceService returns a new instance of resourceService
func NewResourceService(resourceRepository repository.ResourceRepository, jobRepository repository.JobRepository) ResourceService {
	return &resourceService{resourceRepository: resourceRepository,
		jobRepository: jobRepository}
}

// SaveResource saves a new resource
func (s *resourceService) SaveResource(r *models.Resource) (*models.Resource, error) {
	return s.resourceRepository.SaveResource(r)
}

// UpdateAResource updates an existing resource
func (s *resourceService) UpdateAResource(r *models.Resource) (*models.Resource, error) {
	return s.resourceRepository.UpdateAResource(r)
}

// AddCondition adds a condition to a resource
func (s *resourceService) AddCondition(r *models.Resource, c *models.Condition) (*models.Resource, error) {
	return s.resourceRepository.AddCondition(r, c)
}

// RemoveConditions removes all conditions from a resource
func (s *resourceService) RemoveConditions(r *models.Resource) (*models.Resource, error) {
	return s.resourceRepository.RemoveConditions(r)
}

// FindResourceByJobUUID finds a resource by its job UUID
func (s *resourceService) FindResourceByJobUUID(jobId string) (*models.Resource, error) {
	return s.resourceRepository.FindResourceByJobUUID(jobId)
}

// UpdateResourceState updates the state of a resource
func (s *resourceService) UpdateResourceState(resourceBody []byte) (*models.Resource, error) {
	resource := models.Resource{}

	// Parse to application objects
	err := json.Unmarshal(resourceBody, &resource)
	if err != nil {
		return nil, err
	}

	// Get resource from db, retrieve the job first
	jobGotten, err := s.jobRepository.FindJobByResourceUUID(resource.ResourceUID)
	if err != nil {
		return nil, err
	}

	j, err := json.Marshal(jobGotten)
	if err != nil {
		logs.Logger.Println("ERROR during debug" + err.Error())
	}
	logs.Logger.Println("Job contents: " + string(j))

	// Update resource details
	resource.ResourceUID = jobGotten.ID
	resource.ID = jobGotten.Resource.ID
	resource.JobID = jobGotten.ID
	logs.Logger.Println("Updating Resource Status, Resource ID: " + resource.ID)

	_, err = s.resourceRepository.UpdateAResource(&resource)
	if err != nil {
		return nil, err
	}
	s.resourceRepository.RemoveConditions(&resource)
	for _, condition := range resource.Conditions {
		condition.ResourceID = resource.ID
		_, err = s.resourceRepository.AddCondition(&resource, &condition)
		if err != nil {
			return nil, err
		}
	}
	return &resource, nil
}
