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

package controllers

import (
	"icos/server/jobmanager-service/responses"
	"icos/server/jobmanager-service/utils/logs"
	"io"
	"net/http"
)

// CreatePolicyIncompliance godoc
//
//	@Summary		Create new Policy Incompliance
//	@Description	create new policy incompliance
//	@Tags			policies
//	@Accept			plain
//	@Produce		json
//	@Param			application	body		string	true	"Incompliance Object"
//	@Success		200			{object}	models.Incompliance
//	@Failure		400			{object}	string	"Incompliance Object is not correct"
//	@Failure		422			{object}	string	"Unprocessable Entity"
//	@Router			/jobmanager/policies/incompliance [post]
func (server *Server) CreatePolicyIncompliance(w http.ResponseWriter, r *http.Request) {
	incomplianceBody, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Handle the incompliance through the service
	incompliance, err := server.PolicyService.HandlePolicyIncompliance(incomplianceBody)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, incompliance)
}
