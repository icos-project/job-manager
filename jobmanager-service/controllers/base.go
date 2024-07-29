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
	"context"
	"flag"
	"fmt"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/service"
	"icos/server/jobmanager-service/utils/logs"
	"net/http"
	"os"
	"os/signal"
	"time"

	"icos/server/jobmanager-service/repository"
	"log"

	go_driver "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Server struct {
	DB              *gorm.DB
	Router          *mux.Router
	JobService      service.JobService
	JobGroupService service.JobGroupService
	PolicyService   service.PolicyService
	ResourceService service.ResourceService
}

func (server *Server) Init() {
	server.Router = mux.NewRouter()
	server.initializeRoutes()
}

func (server *Server) Initialize(dbdriver, dbUser, dbPassword, dbPort, dbHost, dbName string) {

	var err error

	if dbdriver == "mysql" {
		DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
		config := go_driver.Config{
			AllowNativePasswords: true, // deprecate in the future
		}
		server.DB, err = gorm.Open(mysql.New(
			mysql.Config{
				DSN:       DBURL,
				DSNConfig: &config,
			}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err != nil {
			fmt.Printf("Cannot connect to %s database", dbdriver)
			log.Fatal("This is the error:", err)
			DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort)
			config := go_driver.Config{
				AllowNativePasswords: true, // deprecate in the future
			}
			server.DB, err = gorm.Open(mysql.New(
				mysql.Config{
					DSN:       DBURL,
					DSNConfig: &config,
				}), &gorm.Config{})
			server.DB.Exec("USE " + dbName)
		} else {
			fmt.Printf("We are connected to the %s database", dbdriver)
		}
		// first time schema creation
		server.DB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";")
		server.DB.Exec("USE " + dbName)
	}

	server.DB.Debug().
		AutoMigrate(
			&models.JobGroup{},
			&models.Job{},
			&models.PlainManifest{},
			&models.Target{},
			&models.Resource{},
			&models.Condition{},
			&models.Incompliance{},
			&models.Subject{})

	server.Router = mux.NewRouter()

	// Initialize repositories
	jobRepo := repository.NewJobRepository(server.DB)
	jobGroupRepo := repository.NewJobGroupRepository(server.DB)
	policyRepo := repository.NewPolicyRepository(server.DB)
	resourceRepo := repository.NewResourceRepository(server.DB)
	httpClient := &http.Client{}

	// Initialize services
	server.JobService = service.NewJobService(jobRepo)
	server.JobGroupService = service.NewJobGroupService(jobGroupRepo)
	// TODO: we should reference a single httpclient for all services
	server.PolicyService = service.NewPolicyService(policyRepo, jobRepo, httpClient)
	server.ResourceService = service.NewResourceService(resourceRepo, jobRepo)

	// swagger
	server.Router.PathPrefix("/jobmanager/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	logs.Logger.Println("Listening to port " + addr + " ...")
	handler := cors.AllowAll().Handler(server.Router)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		// init server
		if err := http.ListenAndServe(addr, handler); err != nil {
			if err != http.ErrServerClosed {
				logs.Logger.Fatal(err)
			}
		}
	}()

	<-stop

	// after stopping server
	logs.Logger.Println("Closing connections ...")

	var shutdownTimeout = flag.Duration("shutdown-timeout", 10*time.Second, "shutdown timeout (5s,5m,5h) before connections are cancelled")
	_, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()
}
