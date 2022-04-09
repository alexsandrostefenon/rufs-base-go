package rufsBase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	//"github.com/hetiansu5/urlquery"
	"github.com/derekstavis/go-qs"
)

type HttpRestRequest struct {
	Url            string
	MessageWorking string
	MessageError   string
	Token          string
}

func (hrq *HttpRestRequest) Init(url string) {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	hrq.Url = url
}

func RufsRestRequest[T, U any](hrq *HttpRestRequest, path string, method string, params map[string]any, objSend *T, objReceive *U) (resp *http.Response, err error) {
	url := hrq.Url

	if !strings.HasSuffix(url, "/") && !strings.HasPrefix(path, "/") {
		url = url + "/"
	}

	url = url + path

	var bodyBuffer []byte
	contentType := "text/plain"

	if objSend != nil {
		contentType = "application/json"
		bodyBuffer, err = json.Marshal(objSend)

		if err != nil {
			return nil, fmt.Errorf("[RufsRestRequest] broken object to send : %s", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBuffer))

	if err != nil {
		return nil, fmt.Errorf("[RufsRestRequest] broken new request : %s", err)
	}

	if params != nil {
		req.URL.RawQuery, err = qs.Marshal(params)

		if err != nil {
			return nil, fmt.Errorf("[RufsRestRequest] broken url query parameters : %s", err)
		}
	}

	req.Header.Add("content-type", contentType)

	if hrq.Token != "" {
		req.Header.Add("Authorization", "Bearer "+hrq.Token)
	}

	hrq.MessageWorking = "Processing request to " + url
	hrq.MessageError = ""

	client := &http.Client{}
	resp, errReq := client.Do(req)

	if errReq != nil {
		hrq.MessageError = fmt.Sprint(errReq)
		return nil, errReq
	}

	hrq.MessageWorking = ""
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK && objReceive != nil {
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(objReceive)

		if err != nil {
			return nil, fmt.Errorf("[RufsRestRequest] broken object received : %s", err)
		}
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, fmt.Errorf("[RufsRestRequest] broken message received : %s", err)
		}

		hrq.MessageError = string(bodyBytes)
	}

	return resp, err
}

/*
type RufsService struct {
	//DataStoreItem
}

func (rs *RufsService) Init(name, schema, serverConnection, httpRest) {
		super(name, schema, [], serverConnection);
		sc.httpRest = httpRest;
        sc.serverConnection = serverConnection;
        sc.params = schema;
        let appName = schema.appName != undefined ? schema.appName : "crud";
        sc.path = CaseConvert.camelToUnderscore(name);
        sc.pathRest = appName + "/rest/" + sc.path;
	}

func (rs *RufsService) process(action, params) {
		return super.process(action, params).then(() => {
			if (action == "search") {
				if (params.filter != undefined || params.filterRangeMin != undefined || params.filterRangeMax != undefined) {
					return sc.queryRemote(params);
				}
			}

			return Promise.resolve();
		})
	}

func (rs *RufsService) request(path, method, params, objSend) {
        return sc.httpRest.request(sc.pathRest + "/" + path, method, params, objSend);
	}

func (rs *RufsService) get(primaryKey) {
		return sc.serverConnection.get(sc.name, primaryKey, false);
	}

func (rs *RufsService) save(itemSend) {
		let schema = OpenApi.getSchemaFromRequestBodies(sc.serverConnection.openapi, sc.name);

		if (schema == undefined)
			schema = this;

		const dataOut = OpenApi.copyFields(schema, itemSend);
    	return sc.httpRest.save(sc.pathRest, dataOut).
    	then(dataIn => {
    		return sc.updateList(dataIn);
    	});
	}

func (rs *RufsService) update(primaryKey, itemSend) {
        return sc.httpRest.update(sc.pathRest, primaryKey, OpenApi.copyFields(this, itemSend)).then(data => {
            let pos = sc.findPos(primaryKey);
        	return sc.updateList(data, pos, pos);
        });
	}

func (rs *RufsService) patch(itemSend) {
    	return sc.httpRest.patch(sc.pathRest, OpenApi.copyFields(this, itemSend)).then(data => sc.updateList(data));
	}

func (rs *RufsService) remove(primaryKey) {
        return sc.httpRest.remove(sc.pathRest, primaryKey);//.then(data => sc.serverConnection.removeInternal(sc.name, primaryKey));
	}

func (rs *RufsService) queryRemote(params) {
        return sc.httpRest.query(sc.pathRest + "/query", params).then(list => {
			for (let [fieldName, field] of Object.entries(sc.properties))
				if (field.type.includes("date") || field.type.includes("time"))
					list.forEach(item => item[fieldName] = new Date(item[fieldName]));
        	sc.list = list;
        	return list;
        });
	}

}
*/
type ServerConnection struct {
	// DataStoreManager
	httpRest      HttpRestRequest
	loginResponse LoginResponse
	webSocket     *websocket.Conn
	lastMessage   NotifyMessage
}

