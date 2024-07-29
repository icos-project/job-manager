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
	"errors"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/repository"
	"icos/server/jobmanager-service/utils/logs"
	"net/http"

	"github.com/gorilla/mux"
)

type JobService interface {
	SaveJob(*models.Job) (*models.Job, error)
	UpdateJob(*models.Job) (*models.Job, error)
	DeleteJob(string) (int64, error)
	FindJobByUUID(string) (*models.Job, error)
	FindJobByResourceUUID(string) (*models.Job, error)
	FindAllJobs() (*[]models.Job, error)
	FindJobsByState(state int) (*[]models.Job, error)
	FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error)
	JobPromote(r *http.Request) (*models.Job, error)
}

type jobService struct {
	repo repository.JobRepository
}

func NewJobService(repo repository.JobRepository) JobService {
	return &jobService{repo: repo}
}

func (s *jobService) SaveJob(job *models.Job) (*models.Job, error) {
	return s.repo.SaveJob(job)
}

func (s *jobService) UpdateJob(job *models.Job) (*models.Job, error) {
	return s.repo.UpdateJob(job)
}

func (s *jobService) DeleteJob(id string) (int64, error) {
	return s.repo.DeleteJob(id)
}

func (s *jobService) FindJobByUUID(id string) (*models.Job, error) {
	return s.repo.FindJobByUUID(id)
}

func (s *jobService) FindJobByResourceUUID(id string) (*models.Job, error) {
	return s.repo.FindJobByResourceUUID(id)
}

func (s *jobService) FindAllJobs() (*[]models.Job, error) {
	return s.repo.FindAllJobs()
}

func (s *jobService) FindJobsByState(state int) (*[]models.Job, error) {
	return s.repo.FindJobsByState(state)
}

func (s *jobService) FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error) {
	return s.repo.FindJobsToExecute(orchestratorType, ownerID)
}

func (s *jobService) JobPromote(r *http.Request) (*models.Job, error) {
	vars := mux.Vars(r)
	stringJobID := vars["job_uuid"]
	if stringJobID == "" {
		err := errors.New("job ID Cannot be empty")
		logs.Logger.Println("job ID Cannot be empty")
		return nil, err
	}

	var jobOwnershipDTO models.JobOwnershipDTO
	err := json.NewDecoder(r.Body).Decode(&jobOwnershipDTO)
	if err != nil {
		logs.Logger.Printf("Error decoding job patch body: %v", err)
		return nil, err
	}
	if jobOwnershipDTO.OwnerID == "" {
		err := errors.New("owner ID Cannot be empty")
		logs.Logger.Println("owner ID Cannot be empty")
		return nil, err
	}

	jobGotten, err := s.FindJobByUUID(stringJobID)
	if err != nil {
		logs.Logger.Printf("Error retrieving job: %v", err)
		return nil, err
	}

	switch jobGotten.State {
	case models.JobCreated:
		jobGotten.OwnerID = jobOwnershipDTO.OwnerID
		jobGotten.State = models.JobProgressing
	case models.JobProgressing, models.JobFinished, models.JobDegraded:
		err := errors.New("job cannot be promoted")
		logs.Logger.Printf("Job with ID %s is in state %d, cannot be promoted", jobGotten.ID, jobGotten.State)
		return nil, err
	default:
		err := errors.New("job cannot be promoted")
		logs.Logger.Printf("Job with ID %s is in an unknown state, cannot be promoted", jobGotten.ID)
		return nil, err
	}

	updatedJob, err := s.repo.JobPromote(jobGotten)
	if err != nil {
		logs.Logger.Printf("Error updating job state: %v", err)
		return nil, err
	}

	return updatedJob, nil
}
