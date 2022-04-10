package rufsBase

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBase(t *testing.T) {
	service := &RufsMicroService{MicroServiceServer: MicroServiceServer{ServeStaticPaths: "../rufs-base-es6/webapp,../rufs-crud-es6/webapp"}}
	//	var service IMicroServiceServer = &RufsMicroService{}
	service.Init(service)
	serviceRunning := make(chan struct{})
	serviceDone := make(chan struct{})

	go func() {
		close(serviceRunning)
		log.Print("[TestBase] Listen...")

		if err := service.Listen(); err != nil && err != http.ErrServerClosed {
			log.Fatal("[TestBase] Unexpected server closed !")
		}

		log.Print("[TestBase] ...Listen.")
		defer close(serviceDone)
	}()

	log.Printf("[TestLogin] wait serviceRunning")
	<-serviceRunning
	time.Sleep(1000 * time.Millisecond)
	loginDataReq := RufsUser{Name: "admin", Password: "21232f297a57a5a743894a0e4a801fc3"}
	sc := ServerConnection{}
	resp, err := sc.Login("http://localhost:8080", "", "/rest/login", loginDataReq.Name, loginDataReq.Password)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("[TestBase] Error in login request : %d : %s", resp.StatusCode, err)
	}

	if role, ok := sc.loginResponse.Roles["rufsUser"]; ok {
		log.Printf("[TestBase] Role mask of user %s : %b", sc.loginResponse.Name, role)
	} else {
		log.Fatalf("[TestBase] Missing Role mask of user %s", sc.loginResponse.Name)
	}

	listUser := []*RufsUser{}
	resp, err = RufsRestRequest(&sc.httpRest, "/rest/rufs_user/query", http.MethodGet, nil, &listUser, &listUser)

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

	if err != nil || resp.StatusCode != http.StatusOK || sc.lastMessage.Action != "notify" || updatedUser.FullName != foundUser.FullName {
		log.Fatalf("[TestBase] error in update user request : %d : %s", resp.StatusCode, err)
	}

	sc.lastMessage = NotifyMessage{}
	newUserOut := &RufsUser{Name: "tmp"}
	newUserIn := &RufsUser{}
	resp, err = RufsRestRequest(&sc.httpRest, "/rest/rufs_user", http.MethodPost, nil, newUserOut, newUserIn)

	if err != nil || resp.StatusCode != http.StatusOK || sc.lastMessage.Action != "notify" || newUserOut.Name != newUserIn.Name || newUserIn.Id <= 0 {
		log.Fatalf("[TestBase] error in update user request : %d : %s", resp.StatusCode, err)
	}

	sc.lastMessage = NotifyMessage{}
	query = map[string]any{"id": newUserIn.Id}
	resp, err = RufsRestRequest[RufsUser, RufsUser](&sc.httpRest, "/rest/rufs_user", http.MethodDelete, query, nil, nil)

	if err != nil || resp.StatusCode != http.StatusOK || sc.lastMessage.Action != "delete" {
		log.Fatalf("[TestBase] error in update user request : %d : %s", resp.StatusCode, err)
	}

	resp, err = RufsRestRequest[any, any](&sc.httpRest, "/manifest.json", http.MethodGet, nil, nil, nil)

	if err != nil || resp.StatusCode != http.StatusOK || !strings.HasPrefix(sc.httpRest.MessageWorking, "{") {
		log.Fatalf("[TestBase] error in get file 'index.html' : %d : %s", resp.StatusCode, err)
	}

	resp, err = RufsRestRequest[any, any](&sc.httpRest, "/es6/CaseConvert.js", http.MethodGet, nil, nil, nil)

	if err != nil || resp.StatusCode != http.StatusOK || !strings.HasPrefix(sc.httpRest.MessageWorking, "class CaseConvert") {
		log.Fatalf("[TestBase] error in get file 'index.html' : %d : %s", resp.StatusCode, err)
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
	service.Init(service)

	if err := service.Listen(); err != nil && err != http.ErrServerClosed {
		log.Fatal("[TestBase] Unexpected server closed !")
	}
}
