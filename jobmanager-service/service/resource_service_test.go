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

package service_test

import (
	"encoding/json"
	"testing"

	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/service"
	repository "icos/server/jobmanager-service/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResourceService(t *testing.T) {
	mockResourceRepo := new(repository.MockResourceRepository)
	mockJobRepo := new(repository.MockJobRepository)
	resourceService := service.NewResourceService(mockResourceRepo, mockJobRepo)

	resource := &models.Resource{}
	condition := &models.Condition{Type: "Ready"}
	job := &models.Job{
		BaseUUID: models.BaseUUID{
			ID: "54b68f2f-72c4-4df8-8b9d-f9ebc31bdf7f",
		},
		Resource: &models.Resource{
			BaseUUID: models.BaseUUID{
				ID: "91114c14-3ae0-442b-835b-a4f5e24c99c9",
			},
		},
	}

	t.Run("SaveResource", func(t *testing.T) {
		mockResourceRepo.On("SaveResource", resource).Return(resource, nil)
		result, err := resourceService.SaveResource(resource)
		assert.NoError(t, err)
		assert.Equal(t, resource, result)
		mockResourceRepo.AssertExpectations(t)
	})

	t.Run("UpdateAResource", func(t *testing.T) {
		mockResourceRepo.On("UpdateAResource", resource).Return(resource, nil)
		result, err := resourceService.UpdateAResource(resource)
		assert.NoError(t, err)
		assert.Equal(t, resource, result)
		mockResourceRepo.AssertExpectations(t)
	})

	t.Run("AddCondition", func(t *testing.T) {
		mockResourceRepo.On("AddCondition", resource, condition).Return(resource, nil)
		result, err := resourceService.AddCondition(resource, condition)
		assert.NoError(t, err)
		assert.Equal(t, resource, result)
		mockResourceRepo.AssertExpectations(t)
	})

	t.Run("RemoveConditions", func(t *testing.T) {
		mockResourceRepo.On("RemoveConditions", resource).Return(resource, nil)
		result, err := resourceService.RemoveConditions(resource)
		assert.NoError(t, err)
		assert.Equal(t, resource, result)
		mockResourceRepo.AssertExpectations(t)
	})

	t.Run("FindResourceByJobUUID", func(t *testing.T) {
		mockResourceRepo.On("FindResourceByJobUUID", "54b68f2f-72c4-4df8-8b9d-f9ebc31bdf7f").Return(resource, nil)
		result, err := resourceService.FindResourceByJobUUID("54b68f2f-72c4-4df8-8b9d-f9ebc31bdf7f")
		assert.NoError(t, err)
		assert.Equal(t, resource, result)
		mockResourceRepo.AssertExpectations(t)
	})

	t.Run("UpdateResourceState", func(t *testing.T) {
		resourceBody := []byte(`{
			"ResourceUID": "91114c14-3ae0-442b-835b-a4f5e24c99c9",
			"Conditions": [
				{
					"Type": "Ready",
					"ResourceID": ""
				}
			]
		}`)
		resource := models.Resource{}
		err := json.Unmarshal(resourceBody, &resource)
		assert.NoError(t, err)

		mockJobRepo.On("FindJobByResourceUUID", resource.ResourceUID).Return(job, nil)
		mockResourceRepo.On("UpdateAResource", mock.Anything).Return(&resource, nil)
		mockResourceRepo.On("RemoveConditions", mock.Anything).Return(&resource, nil)
		mockResourceRepo.On("AddCondition", mock.Anything, mock.Anything).Return(&resource, nil)

		result, err := resourceService.UpdateResourceState(resourceBody)
		assert.NoError(t, err)
		assert.Equal(t, job.ID, result.ResourceUID)
		assert.Equal(t, job.Resource.ID, result.ID)
		assert.Equal(t, job.ID, result.JobID)
		mockJobRepo.AssertExpectations(t)
		mockResourceRepo.AssertExpectations(t)
	})
}
