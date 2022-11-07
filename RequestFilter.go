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
type RequestFilter struct {
	microService  *RufsMicroService
	entityManager EntityManager
	//	req            *http.Request
	tokenPayload *TokenPayload
	path         string
	method       string
	schemaName   string
	parameters   map[string]any
	objIn        map[string]any
}

func RequestFilterInitialize(req *http.Request, rms *RufsMicroService) (*RequestFilter, error) {
	var err error
	rf := &RequestFilter{}
	rf.microService = rms
	rf.method = strings.ToLower(req.Method)

	if rf.method == "post" || rf.method == "put" || rf.method == "patch" {
		err := json.NewDecoder(req.Body).Decode(&rf.objIn)

		if err != nil {
			return nil, err
		}
	}

	if req.URL.RawQuery != "" {
		rf.parameters, err = qs.Unmarshal(req.URL.RawQuery)

		if err != nil {
			return nil, fmt.Errorf("[RequestFilter.Initialize] fail to parse url query parameters : %s", err)
		}
	} else {
		rf.parameters = map[string]any{}
	}

	uriPath := req.URL.Path

	if strings.HasPrefix(uriPath, "/"+rms.apiPath+"/") {
		uriPath = uriPath[len(rms.apiPath)+1:]
	}

	rf.path, err = rms.openapi.getPathParams(uriPath, rf.parameters)

	if rf.path == "" {
		return nil, fmt.Errorf("[RufsMicroService.Initialize] : missing path for %s", uriPath)
	}

	if rf.schemaName, err = rms.openapi.getSchemaName(rf.path, rf.method); err != nil {
		return nil, fmt.Errorf("[RufsMicroService.Initialize] : %s", err)
	}

	if _, ok := rf.microService.fileDbAdapter.fileTables[rf.schemaName]; ok {
		rf.entityManager = rms.fileDbAdapter
	} else {
		rf.entityManager = rms.entityManager
	}

	return rf, nil
}

// private to create,update,delete,read
func (rf *RequestFilter) checkObjectAccess(obj map[string]any) Response {
	if _, ok := rf.microService.openapi.getSchemaFromSchemas(rf.schemaName); !ok {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to find schema %s", rf.schemaName))
	}

	var response Response
	userRufsGroupOwner := rf.tokenPayload.RufsGroupOwner
	rufsGroupOwnerEntries, _ := rf.microService.openapi.getPropertiesWithRef(rf.schemaName, "#/components/schemas/rufsGroupOwner")

	if userRufsGroupOwner > 1 && len(rufsGroupOwnerEntries) > 0 {
		objRufsGroupOwner, _ := rf.microService.openapi.getPrimaryKeyForeign(rf.schemaName, "rufsGroupOwner", obj)

		if objRufsGroupOwner == nil {
			obj["rufsGroupOwner"] = userRufsGroupOwner
			objRufsGroupOwner.PrimaryKey["id"] = userRufsGroupOwner
		}

		if objRufsGroupOwner.PrimaryKey["id"].(int) == userRufsGroupOwner {
			rufsGroup, _ := rf.microService.openapi.getPrimaryKeyForeign(rf.schemaName, "rufsGroup", obj)

			if rufsGroup != nil {
				found := false

				for _, group := range rf.tokenPayload.Groups {
					if group == rufsGroup.PrimaryKey["id"].(int) {
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

	newObj, err := rf.entityManager.Insert(rf.schemaName, rf.objIn)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processCreate] : %s", err))
	}

	rf.notify(newObj, false)
	return ResponseOk(newObj)
}

func (rf *RequestFilter) getObject(useDocument bool) (map[string]any, error) {
	primaryKey, err := rf.parseQueryParameters()

	if err != nil {
		return nil, err
	}

	obj, err := rf.entityManager.FindOne(rf.schemaName, primaryKey)

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

	primaryKey, err := rf.parseQueryParameters()

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processUpdate] : %s", err))
	}

	newObj, err := rf.entityManager.Update(rf.schemaName, primaryKey, rf.objIn)

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

	primaryKey, err := rf.parseQueryParameters()

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processDelete] : %s", err))
	}

	err = rf.entityManager.DeleteOne(rf.schemaName, primaryKey)

	if err != nil {
		return ResponseInternalServerError(fmt.Sprintf("[RequestFilter.processDelete] : %s", err))
	}

	rf.notify(objDeleted, true)
	return ResponseOk(map[string]any{})
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