/*
func (sc *ServerConnection) Init() {
    	super();
    	sc.pathname = "";
		sc.remoteListeners = [];
	}

func (sc *ServerConnection) clearRemoteListeners() {
		sc.remoteListeners = [];
	}

func (sc *ServerConnection) addRemoteListener(listenerInstance) {
		sc.remoteListeners.push(listenerInstance);
	}

func (sc *ServerConnection) removeInternal(schemaName, primaryKey) {
		const ret =  super.removeInternal(schemaName, primaryKey);
		for (let listener of sc.remoteListeners) listener.onNotify(schemaName, primaryKey, "delete");
		return ret;
	}

func (sc *ServerConnection) get(schemaName, primaryKey, ignoreCache) {
		return super.get(schemaName, primaryKey, ignoreCache).
		then(res => {
			if (res != null && res != undefined) {
				res.isCache = true;
				return Promise.resolve(res);
			}

			const service = sc.getSchema(schemaName);
			if (service == null || service == undefined) return Promise.resolve(null);
			return sc.httpRest.get(service.pathRest, primaryKey).
			then(data => {
				if (data == null) return null;
				data.isCache = false;
				return service.cache(primaryKey, data);
			});
		});
	}
*/
func (sc *ServerConnection) webSocketConnect(path string) (err error) {
	// 'wss://localhost:8443/xxx/websocket'
	var url = sc.httpRest.Url

	if strings.HasPrefix(url, "https://") {
		url = "wss://" + url[8:]
	} else if strings.HasPrefix(url, "http://") {
		url = "ws://" + url[7:]
	}

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	url = url + path

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	url = url + "websocket"
	sc.webSocket, _, err = websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		log.Fatalf("[webSocketConnect] fail dial with %s : %s", url, err)
	}

	err = sc.webSocket.WriteMessage(websocket.TextMessage, []byte(sc.httpRest.Token))

	if err != nil {
		log.Println("write:", err)
		return
	}

	go func() {
		for {
			_, buffer, err := sc.webSocket.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			json.Unmarshal(buffer, &sc.lastMessage)
			log.Printf("recv: %s", sc.lastMessage)
			/*
				var item = JSON.parse(event.data);
				console.log("[ServerConnection] webSocketConnect : onMessage :", item);
				var service = sc.services[item.service];

				if (service != undefined) {
					if (item.action == "delete") {
						if (service.findOne(item.primaryKey) != null) {
							sc.removeInternal(item.service, item.primaryKey);
						} else {
							console.log("[ServerConnection] webSocketConnect : onMessage : delete : alread removed", item);
						}
					} else {
						sc.get(item.service, item.primaryKey, true).
						then(res => {
							for (let listener of sc.remoteListeners) listener.onNotify(item.service, item.primaryKey, item.action);
							return res;
						});
					}
				}
			*/
		}
	}()

	return err
}

func (sc *ServerConnection) Login(server string, path string, loginPath string, user string, password string) (resp *http.Response, err error) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	sc.httpRest.Init(server)
	resp, err = RufsRestRequest(&sc.httpRest, loginPath, http.MethodPost, nil, &RufsUser{Name: user, Password: password}, &sc.loginResponse)

	if err != nil || resp.StatusCode != http.StatusOK {
		return resp, err
	}

	sc.httpRest.Token = sc.loginResponse.JwtHeader
	/*
		const schemas = [];

		for (let [schemaName, schema] of Object.entries(loginResponse.openapi.components.schemas)) {
			if (schema.appName == undefined) schema.appName = path;
			const service = sc.services[schemaName] = new RufsServiceClass(schemaName, schema, this, sc.httpRest);
			const methods = ["get", "post", "patch", "put", "delete"];
			const servicePath = loginResponse.openapi.paths["/" + schemaName];
			service.access = {};

			for (let method of methods) {
				if (servicePath[method] != undefined)
					service.access[method] = true;
				else
					service.access[method] = false;
			}

			if (service.properties.rufsGroupOwner != undefined && sc.rufsGroupOwner != 1) service.properties.rufsGroupOwner.hiden = true;
			if (service.properties.rufsGroupOwner != undefined && service.properties.rufsGroupOwner.default == undefined) service.properties.rufsGroupOwner.default = sc.rufsGroupOwner;
			schemas.push(service);
		}

		sc.setSchemas(schemas, loginResponse.openapi);
		let listDependencies = [];

		for (let serviceName in sc.services) {
			if (listDependencies.includes(serviceName) == false) {
				listDependencies.push(serviceName);
				sc.getDependencies(serviceName, listDependencies);
			}
		}

		if (user == "admin") listDependencies = ["rufsUser", "rufsGroupOwner", "rufsGroup", "rufsGroupUser"];
		const listQueryRemote = [];

		for (let $ref of listDependencies) {
			const service = sc.getSchema($ref);

			if (service.access.get == true) {
				listQueryRemote.push(service);
			}
		}

		return new Promise((resolve, reject) => {
			var queryRemoteServices = () => {
				if (listQueryRemote.length > 0) {
					let service = listQueryRemote.shift();
					console.log("[ServerConnection] loading", service.label, "...");
					callbackPartial("loading... " + service.label);

					service.queryRemote(null).then(list => {
						console.log("[ServerConnection] ...loaded", service.label, list.length);
						queryRemoteServices();
					}).catch(error => reject(error));
				} else {
					console.log("[ServerConnection] ...loaded services");
					resolve(loginResponse);
				}
			}

			queryRemoteServices();
		});
	*/
	sc.webSocketConnect(path)
	return resp, err
}

func (sc *ServerConnection) logout() {
	//	sc.webSocket.close()
	sc.httpRest.Token = ""
	/*
		// limpa todos os dados da sessão anterior
		for (let serviceName in sc.services) {
			delete sc.services[serviceName];
		}
	*/
}
