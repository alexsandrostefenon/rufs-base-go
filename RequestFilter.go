package rufsBase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/derekstavis/go-qs"
	"github.com/golang-jwt/jwt"
	"golang.org/x/exp/slices"
)

type RequestFilter struct {
	microService    *RufsMicroService
	entityManager   EntityManager
	req             *http.Request
	tokenPayload    *TokenPayload
	serviceName     string
	uriPath         string
	queryParameters map[string]any
	objIn           map[string]any
	primaryKey      map[string]any
}

/*
type DataStoreManagerDb struct {
	// DataStoreManager
	entityManager EntityManager
}

func (dsmd *DataStoreManagerDb) Init(listService, openapi, entityManager) {
	super(listService, openapi);
	this.entityManager = entityManager;
}

func (dsmd *DataStoreManagerDb) Get(schemaName, primaryKey, ignoreCache) {
	return super.get(schemaName, primaryKey, ignoreCache).
	then(res => {
		if (res != null && res != undefined) return Promise.resolve(res);
		const service = this.getSchema(schemaName);
		if (service == null || service == undefined) return Promise.resolve(null);
		return this.entityManager.findOne(schemaName, primaryKey).then(data => service.cache(primaryKey, data));
	});
}
*/

// private to create,update,delete,read
func (rf *RequestFilter) checkObjectAccess(obj map[string]any) Response {
	_, ok := rf.microService.dataStoreManager.openapi.getSchemaFromSchemas(rf.serviceName)

	if !ok {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to find schema of %s", rf.serviceName))
	}

	var response Response
	userRufsGroupOwner := rf.tokenPayload.RufsGroupOwner
	rufsGroupOwnerEntries, _ := rf.microService.dataStoreManager.openapi.getForeignKeyEntries(rf.serviceName, "#/components/schemas/rufsGroupOwner")

	if userRufsGroupOwner > 1 && len(rufsGroupOwnerEntries) > 0 {
		objRufsGroupOwner, _ := rf.microService.dataStoreManager.openapi.getPrimaryKeyForeign(rf.serviceName, "rufsGroupOwner", obj)

		if objRufsGroupOwner == nil {
			obj["rufsGroupOwner"] = userRufsGroupOwner
			objRufsGroupOwner.PrimaryKey["id"] = userRufsGroupOwner
		}

		if objRufsGroupOwner.PrimaryKey["id"] == userRufsGroupOwner {
			rufsGroup, _ := rf.microService.dataStoreManager.openapi.getPrimaryKeyForeign(rf.serviceName, "rufsGroup", obj)

			if rufsGroup != nil {
				found := false

				for _, group := range rf.tokenPayload.Groups {
					if group == rufsGroup.PrimaryKey["id"] {
						found = true
						break
					}
				}

				if !found {
					response = ResponseUnauthorized("unauthorized object rufsGroup")
				}
			}
		} else {
			response = ResponseUnauthorized("unauthorized object rufsGroupOwner")
		}
	}

	return response
}

func (rf *RequestFilter) processCreate() Response {
	response := rf.checkObjectAccess(rf.objIn)

	if response.StatusCode != 0 {
		return response
	}

	newObj, err := rf.entityManager.Insert(rf.serviceName, rf.objIn)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processCreate] : %s", err))
	}

	rf.notify(newObj, false)
	return ResponseOk(newObj)
}

func (rf *RequestFilter) getObject(useDocument bool) (map[string]any, error) {
	primaryKey, err := rf.parseQueryParameters(true)

	if err != nil {
		return nil, err
	}

	obj, err := rf.entityManager.FindOne(rf.serviceName, primaryKey)

	if err != nil {
		return nil, err
	}

	if useDocument != true {
		return obj, nil
	}
	/*
		const service = this.getSchema(entityManager, tokenData, serviceName)
		return entityManager.dataStoreManager.getDocument(service, obj, true, tokenData)
	*/
	return nil, fmt.Errorf("[ResquestFilter.GetObject] don't implemented com rf useDocument == true")
}

