package rufsBase

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/slices"
)

func TestBase(t *testing.T) {
	os.Remove("./openapi-base.json")
	service := &RufsMicroService{MicroServiceServer: MicroServiceServer{ServeStaticPaths: "../rufs-base-es6/webapp,../rufs-crud-es6/webapp"}}

	serviceRunning := make(chan struct{})
	serviceDone := make(chan struct{})

	go func() {
		close(serviceRunning)
		log.Print("[TestBase] Listen...")

		if err := service.Listen(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[TestBase] Server.Listen : %s", err)
		}

		log.Print("[TestBase] ...Listen is done.")
		defer close(serviceDone)
	}()

	log.Printf("[TestLogin] wait serviceRunning")
	<-serviceRunning
	time.Sleep(1000 * time.Millisecond)
	loginDataReq := RufsUser{RufsUserProteced: RufsUserProteced{Name: "admin"}, Password: "21232f297a57a5a743894a0e4a801fc3"}
	sc := ServerConnection{}
	resp, err := sc.Login("http://localhost:8080", "", "/rest/login", loginDataReq.Name, loginDataReq.Password)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("[TestBase] Error in login request : %d : %s", resp.StatusCode, err)
	}

	if idx := slices.IndexFunc(sc.loginResponse.Roles, func(e Role) bool { return e.Path == "/rufs_user" }); idx >= 0 {
		log.Printf("[TestBase] Role mask of user %s : %b", sc.loginResponse.Name, sc.loginResponse.Roles[idx].Mask)
	} else {
		log.Fatalf("[TestBase] Missing Role mask of user %s", sc.loginResponse.Name)
	}

	for _, schemaName := range []string{"rufsUser"} {
		if _, ok := sc.loginResponse.Openapi.Components.Schemas[schemaName]; !ok {
			log.Fatalf("[TestBase] Missing schema %s", schemaName)
		}
	}

	listUser := []*RufsUser{}
	resp, err = RufsRestRequest(&sc.httpRest, "/rest/rufs_user", http.MethodGet, nil, &listUser, &listUser)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("[TestBase] error in users query request : %d : %s", resp.StatusCode, err)
	}

	if len(listUser) == 0 {
		log.Fatal("[TestBase] retrieved users empty list")
	}

	var foundUser *RufsUser

	for _, user := range listUser {
		if user.Name == "admin" {
			foundUser = user
		}
	}

	if foundUser == nil {
		log.Fatal("[TestBase] don't find user admin")
	}

	sc.lastMessage = NotifyMessage{}
	foundUser.FullName = time.Now().String()
	updatedUser := &RufsUser{}
	query := map[string]any{"id": fmt.Sprint(foundUser.Id)}
	resp, err = RufsRestRequest(&sc.httpRest, "/rest/rufs_user", http.MethodPut, query, foundUser, updatedUser)
	time.Sleep(1000 * time.Millisecond)

	if err != nil || resp.StatusCode != http.StatusOK || sc.lastMessage.Action != "notify" || updatedUser.FullName != foundUser.FullName {
		log.Fatalf("[TestBase] error in update user request : %d : %s", resp.StatusCode, err)
	}

	sc.lastMessage = NotifyMessage{}
	newUserOut := &RufsUser{RufsUserProteced: RufsUserProteced{Name: "tmp"}}
	newUserIn := &RufsUser{}
	resp, err = RufsRestRequest(&sc.httpRest, "/rest/rufs_user", http.MethodPost, nil, newUserOut, newUserIn)
	time.Sleep(1000 * time.Millisecond)

	if err != nil || resp.StatusCode != http.StatusOK || sc.lastMessage.Action != "notify" || newUserOut.Name != newUserIn.Name || newUserIn.Id <= 0 {
		log.Fatalf("[TestBase] error in update user request : %d : %s", resp.StatusCode, err)
	}

	sc.lastMessage = NotifyMessage{}
	query = map[string]any{"id": newUserIn.Id}
	resp, err = RufsRestRequest[RufsUser, RufsUser](&sc.httpRest, "/rest/rufs_user", http.MethodDelete, query, nil, nil)
	time.Sleep(1000 * time.Millisecond)

	if err != nil || resp.StatusCode != http.StatusOK || sc.lastMessage.Action != "delete" {
		log.Fatalf("[TestBase] error in update user request : %d : %s", resp.StatusCode, err)
	}

	resp, err = RufsRestRequest[any, any](&sc.httpRest, "/", http.MethodGet, nil, nil, nil)

	if err != nil || resp.StatusCode != http.StatusOK || !strings.HasPrefix(sc.httpRest.MessageWorking, "<!doctype html>") {
		log.Fatalf("[TestBase] error in get file root : %d : %s", resp.StatusCode, err)
	}

	resp, err = RufsRestRequest[any, any](&sc.httpRest, "/manifest.json", http.MethodGet, nil, nil, nil)

	if err != nil || resp.StatusCode != http.StatusOK || !strings.HasPrefix(sc.httpRest.MessageWorking, "{") {
		log.Fatalf("[TestBase] error in get file 'manifest.json' : %d : %s", resp.StatusCode, err)
	}

	resp, err = RufsRestRequest[any, any](&sc.httpRest, "/es6/CaseConvert.js", http.MethodGet, nil, nil, nil)

	if err != nil || resp.StatusCode != http.StatusOK || !strings.HasPrefix(sc.httpRest.MessageWorking, "class CaseConvert") {
		log.Fatalf("[TestBase] error in get file '/es6/CaseConvert.js' : %d : %s", resp.StatusCode, err)
	}

	resp, err = RufsRestRequest[any, any](&sc.httpRest, "../README.md", http.MethodGet, nil, nil, nil)

	if err == nil && (resp.StatusCode == http.StatusOK || strings.HasPrefix(sc.httpRest.MessageWorking, "#")) {
		log.Fatalf("[TestBase] security fail, application is serving content out of limited scope : %s", sc.httpRest.MessageWorking)
	}

	//	time.Sleep(2000 * time.Millisecond)
	log.Printf("[TestLogin] service.Shutdown()")
	service.Shutdown()
	//	time.Sleep(2000 * time.Millisecond)
	log.Printf("[TestLogin] wait serviceDone")
	<-serviceDone
	log.Printf("[TestLogin] serviceDone")
}

