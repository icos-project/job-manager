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
	mocks "icos/server/jobmanager-service/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func initPolicyRepo(db *gorm.DB) interface{} {
	return NewPolicyRepository(db)
}

func TestSaveIncompliance(t *testing.T) {
	repo := mocks.SetupTest(t, initPolicyRepo).(PolicyRepository)

	incom := &models.Incompliance{}

	result, err := repo.SaveIncompliance(incom)
	assert.NoError(t, err)
	assert.Equal(t, incom.ID, result.ID)
}
