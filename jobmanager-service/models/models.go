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

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"icos/server/jobmanager-service/utils/logs"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Metadata for the database entities
type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Base entities with UUID and UINT
type BaseUUID struct {
	Metadata
	ID string `gorm:"type:char(36);primary_key;"`
}

type BaseUINT struct {
	Metadata
	ID uint32 `gorm:"primary_key;autoIncrement"`
}

// Validator instance
var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// JobGroup entity
type JobGroup struct {
	BaseUUID
	AppName        string `json:"appName"`        // add validation when unmocking mm
	AppDescription string `json:"appDescription"` // add validation when unmocking mm
	Jobs           []Job  `json:"jobs" validate:"dive,required"`
}

func (jg *JobGroup) Validate() error {
	return validate.Struct(jg)
}

// GORM hooks for JobGroup
func (jg *JobGroup) BeforeCreate(tx *gorm.DB) (err error) {
	if jg.ID == "" {
		jg.ID = uuid.New().String()
	}
	logs.Logger.Print("JobGroup ID: ", jg.ID)
	return nil
}

func (jg *JobGroup) AfterCreate(tx *gorm.DB) (err error) {
	if err := jg.Validate(); err != nil {
		return err
	}
	return nil
}

func (jg *JobGroup) BeforeUpdate(tx *gorm.DB) (err error) {
	return jg.Validate()
}

