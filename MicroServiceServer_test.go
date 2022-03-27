package rufsBase

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestLogin(t *testing.T) {
	var service IMicroServiceServer = &RufsMicroService{}
	service.Init(service)
	serviceRunning := make(chan struct{})
	serviceDone := make(chan struct{})

	go func() {
		close(serviceRunning)
		fmt.Print("Listen...")

		if err := service.Listen(); err != nil && err != http.ErrServerClosed {
			fmt.Print("Unexpected server closed !")
		}

		fmt.Print("...Listen.")
		defer close(serviceDone)
	}()

	log.Printf("[TestLogin] wait serviceRunning")
	<-serviceRunning
	loginDataReq := User{Name: "admin", Password: "21232f297a57a5a743894a0e4a801fc3"}
	sc := ServerConnection{}
	resp, err := sc.Login("http://localhost:8080", "", "/rest/login", loginDataReq.Name, loginDataReq.Password)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("5 : %d : %s", resp.StatusCode, err)
	}

	log.Print(resp)
	time.Sleep(2000 * time.Millisecond)
	log.Printf("[TestLogin] service.Shutdown()")
	service.Shutdown()
	time.Sleep(2000 * time.Millisecond)
	log.Printf("[TestLogin] wait serviceDone")
	<-serviceDone
	log.Printf("[TestLogin] serviceDone")
}
