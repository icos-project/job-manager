# Job Manager

## Overview
The Job Manager component represents the core module of the ICOS Controller runtime. This module is responsible for the runtime management and provides control, persistence and the coordination between ICOS components. Such a component enables the continuum to be consistent within real time, furthermore becoming the centre of the truth in the mentioned continuum.

For ICOS to become trustable at runtime, and consistently manage ICOS execution in such a diverse continuum, Job Manager provides the notion of a Job. The Job Manager provides job lifecycle management operations (CRUD) that enable consistent and efficient execution of a job, furthermore, managing both the state of the job and the state of the actual resource within the agent that owes that job.
## Table of Contents

- [1. ICOS Concepts](#1-icos-concepts)
    - [Jobs](#jobs)
    - [Application Component Description](#application-component-description)
    - [Target](#target)
    - [Resource](#resource)
    - [ICOS Agent](#icos-agent)
    - [Job Groups](#job-groups)
- [2. Swagger Job Manager API](#2-swagger-job-manager-api)
- [3. Models](#3-models)
- [4. Docker Installation](#4-docker-installation)
- [5. Kind Installation](#5-kind-installation)
- [6. Usage](#6-usage)
- [7. Contributing](#7-contributing)
- [8. Legal](#8-legal)

## 1. ICOS Concepts
### Jobs
A Job defines the minimal executable unit to be managed by ICOS Controller at runtime. For this unit to become executable by the different ICOS Agents, more importantly, without considering their underlying runtime technology.
### Job Groups
When an application is composed of multiple components it becomes a set of jobs, in other words, a job group. Job group holds all the information regarding the application, including all the components(jobs) that compose the application, the relationship between the different jobs (application topology) and other information such as requirements and policies such application must meet.
### Application Component Description
Describes, following ICOS syntax, the components an application is composed of, as well as their requirements to be met and policies to be enforced. Currently, the Application Descriptor may vary from version to version, so it is recommended to check the [Application Descriptor Repository](https://production.eng.it/gitlab/icos/meta-kernel/application-descriptor).
### Target
Defines the underlying infrastructure able to execute a single job, taking into consideration the mentioned requirements and the capacity the infrastructure piece provides, since appropriate quality of service must be enforced. This target is selected by the Matchmaker (described in the next section) for each job that comprises an application.
### Resource
Abstract representation of the actual application component executed within a single target. This representation is retrieved from the orchestrator at runtime during the execution of the corresponding job.
### ICOS Agent
The actual underlying orchestrator that manages the infrastructure where the job is executed. From a Job Manager’s perspective, whenever an agent takes a job for execution, it becomes the owner of such a job, meaning that only the mentioned agent can manage this job’s life cycle.

## 2. Swagger Job Manager API
ICOS Job Manager Microservice.

## Version: 1.0

### Terms of service
http://swagger.io/terms/

**Contact information:**  
API Support  
http://www.swagger.io/support  
support@swagger.io  

**License:** [Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)

[OpenAPI](https://swagger.io/resources/open-api/)
### Security
**OAuth 2.0**  

|basic|*Basic*|
|---|---|

## Endpoints

### /jobmanager/groups

#### GET
##### Description:
Get all jobgroups

##### Parameters:
None

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200  | OK          | [models.JobGroup](#models.JobGroup) |
| 400  | Bad Request | string |
| 404  | Not Found   | string |

#### POST
##### Description:
Create new jobgroup

##### Parameters:

| Name        | Located in | Description              | Required | Schema |
| ----------- | ---------- | ------------------------ | -------- | ------ |
| application | body       | Application manifest YAML | Yes      | string |

##### Responses

| Code | Description          | Schema |
| ---- | -------------------- | ------ |
| 201  | Created              | [models.JobGroup](#models.JobGroup) |
| 422  | Unprocessable Entity | string |

#### PUT
##### Description:
Update a jobgroup

##### Parameters:

| Name     | Located in | Description       | Required | Schema |
| -------- | ---------- | ----------------- | -------- | ------ |
| JobGroup | body       | JobGroup information | Yes      | [models.JobGroup](#models.JobGroup) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200  | OK          | [models.JobGroup](#models.JobGroup) |
| 400  | Bad Request | string |
| 404  | Not Found   | string |

### /jobmanager/groups/{group_uuid}

#### DELETE
##### Description:
Delete job group by UUID

##### Parameters:

| Name       | Located in | Description     | Required | Schema |
| ---------- | ---------- | --------------- | -------- | ------ |
| group_uuid | path       | JobGroup UUID   | Yes      | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200  | OK          | string |
| 400  | Bad Request | string |
| 404  | Not Found   | string |

#### GET
##### Description:
Get jobgroup by UUID

##### Parameters:

| Name       | Located in | Description     | Required | Schema |
| ---------- | ---------- | --------------- | -------- | ------ |
| group_uuid | path       | JobGroup UUID   | Yes      | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200  | OK          | [models.JobGroup](#models.JobGroup) |
| 400  | Bad Request | string |
| 404  | Not Found   | string |

### /jobmanager/groups/undeploy/{group_uuid}

#### PUT
##### Description:
Stop jobgroup by UUID

##### Parameters:

| Name       | Located in | Description     | Required | Schema |
| ---------- | ---------- | --------------- | -------- | ------ |
| group_uuid | path       | JobGroup UUID   | Yes      | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200  | OK          | [models.JobGroup](#models.JobGroup) |
| 400  | Bad Request | string |
| 404  | Not Found   | string |

### /jobmanager/jobs

#### GET
##### Description:
Get all jobs

##### Parameters:
None

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200  | OK          | array of [models.Job](#models.Job) |
| 404  | Not Found   | string |

#### PUT
##### Description:
Update a job

##### Parameters:

| Name     | Located in | Description       | Required | Schema |
| -------- | ---------- | ----------------- | -------- | ------ |
| job_uuid | path       | Job UUID          | Yes      | string |
| Job      | body       | Job information   | Yes      | [models.Job](#models.Job) |

##### Responses

| Code | Description          | Schema |
| ---- | -------------------- | ------ |
| 200  | OK                   | [models.Job](#models.Job) |
| 400  | Job UUID is required | string |
| 404  | Not Found            | string |

### /jobmanager/jobs/{job_uuid}

#### DELETE
##### Description:
Delete job by UUID

##### Parameters:

| Name     | Located in | Description | Required | Schema |
| -------- | ---------- | ----------- | -------- | ------ |
| job_uuid | path       | Job UUID    | Yes      | string |

##### Responses

| Code | Description          | Schema |
| ---- | -------------------- | ------ |
| 200  | Ok                   | string |
| 400  | Job UUID is required | string |
| 404  | Not Found            | string |

#### GET
##### Description:
Get job by UUID

##### Parameters:

| Name     | Located in | Description | Required | Schema |
| -------- | ---------- | ----------- | -------- | ------ |
| job_uuid | path       | Job UUID    | Yes      | string |

##### Responses

| Code | Description          | Schema |
| ---- | -------------------- | ------ |
| 200  | Ok                   | [models.Job](#models.Job) |
| 400  | Job UUID is required | string |
| 404  | Not Found            | string |

### /jobmanager/jobs/executable/{orchestrator}/{owner_id}

#### GET
##### Description:
Get jobs to execute

##### Parameters:

| Name        | Located in | Description                      | Required | Schema |
| ----------- | ---------- | -------------------------------- | -------- | ------ |
| orchestrator| path       | Orchestrator type [ocm | nuvla]  | Yes      | string |
| owner_id    | path       | Owner ID                         | Yes      | string |

##### Responses

| Code | Description                       | Schema |
| ---- | --------------------------------- | ------ |
| 200  | OK                                | array of [models.Job](#models.Job) |
| 400  | Orchestrator type is required     | string |
| 404  | Not Found                         | string |

### /jobmanager/jobs/promote/{job_uuid}

#### PATCH
##### Description:
Promote job by UUID

##### Parameters:

| Name     | Located in | Description | Required | Schema |
| -------- | ---------- | ----------- | -------- | ------ |
| job_uuid | path       | Job UUID    | Yes      | string |

##### Responses

| Code | Description          | Schema |
| ---- | -------------------- | ------ |
| 204  | Job Promoted         | string |
| 400  | Job UUID is required | string |
| 404  | Not Found            | string |

### /jobmanager/policies/incompliance

#### POST
##### Description:
Create new policy incompliance

##### Parameters:

| Name        | Located in | Description          | Required | Schema |
| ----------- | ---------- | -------------------- | -------- | ------ |
| application | body       | Incompliance Object  | Yes      | string |

##### Responses

| Code | Description                          | Schema |
| ---- | ------------------------------------ | ------ |
| 200  | OK                                   | [models.Incompliance](#models.Incompliance) |
| 400  | Incompliance Object is not correct   | string |
| 422  | Unprocessable Entity                 | string |

### /jobmanager/resources/status

#### PUT
##### Description:
Update resource status by UUID

##### Parameters:

| Name      | Located in | Description      | Required | Schema |
| --------- | ---------- | ---------------- | -------- | ------ |
| id        | path       | Resource UUID    | Yes      | string |
| resource  | body       | Resource info    | Yes      | [models.Resource](#models.Resource) |

##### Responses

| Code | Description                 | Schema |
| ---- | --------------------------- | ------ |
| 200  | Resource updated            | string |
| 400  | Resource UUID is required   | string |
| 404  | Not Found                   | string |

### /jobmanager/resources/status/{job_uuid}

#### GET
##### Description:
Get resource status by job UUID

##### Parameters:

| Name     | Located in | Description | Required | Schema |
| -------- | ---------- | ----------- | -------- | ------ |
| job_uuid | path       | Job UUID    | Yes      | string |

##### Responses

| Code | Description          | Schema |
| ---- | -------------------- | ------ |
| 200  | OK                   | [models.Resource](#models.Resource) |
| 400  | Job UUID is required | string |
| 404  | Not Found            | string |


## 3. Models
### models.Condition

| Name                | Type   | Description               |
| ------------------- | ------ | ------------------------- |
| created_at          | string |                           |
| id                  | integer|                           |
| lastTransitionTime  | string |                           |
| message             | string |                           |
| observedGeneration  | integer|                           |
| reason              | string |                           |
| status              | string | [models.ConditionStatus](#models.ConditionStatus) |
| type                | string | [models.ResourceState](#models.ResourceState) |
| updated_at          | string |                           |

### models.ConditionStatus

| Name             | Type   | Description               |
| ---------------- | ------ | ------------------------- |
| ConditionTrue    | string | "True"                    |
| ConditionFalse   | string | "False"                   |
| ConditionUnknown | string | "Unknown"                 |

### models.Incompliance

| Name                | Type           | Description                              |
| ------------------- | -------------- | ---------------------------------------- |
| created_at          | string         |                                          |
| currentValue        | string         | ResourceID                               |
| extraLabels         | [models.StringMap](#models.StringMap) |             |
| id                  | string         |                                          |
| measurementBackend  | string         |                                          |
| policyId            | string         |                                          |
| policyName          | string         |                                          |
| remediation         | [models.RemediationType](#models.RemediationType) |    |
| subject             | [models.Subject](#models.Subject) |                  |
| threshold           | string         |                                          |
| updated_at          | string         |                                          |

### models.Job

| Name                    | Type    | Description                               |
| ----------------------- | ------- | ----------------------------------------- |
| created_at              | string  |                                           |
| id                      | string  |                                           |
| job_group_description   | string  |           |
| job_group_id            | string  |                                           |
| job_group_name          | string  |           |
| manifests               | array of [models.PlainManifest](#models.PlainManifest) | |
| namespace               | string  |                                           |
| orchestrator            | [models.OrchestratorType](#models.OrchestratorType) |  |
| owner_id                | string  |                                           |
| resource                | [models.Resource](#models.Resource) |              |
| state                   | [models.JobState](#models.JobState) |                |
| sub_type                | [models.RemediationType](#models.RemediationType) |  |
| targets                 | [models.Target](#models.Target) |                    |
| type                    | [models.JobType](#models.JobType) |                  |
| updated_at              | string  |                                           |

### models.JobGroup

| Name            | Type              | Description                              |
| --------------- | ----------------- | ---------------------------------------- |
| appDescription  | string            |         |
| appName         | string            |          |
| created_at      | string            |                                          |
| id              | string            |                                          |
| jobs            | array of [models.Job](#models.Job) |                            |
| updated_at      | string            |                                          |

### models.JobState

| Name           | Type    | Description               |
| -------------- | ------- | ------------------------- |
| JobCreated     | integer | 1                         |
| JobProgressing | integer | 2                         |
| JobFinished    | integer | 3                         |
| JobDegraded    | integer | 4                         |

### models.JobType

| Name              | Type    | Description               |
| ----------------- | ------- | ------------------------- |
| CreateDeployment  | integer | 5                         |
| DeleteDeployment  | integer | 6                         |
| UpdateDeployment  | integer | 7                         |
| ReplaceDeployment | integer | 8                         |

### models.OrchestratorType

| Name   | Type   | Description               |
| ------ | ------ | ------------------------- |
| OCM    | string | ocm                       |
| Nuvla  | string | nuvla                     |
| None   | string | ""                        |

### models.PlainManifest

| Name         | Type   | Description               |
| ------------ | ------ | ------------------------- |
| created_at   | string |                           |
| id           | integer|                           |
| updated_at   | string |                           |
| yamlString   | string |                           |

### models.RemediationType

| Name              | Type    | Description               |
| ----------------- | ------- | ------------------------- |
| ScaleUp           | string  | scale-up                  |
| ScaleDown         | string  | scale-down                |
| ScaleOut          | string  | scale-out                 |
| ScaleIn           | string  | scale-in                  |
| PatchDeployment   | string  | patch                     |
| Reallocation      | string  | reallocation              |

### models.Resource

| Name                | Type          | Description               |
| ------------------- | ------------- | ------------------------- |
| conditions          | array of [models.Condition](#models.Condition) |    |
| created_at          | string        |                           |
| id                  | string        |                           |
| job_id              | string        |                           |
| resource_name       | string        |                           |
| resource_uuid       | string        |                           |
| updated_at          | string        |                           |

### models.ResourceState

| Name       | Type   | Description               |
| ---------- | ------ | ------------------------- |
| Progressing| string | Progressing               |
| Applied    | string | Applied                   |
| Available  | string | Available                 |
| Degraded   | string | Degraded                  |

### models.StringMap

| Name             | Type   | Description               |
| ---------------- | ------ | ------------------------- |
| additionalProperties | string |                           |

### models.Subject

| Name          | Type   | Description               |
| ------------- | ------ | ------------------------- |
| appComponent  | string |                           |
| appInstance   | string |                           |
| appName       | string |                           |
| created_at    | string |                           |
| id            | string |                           |
| resourceId    | string |                           |
| type          | string |                           |
| updated_at    | string |                           |

### models.Target

| Name          | Type   | Description               |
| ------------- | ------ | ------------------------- |
| cluster_name  | string |                           |
| created_at    | string |                           |
| id            | integer|                           |
| node_name     | string |                           |
| orchestrator  | [models.OrchestratorType](#models.OrchestratorType) |       |
| updated_at    | string |                           |


## 4. Docker Installation
To install and run the `job-manager`, follow these steps:

1. Clone the repository:
    ```sh
    git clone https://production.eng.it/gitlab/icos/meta-kernel/job-manager.git
    ```

2. Navigate to the project directory:
    ```sh
    cd job-manager
    ```

3. Build the Docker image:
    ```sh
    docker build -t job-manager .
    ```

4. Run the Docker container:
    ```sh
    docker run -p 8082:8082 job-manager
    ```

## 5. Kind Installation

Please, refer to the helm suite in [ICOS Controller Repository](https://production.eng.it/gitlab/icos/suites/icos-controller)

## 6. Usage

After running the service, you can use tools like `curl`, `Postman`, `Swagger` or any other API client to interact with the endpoints. Beware that you will need a Keycloak Token to perform requests to this service.

### Example Request
```sh
curl --location 'http://localhost:8082/jobmanager/jobs' \
--header 'Authorization: Bearer (Token)'
```

## 7. Contributing

In order to contribute to this repository, feel free to open a pull request and assign `@x_alvolkov`or `x_magallar` as a reviewer.

## 8. Legal

The Job Manager is released under the Apache 2.0 license.
Copyright © 2022-2024 Eviden. All rights reserved.

🇪🇺 This work has received funding from the European Union's HORIZON research and innovation programme under grant agreement No. 101070177.
