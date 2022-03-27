package rufsBase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt"
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

	log.Printf("[ResponseCreate] : body = %s", string(body))
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
	port                int
	addr                string
	apiPath             string
	serveStaticPaths    string
	wsServerConnections map[string]*websocket.Conn
	wsServerTokens      map[string]jwt.MapClaims
	httpServer          *http.Server
}

type IMicroServiceServer interface {
	Init(imss IMicroServiceServer)
	Listen() error
	Shutdown()
	OnRequest(req *http.Request, resource string, action string) Response
}

func (mss *MicroServiceServer) Init(imss IMicroServiceServer) {
	mss.wsServerConnections = make(map[string]*websocket.Conn)
	mss.wsServerTokens = make(map[string]jwt.MapClaims)
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
		paths := strings.Split(req.RequestURI, "/")
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

		if req.Method == "OPTIONS" {
			fmt.Fprint(res, "Ok")
			return
		}

		ret := imss.OnRequest(req, resource, action)
		res.Header().Set("Content-Type", ret.ContentType)
		log.Printf("[HandleFunc] : ret.Body = %s", string(ret.Body))
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

			mss.onWsMessageFromClient(connection, string(message))
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

func (mss *MicroServiceServer) onWsMessageFromClient(connection *websocket.Conn, tokenString string) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		jwtSecret := os.Getenv("RUFS_JWT_SECRET")

		if jwtSecret == "" {
			jwtSecret = "123456"
		}

		hmacSampleSecret := []byte(jwtSecret)
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		name := claims["name"].(string)
		mss.wsServerConnections[name] = connection
		mss.wsServerTokens[name] = claims
		log.Printf("[MicroServiceServer.onWsMessageFromClient] Ok")
	} else {
		fmt.Println(err)
	}
}

func (mss *MicroServiceServer) Shutdown() {
	mss.httpServer.Shutdown(context.Background())
}