// Job entity
type Job struct {
	BaseUUID
	JobGroupID          string           `gorm:"type:char(36);not null" json:"job_group_id" validate:"omitempty,uuid4"`
	OwnerID             string           `gorm:"type:char(36);default:''" json:"owner_id,omitempty" validate:"omitempty,uuid4"`
	JobGroupName        string           `gorm:"type:text" json:"job_group_name"`                  // add validation when unmocking mm
	JobGroupDescription string           `gorm:"type:text" json:"job_group_description,omitempty"` // add validation when unmocking mm
	Type                JobType          `gorm:"type:text" json:"type,omitempty"`
	SubType             RemediationType  `gorm:"type:text" json:"sub_type,omitempty"`
	State               JobState         `gorm:"type:text" json:"state,omitempty"`
	Manifests           []PlainManifest  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"manifests" validate:"dive"`
	Targets             Target           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"targets,omitempty" validate:"omitempty"`
	Orchestrator        OrchestratorType `gorm:"type:text" json:"orchestrator"` // check why required fails when dm updates job for orchestrator
	Resource            *Resource        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"resource,omitempty"`
	Namespace           string           `gorm:"type:text" json:"namespace,omitempty" validate:"omitempty"`
}

func (j *Job) Validate() error {
	return validate.Struct(j)
}

// GORM hooks for Job
func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	logs.Logger.Print("Job ID: ", j.ID)
	return nil
}

func (j *Job) AfterCreate(tx *gorm.DB) (err error) {
	if err := j.Validate(); err != nil {
		return err
	}

	return nil
}

func (j *Job) BeforeUpdate(tx *gorm.DB) (err error) {
	return j.Validate()
}

// Resource entity
type Resource struct {
	BaseUUID
	JobID        string      `gorm:"type:char(36);not null" json:"job_id" validate:"omitempty,uuid4"`
	ResourceUID  string      `gorm:"type:char(36);default:''" json:"resource_uuid,omitempty"`
	ResourceName string      `gorm:"type:text" json:"resource_name,omitempty" validate:"omitempty"`
	Conditions   []Condition `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"conditions,omitempty" validate:"dive"`
}

// GORM hooks for Resource TODO: Add validation
func (r *Resource) BeforeCreate(tx *gorm.DB) (err error) {
	logs.Logger.Print("Inside Resource BeforeCreate")
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

func (r *Resource) AfterUpdate(tx *gorm.DB) (err error) {
	logs.Logger.Print("Inside Resource AfterUpdate")
	logs.Logger.Print("Resource ID: ", r.ID)
	logs.Logger.Print("Resource Name: ", r.ResourceName)
	return nil
}

// Condition entity
type Condition struct {
	BaseUINT
	ResourceID         string          `gorm:"type:char(36);index" json:"-" validate:"omitempty,uuid4"`
	Type               ResourceState   `gorm:"type:text" json:"type" validate:"required"`
	Status             ConditionStatus `gorm:"type:text" json:"status" validate:"required"`
	ObservedGeneration int64           `gorm:"type:bigint" json:"observedGeneration,omitempty" validate:"omitempty"`
	LastTransitionTime time.Time       `gorm:"type:timestamp" json:"lastTransitionTime" validate:"required"`
	Reason             string          `gorm:"type:text" json:"reason" validate:"required"`
	Message            string          `gorm:"type:text" json:"message" validate:"required"`
}

// PlainManifest entity
type PlainManifest struct {
	BaseUINT
	JobID      string `gorm:"type:char(36);not null" json:"-" validate:"omitempty,uuid4"`
	YamlString string `gorm:"type:text" json:"yamlString" validate:"required"`
}

// Target entity
type Target struct {
	BaseUINT
	JobID        string           `gorm:"type:char(36);index;not null" json:"-" validate:"omitempty,uuid4"`
	ClusterName  string           `json:"cluster_name" validate:"required"`
	NodeName     string           `json:"node_name,omitempty" validate:"omitempty"`
	Orchestrator OrchestratorType `gorm:"type:text" json:"orchestrator" validate:"required"`
}

// Incompliance entity
type Incompliance struct {
	BaseUUID
	//	ResourceID         string          `gorm:"type:char(36);not null" json:"id" validate:"omitempty,uuid4"`
	CurrentValue       string          `gorm:"type:text" json:"currentValue,omitempty" validate:"omitempty"`
	Threshold          string          `gorm:"type:text" json:"threshold,omitempty" validate:"omitempty"`
	PolicyName         string          `gorm:"type:text" json:"policyName" validate:"required"`
	PolicyID           string          `gorm:"type:char(36)" json:"policyId" validate:"omitempty,uuid4"`
	MeasurementBackend string          `gorm:"type:text" json:"measurementBackend,omitempty" validate:"omitempty"`
	ExtraLabels        StringMap       `gorm:"type:json" json:"extraLabels,omitempty" validate:"omitempty"`
	Subject            Subject         `json:"subject,omitempty"`
	Remediation        RemediationType `gorm:"type:text" json:"remediation" validate:"required"`
}

// GORM hooks for Resource TODO: Add validation
func (i *Incompliance) BeforeCreate(tx *gorm.DB) (err error) {
	logs.Logger.Print("Inside Resource BeforeCreate")
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

// Subject entity
type Subject struct {
	BaseUUID
	IncomplianceID string `gorm:"type:char(36);notnull" json:"-" validate:"omitempty,uuid4"`
	Type           string `gorm:"type:text" json:"type,omitempty" validate:"omitempty"`
	AppName        string `gorm:"type:text" json:"appName,omitempty" validate:"omitempty"`
	AppComponent   string `gorm:"type:text" json:"appComponent,omitempty" validate:"omitempty"`
	AppInstance    string `gorm:"type:text" json:"appInstance,omitempty" validate:"omitempty"`
	ResourceID     string `gorm:"type:char(36)" json:"resourceId,omitempty" validate:"omitempty,uuid4"`
}

// GORM hooks for Resource TODO: Add validation
func (s *Subject) BeforeCreate(tx *gorm.DB) (err error) {
	logs.Logger.Print("Inside Resource BeforeCreate")
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// Policy Manager DTOs
type (
	Notification struct {
		AppInstance  string `json:"app_instance"`
		CommonAction Action `json:"common_action"`
		Service      string `json:"service"`
		Manifest     string `json:"app_descriptor"`
	}

	Action struct {
		URI                string            `json:"uri"`
		HTTPMethod         string            `json:"http_method"`
		IncludeAccessToken bool              `json:"include_access_token"`
		ExtraParameters    map[string]string `json:"extra_parameters"`
	}

	ExtraParameters struct {
		JobGroupId uuid.UUID `json:"job_group_id"`
	}
)

// Other DTOs
type (
	// Matchmaker DTOs
	JobGroupHeader struct {
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Components  []Component              `json:"components"`
		Policies    []Policy                 `json:"policies"`
		Manifests   []map[string]interface{} `json:"manifests"`
	}
	/* The `targets` field in the `Component` struct is defined as an `interface{}` because it can
	contain either an object (a single target) or an empty array. This variability requires
	special handling to ensure we correctly parse and utilize the data.*/
	Component struct {
		Name         string        `json:"name" yaml:"name"`
		Type         string        `json:"type" yaml:"type"`
		Manifests    []ManifestDTO `json:"manifests" yaml:"manifests"` // this field only contains the name of the manifest, not the actual manifest
		Requirements Requirement   `json:"requirements,omitempty" yaml:"requirements"`
		Policies     []Policy      `json:"policies,omitempty" yaml:"policies"`
		Targets      interface{}   `json:"targets" yaml:"targets"`
	}

	Policy struct {
		Name         string          `json:"name"`
		Component    string          `json:"component"`
		FromTemplate string          `json:"fromTemplate,omitempty"`
		Spec         *PolicySpec     `json:"spec,omitempty"`
		Remediation  string          `json:"remediation,omitempty"`
		Variables    PolicyVariables `json:"variables"`
	}
	PolicySpec struct {
		Expr       string     `json:"expr"`
		Thresholds Thresholds `json:"thresholds"`
	}

	// Thresholds structure
	Thresholds struct {
		Warning  int `json:"warning"`
		Critical int `json:"critical"`
	}

	// PolicyVariables structure
	PolicyVariables struct {
		ThresholdTimeSeconds int    `json:"thresholdTimeSeconds"`
		CompssTask           string `json:"compssTask"`
	}

	Requirement struct {
		Device       string `json:"devices,omitempty" yaml:"devices"`
		CPU          string `json:"cpu,omitempty" yaml:"cpu"`
		Memory       string `json:"memory,omitempty" yaml:"memory"`
		Architecture string `json:"architecture,omitempty" yaml:"architecture"`
	}

	ManifestDTO struct {
		Name string `json:"name" yaml:"name"`
	}
	// Deploy Manager
	JobOwnershipDTO struct {
		OwnerID string `json:"owner_id"`
	}

	Manifest struct {
		Name string `json:"name"`
	}
)

// Enum-like Types
type (
	RemediationType  string
	ResourceState    string
	ConditionStatus  string
	JobState         int
	JobType          int
	OrchestratorType string
	StringMap        map[string]string
)

// RemediationType Enum
const (
	ScaleUp         RemediationType = "scale-up"
	ScaleDown       RemediationType = "scale-down"
	ScaleOut        RemediationType = "scale-out"
	ScaleIn         RemediationType = "scale-in"
	PatchDeployment RemediationType = "patch"
	Reallocation    RemediationType = "reallocation"
)

// ResourceState Enum
const (
	Progressing ResourceState = "Progressing"
	Applied     ResourceState = "Applied"
	Available   ResourceState = "Available"
	Degraded    ResourceState = "Degraded"
)

// ConditionStatus Enum
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// OrchestratorType Enum
const (
	OCM   OrchestratorType = "ocm"
	Nuvla OrchestratorType = "nuvla"
	None  OrchestratorType = ""
)

// OrchestratorType Enum Mapper
func OrchestratorTypeMapper(orchestratorType string) OrchestratorType {
	switch orchestratorType {
	case string(Nuvla):
		return Nuvla
	case string(OCM):
		return OCM
	default:
		return None
	}
}

// JobState and JobType Enums
const (
	JobCreated JobState = iota + 1
	JobProgressing
	JobFinished
	JobDegraded

	CreateDeployment JobType = iota + 1
	DeleteDeployment
	UpdateDeployment
	ReplaceDeployment
)

// ExtraLabels Mapper
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

var (
	PolicyManagerBaseURL = os.Getenv("POLICYMANAGER_URL")
	// lighthouseBaseURL  = os.Getenv("LIGHTHOUSE_BASE_URL")
	MatchmakerBaseURL = os.Getenv("MATCHMAKING_URL")

	JobTypeFromString = map[string]JobType{
		"CreateDeployment":  CreateDeployment,
		"DeleteDeployment":  DeleteDeployment,
		"UpdateDeployment":  UpdateDeployment,
		"ReplaceDeployment": ReplaceDeployment,
	}
)
