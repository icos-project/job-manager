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

package repository

import (
	"testing"

	"icos/server/jobmanager-service/models"
	mocks "icos/server/jobmanager-service/repository/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func initJobRepo(db *gorm.DB) interface{} {
	return NewJobRepository(db)
}

func TestSaveJob(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job := &models.Job{State: 1}

	result, err := repo.SaveJob(job)
	assert.NoError(t, err)
	assert.Equal(t, job, result)
}

func TestUpdateJob(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job := &models.Job{State: 1}
	repo.SaveJob(job)

	job.State = 2
	result, err := repo.UpdateJob(job)
	assert.NoError(t, err)
	assert.Equal(t, job, result)
}

func TestDeleteJob(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job := &models.Job{State: 1}
	repo.SaveJob(job)

	rowsAffected, err := repo.DeleteJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
}

func TestFindJobByUUID(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job := &models.Job{State: 1}
	repo.SaveJob(job)

	result, err := repo.FindJobByUUID(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, job, result)
}

// func TestFindJobByResourceUUID(t *testing.T) {
// 	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

// 	job := &models.Job{BaseUUID: models.BaseUUID{
// 		ID: uuid.New().String(),
// 	}, State: 1}
// 	resource := &models.Resource{JobID: job.ID, ResourceUID: "test-resource-uuid"}
// 	repo.SaveJob(job)

// 	result, err := repo.FindJobByResourceUUID(resource.ResourceUID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, job, result)
// }

func TestFindAllJobs(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job1 := &models.Job{State: 1}
	job2 := &models.Job{State: 2}
	repo.SaveJob(job1)
	repo.SaveJob(job2)

	result, err := repo.FindAllJobs()
	assert.NoError(t, err)
	assert.Len(t, *result, 2)
}

func TestFindJobsByState(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job1 := &models.Job{State: 1}
	job2 := &models.Job{State: 2}
	repo.SaveJob(job1)
	repo.SaveJob(job2)

	result, err := repo.FindJobsByState(1)
	assert.NoError(t, err)
	assert.Len(t, *result, 1)
	assert.Equal(t, job1.ID, (*result)[0].ID)
}

func TestFindJobsToExecute(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job1 := &models.Job{Type: models.CreateDeployment, State: models.JobCreated, Orchestrator: "ocm"}
	job2 := &models.Job{Type: models.CreateDeployment, State: models.JobCreated, Orchestrator: "ocm"}
	repo.SaveJob(job1)
	repo.SaveJob(job2)

	result, err := repo.FindJobsToExecute("ocm", "")
	assert.NoError(t, err)
	assert.Len(t, *result, 2)
}

func TestJobPromote(t *testing.T) {
	repo := mocks.SetupTest(t, initJobRepo).(JobRepository)

	job := MockJob()
	repo.SaveJob(&job)

	job.State = 2
	result, err := repo.JobPromote(&job)
	assert.NoError(t, err)
	assert.Equal(t, &job, result)
}

// TODO: add better mocks
func MockJob() models.Job {
	return models.Job{
		BaseUUID: models.BaseUUID{
			ID: uuid.New().String(),
		},
		JobGroupID:          uuid.New().String(),
		OwnerID:             uuid.New().String(),
		JobGroupName:        "Mock Job Group Name",
		JobGroupDescription: "Mock Job Group Description",
		Type:                1, // CreateDeployment
		SubType:             "",
		State:               1, // JobCreated
		Manifests:           []models.PlainManifest{},
		Targets:             models.Target{},
		Orchestrator:        models.OrchestratorType("Mock Orchestrator"),
		Resource:            &models.Resource{},
		Namespace:           "Mock Namespace",
	}
}
