package rufsBase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(obj)

	if err != nil {
		return ResponseCreate([]byte(fmt.Sprint(err)), 500)
	}

	return ResponseCreate(buffer.Bytes(), status)
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

type IMicroServiceServer interface {
	LoadOpenApi() error
	Listen() error
	Shutdown()
	OnRequest(req *http.Request) Response
	OnWsMessageFromClient(connection *websocket.Conn, tokenString string)
}

type MicroServiceServer struct {
	appName                string
	protocol               string
	port                   int
	addr                   string
	apiPath                string
	security               string
	requestBodyContentType string
	ServeStaticPaths       string
	openapiFileName        string
	openapi                *OpenApi
	wsServerConnections    map[string]*websocket.Conn
	httpServer             *http.Server
	Imss                   IMicroServiceServer
}

func (mss *MicroServiceServer) OnRequest(req *http.Request) Response {
	log.Printf("[MicroServiceServer.OnRequest] : %s", req.URL.Path)
	return ResponseOk("OnRequest")
}

func (mss *MicroServiceServer) Listen() error {
	mss.wsServerConnections = make(map[string]*websocket.Conn)
	serveStaticPaths := path.Join(path.Dir(reflect.TypeOf(mss).PkgPath()), "webapp")

	if mss.ServeStaticPaths == "" {
		mss.ServeStaticPaths = serveStaticPaths
	} else {
		mss.ServeStaticPaths += "," + serveStaticPaths
	}

	if mss.port == 0 {
		mss.port = 8080
	}

	if mss.apiPath == "" {
		mss.apiPath = "rest"
	}

	if mss.Imss == nil {
		mss.Imss = mss
	}

	mss.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", mss.addr, mss.port)}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		found := false
		name := req.RequestURI

		if strings.HasSuffix(name, "/") {
			name = name + "index.html"
		}

		for _, folder := range strings.Split(mss.ServeStaticPaths, ",") {
			absFolder, _ := filepath.Abs(folder)
			fileName := path.Join(absFolder, name)

			if fileInfo, err := os.Stat(fileName); err == nil && !fileInfo.IsDir() {
				http.ServeFile(res, req, fileName)
				found = true
				log.Printf("[MicroServiceServer.Init] served file : %s : %s : %s", folder, req.RequestURI, fileName)
				break
			}
		}

		if !found {
			log.Printf("[MicroServiceServer.HandleFunc] : searching file %s is not result", req.RequestURI)
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte{})
		}
	})

	http.HandleFunc("/"+mss.apiPath+"/", func(res http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadAll(req.Body)
		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
		log.Printf("authorization='%s';", req.Header.Get("Authorization"))
		log.Printf("curl -X '%s' %s -d '%s' -H \"Authorization: $authorization\";", req.Method, req.RequestURI, rdr1)
		req.Body = rdr2
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS, POST, DELETE")
		res.Header().Set("Access-Control-Allow-Headers", req.Header.Get("Access-Control-Request-Headers"))

		if req.Method == http.MethodOptions {
			fmt.Fprint(res, "Ok")
			return
		}

		ret := mss.Imss.OnRequest(req)
		res.Header().Set("Content-Type", ret.ContentType)
		//log.Printf("[HandleFunc] : ret.Body = %s", string(ret.Body))
		res.WriteHeader(ret.StatusCode)
		res.Write(ret.Body)
	})

	upgrader := websocket.Upgrader{}
	log.Printf("[MicroServiceServer.Init] : websocket")

	http.HandleFunc("/websocket", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[MicroServiceServer.HandleFunc] : received websocket request %s from %s", req.RequestURI, req.RemoteAddr)
		connection, err := upgrader.Upgrade(w, req, nil)

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

			mss.Imss.OnWsMessageFromClient(connection, string(message))
		}
	})

	log.Print("[MicroServiceServer.Listen]")
	return mss.httpServer.ListenAndServe()
}

func (mss *MicroServiceServer) LoadOpenApi() error {
	//if (fileName == null) fileName = this.constructor.getArg("openapi-file");
	//if (fileName == null) fileName = `openapi-${this.config.appName}.json`;
	if mss.openapiFileName == "" {
		mss.openapiFileName = fmt.Sprintf("openapi-%s.json", mss.appName)
	}
	//console.log(`[${this.constructor.name}.loadOpenApi()] loading ${fileName}`);
	if mss.security == "" {
		mss.security = "jwt"
	}

	if mss.openapi == nil {
		mss.openapi = &OpenApi{}
	}

	if data, err := ioutil.ReadFile(mss.openapiFileName); err == nil {
		if err = json.Unmarshal(data, mss.openapi); err != nil {
			//console.log(`[${this.constructor.name}.loadOpenApi()] : fail to parse file :`, err);
			UtilsShowJsonUnmarshalError(string(data), err)
			log.Fatalf("[MicroServiceServer.LoadOpenApi] : %s", err)
			OpenApiCreate(mss.openapi, mss.security)
		}
	} else {
		OpenApiCreate(mss.openapi, mss.security)
	}

	if len(mss.openapi.Servers) == 0 {
		mss.openapi.Servers = append(mss.openapi.Servers, &ServerObject{Url: fmt.Sprintf("%s://localhost:%d/%s", mss.protocol, mss.port, mss.apiPath)})
		mss.openapi.Servers = append(mss.openapi.Servers, &ServerObject{Url: fmt.Sprintf("%s://localhost:%d/%s/%s", mss.protocol, (mss.port/10)*10, mss.appName, mss.apiPath)})
	}

	mss.openapi.convertStandartToRufs()
	return nil
}

func (mss *MicroServiceServer) StoreOpenApi(fileName string) (err error) {
	if fileName == "" {
		fileName = fmt.Sprintf("openapi-%s.json", mss.appName)
	}

	if data, err := json.MarshalIndent(mss.openapi, "", "\t"); err != nil {
		log.Fatalf("[FileDbAdapterStore] : failt to marshal list before wrinting file %s : %s", fileName, err)
	} else if err = ioutil.WriteFile(fileName, data, fs.ModePerm); err != nil {
		log.Fatalf("[FileDbAdapterStore] : failt to write file %s : %s", fileName, err)
	}

	return err
}

func (mss *MicroServiceServer) OnWsMessageFromClient(connection *websocket.Conn, tokenString string) {
}

func (mss *MicroServiceServer) Shutdown() {
	mss.httpServer.Shutdown(context.Background())
}
