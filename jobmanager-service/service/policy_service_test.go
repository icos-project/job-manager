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
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/service"
	repository "icos/server/jobmanager-service/service/mocks"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockHTTPClient is a mock of the HTTPClient
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestPolicyService(t *testing.T) {
	mockPolicyRepo := new(repository.MockPolicyRepository)
	mockJobRepo := new(repository.MockJobRepository)
	mockHTTPClient := new(MockHTTPClient)
	policyService := service.NewPolicyService(mockPolicyRepo, mockJobRepo, mockHTTPClient)

	t.Run("HandlePolicyIncompliance", func(t *testing.T) {
		incomplianceBody := []byte(`{
			"currentValue": "235.365",
			"threshold": "warning",
			"policyName": "compss-low-performance-20",
			"policyId": "fd745e25-0ca3-418b-b310-257693e3f3bf",
			"measurementBackend": "prom-1",
			"extraLabels": {},
			"subject": {
				"type": "app",
				"appName": "",
				"appComponent": "producer",
				"appInstance": "27a69131-f34d-44b3-9063-81501a1c0fc8",
				"resourceId": "19d0baf9-0b90-4c36-8e41-5b18e8acbeec"
			},
			"remediation": "reallocation"
		}`)
		incompliance := models.Incompliance{}
		err := json.Unmarshal(incomplianceBody, &incompliance)
		require.NoError(t, err)

		job := &models.Job{
			BaseUUID: models.BaseUUID{
				ID: "6616b77c-dbb0-47aa-bc9b-ff45548db029",
			},
			JobGroupID: "27a69131-f34d-44b3-9063-81501a1c0fc8",
			OwnerID:    "owner-123",
			State:      models.JobFinished,
			Type:       models.UpdateDeployment,
			Resource: &models.Resource{
				BaseUUID: models.BaseUUID{
					ID: "19d0baf9-0b90-4c36-8e41-5b18e8acbeec",
				},
				ResourceName: "producer",
			},
		}

		expectedUpdatedJob := &models.Job{
			BaseUUID: models.BaseUUID{
				ID: "6616b77c-dbb0-47aa-bc9b-ff45548db029",
			},
			JobGroupID: "27a69131-f34d-44b3-9063-81501a1c0fc8",
			OwnerID:    "owner-123",
			State:      models.JobCreated,
			SubType:    models.Reallocation,
			Type:       models.UpdateDeployment,
			Resource: &models.Resource{
				BaseUUID: models.BaseUUID{
					ID: "19d0baf9-0b90-4c36-8e41-5b18e8acbeec",
				},
				ResourceName: "producer",
			},
		}

		mockPolicyRepo.On("SaveIncompliance", &incompliance).Return(&incompliance, nil)
		mockJobRepo.On("FindJobByResourceUUID", incompliance.Subject.ResourceID).Return(job, nil)
		mockJobRepo.On("UpdateJob", mock.MatchedBy(func(j *models.Job) bool {
			return j.State == expectedUpdatedJob.State && j.SubType == expectedUpdatedJob.SubType && j.Type == expectedUpdatedJob.Type
		})).Return(expectedUpdatedJob, nil)

		result, err := policyService.HandlePolicyIncompliance(incomplianceBody)
		assert.NoError(t, err)
		assert.Equal(t, &incompliance, result)
		mockPolicyRepo.AssertExpectations(t)
		mockJobRepo.AssertExpectations(t)
	})
}
