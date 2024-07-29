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
	"errors"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/utils/logs"

	"gorm.io/gorm"
)

type ResourceRepository interface {
	SaveResource(*models.Resource) (*models.Resource, error)
	UpdateAResource(*models.Resource) (*models.Resource, error)
	AddCondition(*models.Resource, *models.Condition) (*models.Resource, error)
	RemoveConditions(*models.Resource) (*models.Resource, error)
	FindResourceByJobUUID(string) (*models.Resource, error)
}

type resourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (repo *resourceRepository) UpdateAResource(resource *models.Resource) (*models.Resource, error) {
	logs.Logger.Println("Updating the resource: " + resource.ID)
	repo.db = repo.db.Session(&gorm.Session{FullSaveAssociations: true}).Where("job_id = ?", resource.JobID).Updates(&models.Resource{ResourceUID: resource.ResourceUID, ResourceName: resource.ResourceName})
	if repo.db.Error != nil {
		return &models.Resource{}, repo.db.Error
	}

	// This is the display the updated Job
	err := repo.db.Debug().Model(models.Resource{}).Where("job_id = ?", resource.JobID).Preload("Conditions").Take(&resource).Error
	if err != nil {
		return &models.Resource{}, err
	}
	return resource, nil
}

func (repo *resourceRepository) AddCondition(resource *models.Resource, condition *models.Condition) (*models.Resource, error) {
	logs.Logger.Println("Updating the resource: " + resource.ID)
	err := repo.db.Debug().Create(&condition)
	if err != nil {
		return &models.Resource{}, repo.db.Error
	}
	// This is the display the updated Job
	err = repo.db.Debug().Model(models.Resource{}).Where("resource_uuid =?", resource.ResourceUID).Preload("Conditions").Take(&resource)
	return resource, err.Error
}

func (repo *resourceRepository) RemoveConditions(resource *models.Resource) (*models.Resource, error) {
	logs.Logger.Println("Removing old status of the resource: " + resource.ID)
	err := repo.db.Debug().Model(&models.Resource{}).Where("resource_id =?", resource.ID).Delete(&models.Condition{}).Error
	if err != nil {
		return &models.Resource{}, repo.db.Error
	}
	return resource, err
}

func (repo *resourceRepository) SaveResource(resource *models.Resource) (*models.Resource, error) {
	err := repo.db.Debug().Create(&resource).Error
	if err != nil {
		return &models.Resource{}, err
	}
	return resource, nil
}

func (repo *resourceRepository) FindResourceByJobUUID(jobId string) (*models.Resource, error) {
	resource := &models.Resource{}
	err := repo.db.Debug().Model(models.Resource{}).Where("job_id = ?", jobId).Preload("Conditions").Take(&resource).Error
	if err != nil {
		return &models.Resource{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.Resource{}, errors.New("job Not Found")
	}
	return resource, err
}