func TestExternal(t *testing.T) {
	service := &RufsMicroService{MicroServiceServer: MicroServiceServer{ServeStaticPaths: "../rufs-base-es6/webapp,../rufs-crud-es6/webapp"}}

	if err := service.Listen(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[TestExternal] Server.Listen : %s", err)
	}
}

type SimulatorMicroService struct {
	RufsMicroService
}

func (rms *SimulatorMicroService) OnRequest(req *http.Request) Response {
	return rms.RufsMicroService.OnRequest(req)
}

func (rms *SimulatorMicroService) LoadFileTables() error {
	rms.RufsMicroService.LoadFileTables()
	var emptyList []map[string]any
	var next func(list []string, idx int)

	next = func(list []string, idx int) {
		if idx < len(list) {
			rms.fileDbAdapter.Load(list[idx], emptyList)
			idx++
			next(list, idx)
		}
	}

	listNames := []string{"Condominium", "Contact", "Unit", "Reading", "Invoice", "Receivable"}
	next(listNames, 0)
	listPath := []string{"/condominium", "/condominium/{cnpj}", "/contact", "/unit", "/reading", "/invoice", "/receivable"}

	if user, err := rms.fileDbAdapter.FindOne("rufsUser", map[string]any{"password": "9CC6D224E7DF4292BD510FD8279DAB35"}); err == nil && user == nil {
		roles := []Role{}

		for _, path := range listPath {
			roles = append(roles, Role{Path: path, Mask: 0xff})
		}

		rms.fileDbAdapter.Insert("rufsUser", map[string]any{"password": "9CC6D224E7DF4292BD510FD8279DAB35", "roles": roles})
	}

	return nil
}

func (rms *SimulatorMicroService) Listen() (err error) {
	if rms.Irms == nil {
		rms.Irms = rms
	}

	if rms.Imss == nil {
		rms.Imss = rms
	}

	return rms.RufsMicroService.Listen()
}

func TestSimulator(t *testing.T) {
	service := &SimulatorMicroService{}
	service.port = 9080
	service.ServeStaticPaths = "../rufs-base-es6/webapp,../rufs-crud-es6/webapp"
	service.appName = "condoconta"
	service.openapiFileName = "../rufs-inetsoft-es6/openapi/condoconta.json"

	if err := service.Listen(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[TestSimulator] Server.Listen : %s", err)
	}
}

type NfeMicroService struct {
	RufsMicroService
}

func TestNfe(t *testing.T) {
	reset := func(dbConfig *DbConfig, fake bool) {
		if fake {
			return
		}

		os.Remove(`openapi-nfe.json`)
		dataSourceName := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s", dbConfig.user, dbConfig.password, dbConfig.database+"_development")
		client, err := sql.Open("pgx", dataSourceName)

		if err != nil {
			log.Fatalf("[TestNfe] : %s", err)
		}

		_, err = client.Exec(fmt.Sprintf(`drop database if exists %s`, dbConfig.database))

		if err != nil {
			log.Fatalf("[TestNfe] : %s", err)
		}

		_, err = client.Exec(fmt.Sprintf(`create database %s`, dbConfig.database))

		if err != nil {
			log.Fatalf("[TestNfe] : %s", err)
		}
	}

	dbConfig := &DbConfig{user: "development", password: "123456", database: "rufs_nfe"}
	reset(dbConfig, false)
	service := &NfeMicroService{RufsMicroService: RufsMicroService{}}
	service.port = 9090
	service.ServeStaticPaths = "../rufs-base-es6/webapp,../rufs-crud-es6/webapp,../rufs-nfe-es6/webapp"
	service.appName = "nfe"
	service.checkRufsTables = true
	service.dbConfig = dbConfig
	service.migrationPath = "../rufs-nfe-es6/sql"

	if err := service.Listen(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[TestSimulator] Server.Listen : %s", err)
	}
}
