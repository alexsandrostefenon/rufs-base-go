package rufsBase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

func ResponseCreate(body []byte, status int) Response {
	resp := Response{}
	resp.StatusCode = status
	resp.Body = body

	if status == 200 {
		resp.ContentType = "application/json"
	} else {
		resp.ContentType = "text"
	}

	//log.Printf("[ResponseCreate] : body = %s", string(body))
	return resp
}

func ResponseOk[T any](obj T) Response {
	status := http.StatusOK
	body, err := json.Marshal(obj)

	if err != nil {
		status = 500
		body = []byte(fmt.Sprint(err))
	}

	return ResponseCreate(body, status)
}

func ResponseUnauthorized(msg string) Response {
	return ResponseCreate([]byte(msg), 401)
}

func ResponseBadRequest(msg string) Response {
	return ResponseCreate([]byte(msg), 400)
}

func ResponseInternalServerError(msg string) Response {
	return ResponseCreate([]byte(msg), 500)
}

type MicroServiceServer struct {
	appName             string
	protocol            string
	port                int
	addr                string
	apiPath             string
	security            string
	serveStaticPaths    string
	wsServerConnections map[string]*websocket.Conn
	httpServer          *http.Server
}

type IMicroServiceServer interface {
	Init(imss IMicroServiceServer) error
	Listen() error
	Shutdown()
	OnRequest(req *http.Request, resource string, action string) Response
	OnWsMessageFromClient(connection *websocket.Conn, tokenString string)
}

func (mss *MicroServiceServer) Init(imss IMicroServiceServer) {
	mss.wsServerConnections = make(map[string]*websocket.Conn)
	serveStaticPaths := path.Join(path.Dir(reflect.TypeOf(mss).PkgPath()), "webapp")

	if mss.serveStaticPaths == "" {
		mss.serveStaticPaths = serveStaticPaths
	} else {
		mss.serveStaticPaths += "," + serveStaticPaths
	}

	if mss.port == 0 {
		mss.port = 8080
	}

	if mss.apiPath == "" {
		mss.apiPath = "rest"
	}

	mss.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", mss.addr, mss.port)}

	for _, path := range strings.Split(mss.serveStaticPaths, ",") {
		log.Printf("[MicroServiceServer.Init] serving static folder : %s", http.Dir(path))
		http.Handle("/", http.FileServer(http.Dir(path)))
	}

	http.HandleFunc("/"+mss.apiPath+"/", func(res http.ResponseWriter, req *http.Request) {
		log.Printf("[MicroServiceServer.ServeHTTP]")
		paths := strings.Split(req.URL.Path, "/")
		var resource string
		var action string

		if len(paths) >= 3 {
			resource = UnderscoreToCamel(paths[2], false)

			if len(paths) >= 4 {
				action = paths[3]
			}
		}

		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS, POST, DELETE")
		res.Header().Set("Access-Control-Allow-Headers", req.Header.Get("Access-Control-Request-Headers"))

		if req.Method == http.MethodOptions {
			fmt.Fprint(res, "Ok")
			return
		}

		ret := imss.OnRequest(req, resource, action)
		res.Header().Set("Content-Type", ret.ContentType)
		//log.Printf("[HandleFunc] : ret.Body = %s", string(ret.Body))
		res.WriteHeader(ret.StatusCode)
		res.Write(ret.Body)
	})

	upgrader := websocket.Upgrader{}
	log.Printf("[MicroServiceServer.Init] : websocket")

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[MicroServiceServer.Init]")
		connection, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		defer connection.Close()

		for {
			messageType, message, err := connection.ReadMessage()

			if err != nil {
				log.Println("read:", err)
				break
			}

			if messageType != 1 {
				log.Println("Invalid Message Type:", messageType)
				break
			}

			imss.OnWsMessageFromClient(connection, string(message))
		}
	})
}

func (mss *MicroServiceServer) OnRequest(req *http.Request, resource string, action string) Response {
	log.Printf("[MicroServiceServer.OnRequest] resource = %s - action = %s", resource, action)
	return ResponseOk("OnRequest")
}

func (mss *MicroServiceServer) Listen() error {
	return mss.httpServer.ListenAndServe()
}

func (mss *MicroServiceServer) LoadOpenApi() (*OpenApi, error) {
	//if (fileName == null) fileName = this.constructor.getArg("openapi-file");
	//if (fileName == null) fileName = `openapi-${this.config.appName}.json`;
	fileName := fmt.Sprintf("openapi-%s.json", mss.appName)
	//console.log(`[${this.constructor.name}.loadOpenApi()] loading ${fileName}`);
	openapi := &OpenApi{}

	if data, err := ioutil.ReadFile(fileName); err == nil {
		if err = json.Unmarshal(data, &openapi); err != nil {
			//console.log(`[${this.constructor.name}.loadOpenApi()] : fail to parse file :`, err);
			OpenApiCreate(openapi, mss.security)
		}
	} else {
		OpenApiCreate(openapi, mss.security)
	}

	if len(openapi.Servers) == 0 {
		openapi.Servers = append(openapi.Servers, &OpenApiServerComponent{Url: fmt.Sprintf("%s://localhost:%d/%s", mss.protocol, mss.port, mss.apiPath)})
		openapi.Servers = append(openapi.Servers, &OpenApiServerComponent{Url: fmt.Sprintf("%s://localhost:%d/%s/%s", mss.protocol, (mss.port/10)*10, mss.appName, mss.apiPath)})
	}

	openapi.convertStandartToRufs()
	return openapi, nil
}

func (mss *MicroServiceServer) OnWsMessageFromClient(connection *websocket.Conn, tokenString string) {
}

func (mss *MicroServiceServer) Shutdown() {
	mss.httpServer.Shutdown(context.Background())
}