func (rf *RequestFilter) processRead() Response {
	// TODO : check RequestBody schema
	useDocument := false
	obj, err := rf.getObject(useDocument)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processRead] : %s", err))
	}

	return ResponseOk(obj)
}

func (rf *RequestFilter) processUpdate() Response {
	if _, err := rf.getObject(false); err != nil {
		return ResponseUnauthorized(fmt.Sprintf("[RequestFilter.processUpdate] err : %s", err))
	}

	response := rf.checkObjectAccess(rf.objIn)

	if response.StatusCode != 0 {
		return response
	}

	primaryKey, err := rf.parseQueryParameters(true)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processUpdate] : %s", err))
	}

	newObj, err := rf.entityManager.Update(rf.serviceName, primaryKey, rf.objIn)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processUpdate] : %s", err))
	}

	rf.notify(newObj, false)
	return ResponseOk(newObj)
}

func (rf *RequestFilter) processDelete() Response {
	objDeleted, err := rf.getObject(false)

	if err != nil {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processDelete] don't find register with informed parameters : %s", err))
	}

	primaryKey, err := rf.parseQueryParameters(true)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processDelete] : %s", err))
	}

	err = rf.entityManager.DeleteOne(rf.serviceName, primaryKey)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processDelete] : %s", err))
	}

	rf.notify(objDeleted, true)
	return ResponseOk(objDeleted)
}

func (rf *RequestFilter) processPatch() Response {
	return ResponseInternalServerError("TODO")
	/*
		const service = RequestFilter.getSchema(entityManager, user, serviceName);

		const process = keys => {
			if (keys.length > 0) {
				return entityManager.findOne(serviceName, keys.pop()).catch(() => process(keys));
			} else {
				return Promise.resolve(null);
			}
		};

		return process(service.getKeys(obj)).then(foundObj => {
			if (foundObj != null) {
				const primaryKey = service.getPrimaryKey(foundObj);
				return RequestFilter.processUpdate(user, primaryKey, entityManager, serviceName, obj, microService);
			} else {
				return RequestFilter.processCreate(user, entityManager, serviceName, obj, microService);
			}
		});
	*/
}

func (rf *RequestFilter) parseQueryParameters(onlyPrimaryKey bool) (map[string]any, error) {
	// se não for admin, limita os resultados para as rufsGroup vinculadas a empresa do usuário
	/*
		const userRufsGroupOwner = tokenData.rufsGroupOwner;
		const rufsGroupOwnerEntries = entityManager.dataStoreManager.getForeignKeyEntries(serviceName, "rufsGroupOwner");
		const rufsGroupEntries = entityManager.dataStoreManager.getForeignKeyEntries(serviceName, "rufsGroup");

		if (userRufsGroupOwner > 1) {
			if (rufsGroupOwnerEntries.length > 0) queryParameters[rufsGroupOwnerEntries[0].fieldName] = userRufsGroupOwner;
			if (rufsGroupEntries.length > 0) queryParameters[rufsGroupEntries[0].fieldName] = tokenData.groups;
		}
	*/
	schema, err := rf.microService.dataStoreManager.openapi.getSchemaFromParameters(rf.serviceName)

	if err != nil {
		return nil, fmt.Errorf("[RequestFilter.processQuery] Fail to find schema from parameter of %s : %s", rf.serviceName, err)
	}

	obj, err := rf.microService.dataStoreManager.openapi.copyFields(schema, rf.queryParameters, false, false)

	if err != nil {
		return nil, fmt.Errorf("[RequestFilter.processQuery] Fail to parse fields from parameter of %s : %s", rf.serviceName, err)
	}

	if onlyPrimaryKey {
		return obj, nil
	}
	/*
		if (queryParameters.filter != undefined) ret.filter = OpenApi.copyFields(schema, rf.queryParameters.filter);
		if (queryParameters.filterRangeMin != undefined) ret.filterRangeMin = OpenApi.copyFields(schema, rf.queryParameters.filterRangeMin);
		if (queryParameters.filterRangeMax != undefined) ret.filterRangeMax = OpenApi.copyFields(schema, rf.queryParameters.filterRangeMax);
	*/
	return obj, nil
}

