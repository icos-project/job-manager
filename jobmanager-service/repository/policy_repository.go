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
	"icos/server/jobmanager-service/models"

	"gorm.io/gorm"
)

type PolicyRepository interface {
	SaveIncompliance(*models.Incompliance) (*models.Incompliance, error)
}

type policyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) PolicyRepository {
	return &policyRepository{db: db}
}

func (repo *policyRepository) SaveIncompliance(incompliance *models.Incompliance) (*models.Incompliance, error) {
	err := repo.db.Debug().Create(incompliance).Error
	if err != nil {
		return &models.Incompliance{}, err
	}
	return incompliance, nil
}