func (rf *RequestFilter) parseQueryParameters() (map[string]any, error) {
	// se não for admin, limita os resultados para as rufsGroup vinculadas a empresa do usuário
	/*
		const userRufsGroupOwner = tokenData.rufsGroupOwner;
		const rufsGroupOwnerEntries = entityManager.dataStoreManager.getPropertiesWithRef(serviceName, "rufsGroupOwner");
		const rufsGroupEntries = entityManager.dataStoreManager.getPropertiesWithRef(serviceName, "rufsGroup");

		if (userRufsGroupOwner > 1) {
			if (rufsGroupOwnerEntries.length > 0) queryParameters[rufsGroupOwnerEntries[0].fieldName] = userRufsGroupOwner;
			if (rufsGroupEntries.length > 0) queryParameters[rufsGroupEntries[0].fieldName] = tokenData.groups;
		}
	*/
	schema, err := rf.microService.openapi.getSchemaFromParameters(rf.path, rf.method)

	if err != nil {
		return nil, fmt.Errorf("[RequestFilter.processQuery] Fail to find schema from parameter of %s.%s : %s", rf.path, rf.method, err)
	}

	obj, err := rf.microService.openapi.copyFields(schema, rf.parameters, false, false, false)

	if err != nil {
		return nil, fmt.Errorf("[RequestFilter.processQuery] Fail to parse fields from parameter of %s.%s : %s", rf.path, rf.method, err)
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

		schema, err := getParameterSchema(rf.entityManager.openapi, "/"+rf.serviceName, rf.method, "primaryKey")
	*/
	schema, err := rf.microService.openapi.getSchemaFromParameters(rf.path, rf.method)

	if err != nil {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to find schema from parameter of %s.%s : %s", rf.path, rf.method, err))
	}

	//const fields = Object.entries(this.parameters).length > 0 ? this.parseQueryParameters() : null;
	fields, err := rf.parseQueryParameters()

	if err != nil {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to parser parameters of %s.%s : %s", rf.path, rf.method, err))
	}

	orderBy := []string{}

	for fieldName, field := range schema.Properties {
		dataType := field.Type

		if field.Format != "" {
			dataType = field.Format
		}

		if dataType == "integer" || strings.Contains(dataType, "date") || strings.Contains(dataType, "time") {
			orderBy = append(orderBy, fieldName+" desc")
		}
	}

	if list, err := rf.entityManager.Find(rf.schemaName, fields, orderBy); err != nil {
		return ResponseBadRequest(fmt.Sprintf("[RequestFilter.processQuery] Fail to find items of %s : %s", rf.schemaName, err))
	} else {
		return ResponseOk(list)
	}
}

func (rf *RequestFilter) CheckAuthorization(req *http.Request) (access bool, err error) {
	checkMask := func(mask int, method string) (ret bool) {
		if idx := slices.Index([]string{"get", "post", "patch", "put", "delete", "query"}, method); idx >= 0 {
			ret = mask&(1<<idx) != 0
		}

		return ret
	}

	extractTokenPayload := func(tokenRaw string) (*TokenPayload, error) {
		rufsClaims, err := RufsDecryptToken(tokenRaw)

		if err == nil {
			return &rufsClaims.TokenPayload, err
		} else {
			return nil, fmt.Errorf("Authorization token header invalid : %s", err)
		}
	}

	for _, securityItem := range rf.microService.openapi.Security {
		for securityName := range securityItem {
			if securityScheme, ok := rf.microService.openapi.Components.SecuritySchemes[securityName]; ok && rf.tokenPayload == nil {
				if securityScheme.Type == "http" && securityScheme.Scheme == "bearer" && securityScheme.BearerFormat == "JWT" {
					authorizationHeaderPrefix := "Bearer "
					tokenRaw := req.Header["Authorization"][0]

					if strings.HasPrefix(tokenRaw, authorizationHeaderPrefix) {
						tokenRaw = tokenRaw[len(authorizationHeaderPrefix):]

						if rf.tokenPayload, err = extractTokenPayload(tokenRaw); err != nil {
							return false, err
						}
					}
				} else if securityScheme.Type == "apiKey" {
					if securityScheme.In == "header" {
						for headerName, headerArray := range req.Header {
							if strings.ToLower(headerName) == strings.ToLower(securityScheme.Name) {
								tokenRaw := headerArray[0]

								if user, err := rf.microService.fileDbAdapter.FindOne("rufsUser", map[string]any{"password": tokenRaw}); err != nil || user == nil {
									return false, err
								} else {
									rf.tokenPayload = &TokenPayload{}
									buffer, _ := json.Marshal(user)
									json.Unmarshal(buffer, rf.tokenPayload)
								}

								break
							}
						}
					}
				}
			}
		}
	}

	if rf.tokenPayload == nil {
		return false, fmt.Errorf("[RequestFilter.CheckAuthorization] missing security scheme for %s.%s", rf.path, rf.method)
	}

	if idx := slices.IndexFunc(rf.tokenPayload.Roles, func(e Role) bool { return e.Path == rf.path }); idx >= 0 {
		if checkMask(rf.tokenPayload.Roles[idx].Mask, rf.method) {
			access = true
		}
	} else {
		err = fmt.Errorf("[RequestFilter.CheckAuthorization] missing service %s.%s in authorized roles", rf.path, rf.method)
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

func (rf *RequestFilter) ProcessRequest() Response {
	var err error
	var resp Response
	schemaResponse, _ := rf.microService.openapi.getSchema(rf.path, rf.method, "responseObject")

	if rf.method == "get" && schemaResponse != nil && schemaResponse.Type == "array" {
		resp = rf.processQuery()
	} else if rf.method == "post" {
		resp = rf.processCreate()
	} else if rf.method == "put" {
		resp = rf.processUpdate()
	} else if rf.method == "patch" {
		resp = rf.processPatch()
	} else if rf.method == "delete" {
		resp = rf.processDelete()
	} else if rf.method == "get" {
		resp = rf.processRead()
	} else {
		return ResponseInternalServerError(fmt.Sprintf("[RequsetFilter.ProcessRequest] : unknow route for %s", rf.path))
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
	schema, ok := rf.microService.openapi.getSchemaFromSchemas(rf.schemaName)

	if !ok {
		rf.microService.openapi.getSchemaFromSchemas(rf.schemaName)
		log.Panicf("[RequestFilter.notify] missing schema response for path %s.", rf.path)
		return
	}

	rf.parameters, _ = rf.microService.openapi.copyFields(schema, obj, false, false, true)
	msg := NotifyMessage{rf.schemaName, "notify", rf.parameters}

	if isRemove {
		msg.Action = "delete"
	}

	//	dataSend, _ := json.Marshal(msg)
	objRufsGroupOwner, objRufsGroupOwnerErr := rf.microService.openapi.getPrimaryKeyForeign(rf.schemaName, "rufsGroupOwner", obj)
	rufsGroup, rufsGroupErr := rf.microService.openapi.getPrimaryKeyForeign(rf.schemaName, "rufsGroup", obj)
	log.Printf("[RequestFilter.notify] broadcasting %s ...", msg)

	for tokenString, wsServerConnection := range rf.microService.wsServerConnections {
		tokenData := rf.microService.wsServerConnectionsTokens[tokenString]
		// enviar somente para os clients de "rufsGroupOwner"
		checkRufsGroupOwner := objRufsGroupOwner == nil

		if checkRufsGroupOwner == false && objRufsGroupOwnerErr == nil {
			if id, ok := objRufsGroupOwner.PrimaryKey["id"]; ok && id.(int64) == (int64)(tokenData.RufsGroupOwner) {
				checkRufsGroupOwner = true
			}
		}

		checkRufsGroup := rufsGroup == nil

		if checkRufsGroup == false && rufsGroupErr == nil {
			if id, ok := rufsGroup.PrimaryKey["id"]; ok && sort.SearchInts(tokenData.Groups, (int)(id.(int64))) >= 0 {
				checkRufsGroup = true
			}
		}
		// restrição de rufsGroup
		if tokenData.RufsGroupOwner == 1 || (checkRufsGroupOwner && checkRufsGroup) {
			if idx := slices.IndexFunc(tokenData.Roles, func(e Role) bool { return e.Path == rf.path }); idx >= 0 {
				if (tokenData.Roles[idx].Mask & 0x01) != 0 {
					log.Printf("[RequestFilter.notify] send to client %s", tokenData.Name)
					wsServerConnection.WriteJSON(msg)
				}
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