func (rf *RequestFilter) processQuery() Response {
	/*
		getParameterSchema := func(openapi *OpenApi, path string, method string, parameterName string) (*Schema, error) {
			mapOperationObject, ok := openapi.Paths[path]

			if !ok {
				return nil, fmt.Errorf("[DataStoreManager.getParameterSchema] Missing service : %s", rf.serviceName)
			}

			operationObject, ok := mapOperationObject[strings.ToLower(method)]

			for _, parameterObject := range operationObject.Parameters {
				if parameterObject.Name == parameterName {
					return openapi.getSchema(parameterObject.Ref)
				}
			}

			return nil, fmt.Errorf("[OpenApi.getParameterSchema] missing parameter %s of method %s in path %s", parameterName, method, path)
		}

		schema, err := getParameterSchema(rf.entityManager.dataStoreManager.openapi, "/"+rf.serviceName, rf.req.Method, "primaryKey")
	*/
	schema, err := rf.microService.dataStoreManager.openapi.getSchemaFromParameters(rf.serviceName)

	if err != nil {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to find schema from parameter of %s : %s", rf.serviceName, err))
	}

	fields := make(map[string]any) //rf.parseQueryParameters(entityManager, tokenData, serviceName, queryParams) : null;
	orderBy := []string{}

	for fieldName, field := range schema.Properties {
		if field.Type == "integer" || strings.Contains(field.Type, "date") || strings.Contains(field.Type, "time") {
			orderBy = append(orderBy, fieldName+" desc")
		}
	}

	if list, err := rf.entityManager.Find(rf.serviceName, fields, orderBy); err != nil {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to find items of %s : %s", rf.serviceName, err))
	} else {
		return ResponseOk(list)
	}
}

func (rf *RequestFilter) CheckAuthorization(req *http.Request, serviceName string, uriPath string) (access bool, err error) {
	checkMask := func(mask int, method string) (ret bool) {
		if idx := slices.Index([]string{"get", "post", "patch", "put", "delete", "query"}, method); idx >= 0 {
			ret = mask&(1<<idx) != 0
		}

		return ret
	}

	if rf.tokenPayload, err = rf.extractTokenPayload(req.Header["Authorization"][0]); err != nil {
		return false, err
	}

	if mask, ok := rf.tokenPayload.Roles[serviceName]; ok {
		if uriPath == "" {
			uriPath = strings.ToLower(req.Method)
		}

		if checkMask(mask, uriPath) {
			access = true
			rf.uriPath = uriPath
			rf.req = req
			rf.serviceName = serviceName
		}
	} else {
		err = fmt.Errorf("[RequestFilter.CheckAuthorization] missing service %s in authorized roles", serviceName)
	}

	return access, err
}

