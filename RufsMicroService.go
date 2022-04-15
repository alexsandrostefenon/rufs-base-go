package rufsBase

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

type RufsGroupOwner struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Route struct {
	Path       string `json:"path"`
	Controller string `json:"controller"`
}

type MenuItem struct {
	Menu  string `json:"menu"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

type RufsUserProteced struct {
	Id             int            `json:"id"`
	Name           string         `json:"name"`
	RufsGroupOwner int            `json:"rufsGroupOwner"`
	Groups         []int          `json:"groups"`
	Roles          map[string]int `json:"roles"`
}

type RufsUserPublic struct {
	Routes []Route             `json:"routes"`
	Menu   map[string]MenuItem `json:"menu"`
	Path   string              `json:"path"`
}

type RufsUser struct {
	RufsUserProteced
	RufsUserPublic
	FullName string `json:"fullName"`
	Password string `json:"password"`
}

type TokenPayload struct {
	RufsUserProteced
	Ip string `json:"ip"`
}

type LoginResponse struct {
	TokenPayload
	RufsUserPublic
	JwtHeader string   `json:"tokenPayload"`
	Title     string   `json:"title"`
	Openapi   *OpenApi `json:"openapi"`
}

type RufsClaims struct {
	*jwt.StandardClaims
	TokenPayload
}

type EntityManager interface {
	Connect() error
	Find(tableName string, fields map[string]any, orderBy []string) ([]any, error)
	FindOne(tableName string, fields map[string]any) (map[string]any, error)
	Insert(tableName string, obj map[string]any) (map[string]any, error)
	Update(tableName string, key map[string]any, obj map[string]any) (map[string]any, error)
	DeleteOne(tableName string, key map[string]any) error
	GetOpenApi(openapi *OpenApi, options map[string]any) (*OpenApi, error)
}

func EntityManagerFind[T any](em EntityManager, name string) ([]T, error) {
	return nil, fmt.Errorf("Don't implemented")
}

type RufsMicroService struct {
	MicroServiceServer
	wsServerTokens         map[string]*RufsClaims
	checkRufsTables        bool
	requestBodyContentType string
	migrationPath          string
	dataStoreManager       *DataStoreManager
	entityManager          EntityManager
	fileDbAdapter          *FileDbAdapter
	openapi                *OpenApi
	listGroup              []interface{}
	listGroupUser          []interface{}
	listGroupOwner         []RufsGroupOwner
	listUser               []RufsUser
}

func (rms *RufsMicroService) Init(imss IMicroServiceServer) (err error) {
	rms.MicroServiceServer.Init(imss)
	rms.wsServerTokens = make(map[string]*RufsClaims)
	rms.checkRufsTables = true
	rms.requestBodyContentType = "application/json"

	if rms.appName == "" {
		rms.appName = "base"
	}

	if rms.migrationPath == "" {
		rms.migrationPath = `./rufs-${config.appName}-es6/sql`
	}

	if rms.openapi, err = rms.LoadOpenApi(); err != nil {
		return err
	} else {
		if err := json.Unmarshal([]byte(rufsMicroServiceOpenApi), rms.openapi); err != nil {
			UtilsShowJsonUnmarshalError(rufsMicroServiceOpenApi, err)
			return err
		}

		rms.openapi.FillOpenApi(FillOpenApiOptions{schemas: rms.openapi.Components.Schemas, requestBodyContentType: rms.requestBodyContentType})
		//		RequestFilter.updateRufsServices(rms.fileDbAdapter, openapi);
		rms.fileDbAdapter = &FileDbAdapter{fileTables: make(map[string][]any), openapi: rms.openapi}
		rms.entityManager = rms.fileDbAdapter //new DbClientPostgres(rms.config.dbConfig, {missingPrimaryKeys: rms.config.dbMissingPrimaryKeys, missingForeignKeys: rms.config.dbMissingForeignKeys, aliasMap: rms.config.aliasMap});
		rms.dataStoreManager = DataStoreManagerNew([]*Schema{}, rms.openapi)
	}

	return nil
}

func (rms *RufsMicroService) authenticateUser(userName string, userPassword string, remoteAddr string) (*LoginResponse, error) {
	time.Sleep(100 * time.Millisecond)
	rms.loadRufsTables()
	var user *RufsUser

	for _, element := range rms.listUser {
		if element.Name == userName {
			user = &element
			break
		}
	}

	if user == nil || user.Password != userPassword {
		return nil, errors.New("don't match user and password")
	}

	loginResponse := &LoginResponse{TokenPayload: TokenPayload{Ip: remoteAddr, RufsUserProteced: RufsUserProteced{Name: userName}}}
	loginResponse.Title = user.Name
	loginResponse.RufsGroupOwner = user.RufsGroupOwner
	loginResponse.Roles = user.Roles
	loginResponse.Routes = user.Routes
	loginResponse.Path = user.Path
	loginResponse.Menu = user.Menu
	/*
		if loginResponse.rufsGroupOwner > 0 {
			const item = rms.entityManager.dataStoreManager.getPrimaryKeyForeign("rufsUser", "rufsGroupOwner", user);
			const rufsGroupOwner = Filter.findOne(rms.listGroupOwner, item.primaryKey);
			if (rufsGroupOwner != null) loginResponse.title = rufsGroupOwner.name + " - " + userName;
		}

		Filter.find(rms.listGroupUser, {"rufsUser": user.id}).forEach(item => loginResponse.tokenPayload.groups.push(item.rufsGroup));
	*/
	return loginResponse, nil
}

func (rms *RufsMicroService) OnRequest(req *http.Request, resource string, action string) Response {
	if resource == "login" {
		/*
			getRolesMask := func(roles map[string]Role) map[string]int {
				ret := make(map[string]int)

				for schemaName, role := range roles {
					mask := 0
					methods := []string{"get", "post", "patch", "put", "delete", "query"}

					for i, method := range methods {
						if value, ok := role[method]; ok && value {
							mask |= 1 << i
						}
					}

					ret[schemaName] = mask
				}

				return ret
			}
		*/
		loginRequest := map[string]string{}
		err := json.NewDecoder(req.Body).Decode(&loginRequest)

		if err != nil {
			return ResponseUnauthorized(fmt.Sprint(err))
		}

		userName, ok := loginRequest["user"]

		if !ok {
			return ResponseBadRequest(fmt.Sprint("[RufsMicroService.OnRequest.login] missing field 'user'"))
		}

		password, ok := loginRequest["password"]

		if !ok {
			return ResponseBadRequest(fmt.Sprint("[RufsMicroService.OnRequest.login] missing field 'password'"))
		}

		if loginResponse, err := rms.authenticateUser(userName, password, req.RemoteAddr); err == nil {
			if userName == "admin" {
				loginResponse.Openapi = rms.dataStoreManager.openapi
			} else {
				return ResponseOk("TODO")
				/*
					loginResponse.openapi = OpenApi.create({});
					OpenApi.copy(loginResponse.openapi, rms.entityManager.dataStoreManager.openapi, roles)
					rms.storeOpenApi(loginResponse.openapi, `openapi-${userName}.json`)
				*/
			}

			token := jwt.New(jwt.SigningMethodHS256)
			token.Claims = &RufsClaims{&jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 60 * 8).Unix()}, loginResponse.TokenPayload}
			jwtSecret := os.Getenv("RUFS_JWT_SECRET")

			if jwtSecret == "" {
				jwtSecret = "123456"
			}

			loginResponse.JwtHeader, err = token.SignedString([]byte(jwtSecret))
			return ResponseOk(loginResponse)
		} else {
			return ResponseUnauthorized(fmt.Sprint(err))
		}
	} else {
		rf := RequestFilter{}
		rf.microService = rms

		if _, ok := rf.microService.fileDbAdapter.fileTables[rf.serviceName]; ok {
			rf.entityManager = rms.fileDbAdapter
		} else {
			rf.entityManager = rms.entityManager
		}

		if access, err := rf.CheckAuthorization(req, resource, action); err != nil {
			return ResponseBadRequest(fmt.Sprintf("[RufsMicroService.OnRequest.CheckAuthorization] : %s", err))
		} else if !access {
			return ResponseUnauthorized("Explicit Unauthorized")
		}
		/*
			if resource == "rufsService" && req.Method == http.MethodGet && action == "query" {
				const list = OpenApi.getList(Qs, OpenApi.convertRufsToStandart(rms.openapi, true), true, req.tokenPayload.roles)
				return Promise.resolve(ResponseOk(list))
			}
		*/
		return rf.ProcessRequest()
	}
}

func (rms *RufsMicroService) OnWsMessageFromClient(connection *websocket.Conn, tokenString string) {
	rms.MicroServiceServer.OnWsMessageFromClient(connection, tokenString)

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

	if claims, ok := token.Claims.(*RufsClaims); ok && token.Valid {
		rms.wsServerConnections[tokenString] = connection
		rms.wsServerTokens[tokenString] = claims
		log.Printf("[MicroServiceServer.onWsMessageFromClient] Ok")
	} else {
		fmt.Println(err)
	}
}

func loadTable[T any](rms *RufsMicroService, name string, defaultRows []T) ([]T, error) {
	var list []T
	var err error

	if list, err = EntityManagerFind[T](rms.entityManager, name); err != nil {
		if list, err = FileDbAdapterLoad[T](rms.fileDbAdapter, name); err == nil {
			if len(list) == 0 && len(defaultRows) > 0 {
				err = FileDbAdapterStore(rms.fileDbAdapter, name, defaultRows)
				list = defaultRows
			}
		} else {
			err = FileDbAdapterStore(rms.fileDbAdapter, name, defaultRows)
			list = defaultRows
		}
	}

	return list, err
}

func (rms *RufsMicroService) loadRufsTables() error {
	var emptyList []interface{}
	loadTable(rms, "rufsService", emptyList)

	if list, err := loadTable(rms, "rufsGroup", emptyList); err == nil {
		rms.listGroup = list
	} else {
		return err
	}

	if list, err := loadTable(rms, "rufsGroupUser", emptyList); err == nil {
		rms.listGroupUser = list
	} else {
		return err
	}

	if list, err := loadTable(rms, "rufsGroupOwner", defaultGroupOwnerAdmin); err == nil {
		rms.listGroupOwner = list
	} else {
		return err
	}

	if list, err := loadTable(rms, "rufsUser", defaultUserAdmin); err == nil {
		rms.listUser = list
	} else {
		return err
	}

	return nil
}

func UtilsShowJsonUnmarshalError(str string, err error) {
	if jsonError, ok := err.(*json.SyntaxError); ok {
		lineAndCharacter := func(input string, offset int) (line int, character int, err error) {
			lf := rune(0x0A)

			if offset > len(input) || offset < 0 {
				return 0, 0, fmt.Errorf("Couldn't find offset %d within the input.", offset)
			}

			// Humans tend to count from 1.
			line = 1

			for i, b := range input {
				if b == lf {
					line++
					character = 0
				}
				character++
				if i == offset {
					break
				}
			}

			return line, character, nil
		}

		line, character, lcErr := lineAndCharacter(str, int(jsonError.Offset))
		fmt.Fprintf(os.Stderr, "Cannot parse JSON schema due to a syntax error at line %d, character %d: %v\n", line, character, jsonError.Error())

		if lcErr != nil {
			fmt.Fprintf(os.Stderr, "Couldn't find the line and character position of the error due to error %v\n", lcErr)
		}
	}
}

/*
func (rms *RufsMicroService) expressEndPoint(req, res, next) {
	let promise;

	if (rms.fileDbAdapter == undefined) {
		promise = rms.loadRufsTables();
	} else {
		promise = Promise.resolve();
	}

	return promise.then(() => super.expressEndPoint(req, res, next));
}
*/
func (rms *RufsMicroService) Listen() (err error) {
	createRufsTables := func() error {
		if !rms.checkRufsTables {
			return nil
		}

		openapi, err := rms.entityManager.GetOpenApi(&OpenApi{}, map[string]any{})

		if err != nil {
			return err
		}

		tablesMissing := map[string]any{}
		openapiRufs := &OpenApi{}

		if err := json.Unmarshal([]byte(rufsMicroServiceOpenApi), openapiRufs); err != nil {
			UtilsShowJsonUnmarshalError(rufsMicroServiceOpenApi, err)
			return err
		}

		for name, schemaRufs := range openapiRufs.Components.Schemas {
			if _, ok := openapi.Components.Schemas[name]; !ok {
				tablesMissing[name] = schemaRufs
			}
		}
		/*
			const rufsServiceDbSync = new RufsServiceDbSync(rms.entityManager);

			const createTable = iterator => {
				let it = iterator.next();
				if (it.done == true) return Promise.resolve();
				let [name, schema] = it.value;
				console.log(`${rms.constructor.name}.listen().createRufsTables().createTable(${name})`);
				return rufsServiceDbSync.createTable(name, schema).then(() => createTable(iterator));
			};

			createTable(tablesMissing.entries());

			then(() => rms.entityManager.findOne("rufsGroupOwner", {name: "ADMIN"}).catch(() => rms.entityManager.insert("rufsGroupOwner", RufsMicroService.defaultGroupOwnerAdmin))).
			then(() => rms.entityManager.findOne("rufsUser", {name: "admin"}).catch(() => rms.entityManager.insert("rufsUser", RufsMicroService.defaultUserAdmin))).
		*/
		return nil
	}

	syncDb2OpenApi := func() error {
		/*
			const execMigrations = () => {
				if (fs.existsSync(rms.config.migrationPath) == false)
					return Promise.resolve();

				const regExp1 = /^(?<v1>\d{1,3})\.(?<v2>\d{1,3})\.(?<v3>\d{1,3})/;
				const regExp2 = /^(?<v1>\d{3})(?<v2>\d{3})(?<v3>\d{3})/;

				const getVersion = name => {
					const regExpResult = regExp1.exec(name);
					if (regExpResult == null) return 0;
					return Number.parseInt(regExpResult.groups.v1.padStart(3, "0") + regExpResult.groups.v2.padStart(3, "0") + regExpResult.groups.v3.padStart(3, "0"));
				};

				const migrate = (openapi, list) => {
					if (list.length == 0)
						return Promise.resolve(openapi);

					const fileName = list.shift();
					return fsPromises.readFile(`${rms.config.migrationPath}/${fileName}`, "utf8").
					then(text => {
						const execSql = list => {
							if (list.length == 0) return Promise.resolve();
							const sql = list.shift();
							return rms.entityManager.client.query(sql).
							catch(err => {
								console.error(`[${rms.constructor.name}.listen.syncDb2OpenApi.execMigrations.migrate(${fileName}).execSql] :\n${sql}\n${err.message}`);
								throw err;
							}).
							then(() => execSql(list));
						};

						const list = text.split("--split");
						return execSql(list);
					}).
					then(() => {
						let newVersion = getVersion(fileName);
						const regExpResult = regExp2.exec(newVersion.toString().padStart(9, "0"));
						openapi.info.version = `${Number.parseInt(regExpResult.groups.v1)}.${Number.parseInt(regExpResult.groups.v2)}.${Number.parseInt(regExpResult.groups.v3)}`;
						return rms.storeOpenApi(openapi);
					}).
					then(() => migrate(openapi, list));
				};

				return rms.loadOpenApi().
				then(openapi => {
					console.log(`[${rms.constructor.name}.syncDb2OpenApi()] openapi in execMigrations`);
					const oldVersion = getVersion(openapi.info.version);
					return fsPromises.readdir(`${rms.config.migrationPath}`).
					then(list => list.filter(fileName => getVersion(fileName) > oldVersion)).
					then(list => list.sort((a, b) => getVersion(a) - getVersion(b))).
					then(list => migrate(openapi, list));
				});
			};

			execMigrations().
		*/
		openApiDb, _ := rms.entityManager.GetOpenApi(&OpenApi{}, map[string]any{"requestBodyContentType": rms.requestBodyContentType})
		openApiDb.FillOpenApi(FillOpenApiOptions{schemas: rms.openapi.Components.Schemas, requestBodyContentType: rms.requestBodyContentType})

		for name, schemaDb := range openApiDb.Components.Schemas {
			openApiDb.Components.Schemas[name] = MergeSchemas(rms.openapi.Components.Schemas[name], schemaDb, false, name)
		}

		for name, dbSchema := range openApiDb.Components.Schemas {
			if _, ok := rms.openapi.Components.Schemas[name]; !ok {
				rms.openapi.Components.Schemas[name] = dbSchema

				if value, ok := openApiDb.Paths["/"+name]; ok {
					rms.openapi.Paths["/"+name] = value
				}

				if value, ok := openApiDb.Components.Parameters[name]; ok {
					rms.openapi.Components.Parameters[name] = value
				}

				if value, ok := openApiDb.Components.RequestBodies[name]; ok {
					rms.openapi.Components.RequestBodies[name] = value
				}

				if value, ok := openApiDb.Components.Responses[name]; ok {
					rms.openapi.Components.Responses[name] = value
				}
			}
		}

		return rms.StoreOpenApi(rms.openapi, "")
	}

	//console.log(`[${rms.constructor.name}] starting ${rms.config.appName}...`);
	if err := rms.entityManager.Connect(); err != nil {
		return err
	}

	if err := createRufsTables(); err != nil {
		return err
	}

	if err := syncDb2OpenApi(); err != nil {
		return err
	} else {
		//console.log(`[${rms.constructor.name}.listen()] openapi after syncDb2OpenApi`);
		if err := RequestFilterUpdateRufsServices(rms.entityManager, rms.openapi); err != nil {
			return err
		}

		if err := rms.MicroServiceServer.Listen(); err != nil {
			return err
		}
		//then(() => console.log(`[${rms.constructor.name}] ... ${rms.config.appName} started.`));
		return nil
	}
}

var rufsMicroServiceSchemaProperties string = `{
	"x-required":{"type": "boolean", "x-orderIndex": 1, "x-sortType": "asc"},
	"nullable":{"type": "boolean", "x-orderIndex": 2, "x-sortType": "asc"},
	"type":{"options": ["string", "integer", "boolean", "number", "date-time", "date", "time"]},
	"properties":{"type": "object", "properties": {}},
	"items":{"type": "object", "properties": {}},
	"maxLength":{"type": "integer"},
	"format":{},
	"pattern":{},
	"enum": {},
	"x-$ref":{},
	"x-enumLabels": {},
	"default":{},
	"example":{},
	"description":{}
}`

var rufsMicroServiceOpenApi string = `
{
	"components": {
		"schemas": {
			"rufsGroupOwner": {
				"properties": {
					"id":   {"type": "integer", "x-identityGeneration": "BY DEFAULT"},
					"name": {"nullable": false, "unique": true}
				},
				"x-primaryKeys": ["id"]
			},
			"rufsUser": {
				"properties": {
					"id":             {"type": "integer", "x-identityGeneration": "BY DEFAULT"},
					"rufsGroupOwner": {"type": "integer", "nullable": false, "x-$ref": "#/components/schemas/rufsGroupOwner"},
					"name":           {"maxLength": 32, "nullable": false, "unique": true},
					"password":       {"nullable": false},
					"path":           {},
					"roles":          {"type": "object", "properties": {}},
					"routes":         {"type": "array", "items": {"properties": {}}},
					"menu":           {"type": "object", "properties": {}}
				},
				"x-primaryKeys": ["id"],
				"x-uniqueKeys":  {}
			},
			"rufsGroup": {
				"properties": {
					"id":   {"type": "integer", "x-identityGeneration": "BY DEFAULT"},
					"name": {"nullable": false, "unique": true}
				},
				"x-primaryKeys": ["id"]
			},
			"rufsGroupUser": {
				"properties": {
					"rufsUser":  {"type": "integer", "nullable": false, "x-$ref": "#/components/schemas/rufsUser"},
					"rufsGroup": {"type": "integer", "nullable": false, "x-$ref": "#/components/schemas/rufsGroup"}
				},
				"x-primaryKeys": ["rufsUser", "rufsGroup"],
				"x-uniqueKeys":  {}
			},
			"rufsService": {
				"properties": {
					"operationId": {},
					"path":        {},
					"method":      {},
					"parameter":   {"type": "object", "properties": ` + rufsMicroServiceSchemaProperties + `},
					"requestBody": {"type": "object", "properties": ` + rufsMicroServiceSchemaProperties + `},
					"response":    {"type": "object", "properties": ` + rufsMicroServiceSchemaProperties + `}
				},
				"x-primaryKeys": ["operationId"],
				"x-uniqueKeys": {}
			}
		}
	}
}
`
var defaultGroupOwnerAdmin []RufsGroupOwner = []RufsGroupOwner{{Name: "ADMIN"}}

var defaultUserAdmin []RufsUser = []RufsUser{{
	RufsUserProteced: RufsUserProteced{
		Id: 1, Name: "admin", RufsGroupOwner: 1,
		Roles: map[string]int{
			"rufsGroupOwner": 0xff,
			"rufsUser":       0xff,
			"rufsGroup":      0xff,
			"rufsGroupUser":  0xff,
		},
	},
	RufsUserPublic: RufsUserPublic{
		Path: "rufs_user/search",
		Routes: []Route{
			{Path: "/app/rufs_service/:action", Controller: "OpenApiOperationObjectController"},
			{Path: "/app/rufs_user/:action", Controller: "UserController"},
		},
	},
	Password: "21232f297a57a5a743894a0e4a801fc3",
}}

/*
func (rms *RufsMicroService) LoadOpenApi() (*OpenApi, error) {
	openapi, err := rms.MicroServiceServer.LoadOpenApi()

	if err != nil {
		return nil, err
	}

	for name, schemaRufs := range rms.openapiRufs.Components.Schemas {
		if _, ok := openapi.Components.Schemas[name]; !ok {
			tablesMissing[name] = schemaRufs
		}
	}

	return openapi, err
}
*/