func RufsDecryptToken(tokenString string) (*RufsClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RufsClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RufsClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func (rf *RequestFilter) extractTokenPayload(authorizationHeader string) (*TokenPayload, error) {
	authorizationHeaderPrefix := "Bearer "

	if strings.HasPrefix(authorizationHeader, authorizationHeaderPrefix) {
		token := authorizationHeader[len(authorizationHeaderPrefix):]
		rufsClaims, err := RufsDecryptToken(token)

		if err == nil {
			return &rufsClaims.TokenPayload, err
		} else {
			return nil, fmt.Errorf("Authorization token header invalid : %s", err)
		}
	} else {
		return nil, fmt.Errorf("Authorization token header invalid")
	}
}

func (rf *RequestFilter) ProcessRequest() Response {
	var err error
	method := rf.req.Method

	if rf.req.URL.RawQuery != "" {
		rf.queryParameters, err = qs.Unmarshal(rf.req.URL.RawQuery)

		if err != nil {
			return ResponseBadRequest(fmt.Sprintf("[RequestFilter.ProcessRequest] fail to parse url query parameters : %s", err))
		}
	}

	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		err := json.NewDecoder(rf.req.Body).Decode(&rf.objIn)

		if err != nil {
			return ResponseUnauthorized(fmt.Sprint(err))
		}
	}

	var resp Response
	if method == http.MethodGet && rf.uriPath == "query" {
		resp = rf.processQuery()
	} else if method == "POST" {
		resp = rf.processCreate()
	} else if method == "PUT" {
		resp = rf.processUpdate()
	} else if method == "PATCH" {
		resp = rf.processPatch()
	} else if method == "DELETE" {
		resp = rf.processDelete()
	} else if method == "GET" {
		resp = rf.processRead()
	} else {
		return ResponseInternalServerError(fmt.Sprintf("[RequsetFilter.ProcessRequest] : unknow route for method %s and action %s", method, rf.uriPath))
	}

	if err != nil {
		log.Printf("ProcessRequest error : %s", err)
		return ResponseInternalServerError(err.Error())
	} else {
		return resp
	}
}

type NotifyMessage struct {
	Service    string         `json:"service"`
	Action     string         `json:"action"`
	PrimaryKey map[string]any `json:"primaryKey"`
}

func (rf *RequestFilter) notify(obj map[string]any, isRemove bool) {
	msg := NotifyMessage{rf.serviceName, "notify", rf.primaryKey}

	if isRemove {
		msg.Action = "delete"
	}

	//	dataSend, _ := json.Marshal(msg)
	objRufsGroupOwner, _ := rf.microService.dataStoreManager.openapi.getPrimaryKeyForeign(rf.serviceName, "rufsGroupOwner", obj)
	rufsGroup, _ := rf.microService.dataStoreManager.openapi.getPrimaryKeyForeign(rf.serviceName, "rufsGroup", obj)
	log.Printf("[RequestFilter.notify] broadcasting %s ...", msg)

	for tokenString, wsServerConnection := range rf.microService.wsServerConnections {
		tokenData := rf.microService.wsServerTokens[tokenString]
		// enviar somente para os clients de "rufsGroupOwner"
		checkRufsGroupOwner := objRufsGroupOwner == nil || objRufsGroupOwner.PrimaryKey["id"] == tokenData.RufsGroupOwner
		checkRufsGroup := rufsGroup == nil || sort.SearchInts(tokenData.Groups, rufsGroup.PrimaryKey["id"].(int)) >= 0
		// restrição de rufsGroup
		if tokenData.RufsGroupOwner == 1 || (checkRufsGroupOwner && checkRufsGroup) {
			role := tokenData.Roles[rf.serviceName]
			// check get permission
			if (role & 1) != 0 {
				log.Printf("[RequestFilter.notify] send to client %s", tokenData.Name)
				wsServerConnection.WriteJSON(msg)
			}
		}
	}
}

func RequestFilterUpdateRufsServices(entityManager EntityManager, openapi *OpenApi) error {
	//console.log(`RequestFilter.updateRufsServices() : reseting entityManager.dataStoreManager`);
	listDataStore := []*DataStore{}
	// TODO : trocar openapi.components.schemas por openapi.paths
	for name, schema := range openapi.Components.Schemas {
		//log.Printf("[RequestFilterUpdateRufsServices] %s, %t", name, schema != nil)
		listDataStore = append(listDataStore, &DataStore{name, schema})
	}

	//log.Print(listDataStore)
	//	entityManager.dataStoreManager = new DataStoreManagerDb(listDataStore, openapi, entityManager);
	return nil
}
