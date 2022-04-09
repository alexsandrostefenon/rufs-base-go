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

type Role map[string]bool

type RufsGroupOwner struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Route struct {
	Path       string
	Controller string
}

type RufsUser struct {
	Id             int             `json:"id"`
	Name           string          `json:"name"`
	Password       string          `json:"password"`
	RufsGroupOwner int             `json:"rufsGroupOwner"`
	FullName       string          `json:"fullName"`
	Routes         []Route         `json:"routes"`
	Path           string          `json:"path"`
	Menu           string          `json:"menu"`
	Roles          map[string]Role `json:"roles"`
}

type TokenPayload struct {
	Ip             string         `json:"ip"`
	Name           string         `json:"name"`
	RufsGroupOwner int            `json:"rufsGroupOwner"`
	Groups         []int          `json:"groups"`
	Roles          map[string]int `json:"roles"`
}

type LoginResponse struct {
	TokenPayload
	JwtHeader string  `json:"jwtHeader"`
	Title     string  `json:"title"`
	Routes    []Route `json:"routes"`
	Path      string  `json:"path"`
	Menu      string  `json:"menu"`
	Openapi   string  `json:"openapi"`
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
}

func EntityManagerFind[T any](em EntityManager, name string) ([]T, error) {
	return nil, fmt.Errorf("Don't implemented")
}

type RufsMicroService struct {
	MicroServiceServer
	wsServerTokens map[string]*RufsClaims
	//	checkRufsTables bool
	migrationPath    string
	dataStoreManager *DataStoreManager
	entityManager    EntityManager
	fileDbAdapter    *FileDbAdapter
	listGroup        []interface{}
	listGroupUser    []interface{}
	listGroupOwner   []RufsGroupOwner
	listUser         []RufsUser
}

var defaultGroupOwnerAdmin []RufsGroupOwner = []RufsGroupOwner{{Name: "ADMIN"}}

var defaultUserAdmin []RufsUser = []RufsUser{{
	Id: 1, Name: "admin", RufsGroupOwner: 1, Password: "21232f297a57a5a743894a0e4a801fc3", Path: "rufs_user/search",
	Roles:  map[string]Role{"rufsGroupOwner": {"get": true, "query": true, "post": true, "put": true, "delete": true}, "rufsUser": {"get": true, "query": true, "post": true, "put": true, "delete": true}, "rufsGroup": {"get": true, "query": true, "post": true, "put": true, "delete": true}, "rufsGroupUser": {"get": true, "query": true, "post": true, "put": true, "delete": true}},
	Routes: []Route{{Path: "/app/rufs_service/:action", Controller: "OpenApiOperationObjectController"}, {Path: "/app/rufs_user/:action", Controller: "UserController"}},
}}

func (rms *RufsMicroService) Init(imss IMicroServiceServer) error {
	rms.MicroServiceServer.Init(imss)
	rms.wsServerTokens = make(map[string]*RufsClaims)

	if rms.appName == "" {
		rms.appName = "base"
	}

	if rms.migrationPath == "" {
		rms.migrationPath = `./rufs-${config.appName}-es6/sql`
	}

	if openapi, err := rms.LoadOpenApi(); err != nil {
		return err
	} else {
		//		RequestFilter.updateRufsServices(rms.fileDbAdapter, openapi);
		rms.fileDbAdapter = &FileDbAdapter{fileTables: make(map[string][]any), openapi: openapi}
		rms.entityManager = rms.fileDbAdapter //new DbClientPostgres(rms.config.dbConfig, {missingPrimaryKeys: rms.config.dbMissingPrimaryKeys, missingForeignKeys: rms.config.dbMissingForeignKeys, aliasMap: rms.config.aliasMap});
		rms.dataStoreManager = DataStoreManagerNew([]*Schema{}, openapi)
	}

	return nil
}

func (rms *RufsMicroService) authenticateUser(userName string, userPassword string, loginResponse *LoginResponse) (map[string]Role, error) {
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

	loginResponse.Title = user.Name
	loginResponse.RufsGroupOwner = user.RufsGroupOwner
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
	return user.Roles, nil
}

func (rms *RufsMicroService) OnRequest(req *http.Request, resource string, action string) Response {
	if resource == "login" {
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

		var user RufsUser
		err := json.NewDecoder(req.Body).Decode(&user)

		if err != nil {
			return ResponseUnauthorized(fmt.Sprint(err))
		}

		loginResponse := LoginResponse{TokenPayload: TokenPayload{Ip: req.RemoteAddr, Name: user.Name}}

		if roles, err := rms.authenticateUser(user.Name, user.Password, &loginResponse); err == nil {
			if user.Name == "admin" {
				//				loginResponse.Openapi = rms.entityManager.dataStoreManager.openapi
			} else {
				return ResponseOk("TODO")
				/*
					loginResponse.openapi = OpenApi.create({});
					OpenApi.copy(loginResponse.openapi, rms.entityManager.dataStoreManager.openapi, roles)
					rms.storeOpenApi(loginResponse.openapi, `openapi-${userName}.json`)
				*/
			}

			loginResponse.Roles = getRolesMask(roles)
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
				FileDbAdapterStore(rms.fileDbAdapter, name, defaultRows)
				list = defaultRows
			}
		} else {
			FileDbAdapterStore(rms.fileDbAdapter, name, defaultRows)
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
		/*
			if (rms.config.checkRufsTables != true)
				return Promise.resolve();

			return rms.entityManager.getOpenApi().
			then(openapi => {
				let tablesMissing = new Map();
				for (let name in RufsMicroService.openApiRufs.components.schemas) if (openapi.components.schemas[name] == undefined) tablesMissing.set(name, RufsMicroService.openApiRufs.components.schemas[name]);
				const rufsServiceDbSync = new RufsServiceDbSync(rms.entityManager);

				const createTable = iterator => {
					let it = iterator.next();
					if (it.done == true) return Promise.resolve();
					let [name, schema] = it.value;
					console.log(`${rms.constructor.name}.listen().createRufsTables().createTable(${name})`);
					return rufsServiceDbSync.createTable(name, schema).then(() => createTable(iterator));
				};

				return createTable(tablesMissing.entries());
			}).
			then(() => {
				return Promise.resolve().
				then(() => rms.entityManager.findOne("rufsGroupOwner", {name: "ADMIN"}).catch(() => rms.entityManager.insert("rufsGroupOwner", RufsMicroService.defaultGroupOwnerAdmin))).
				then(() => rms.entityManager.findOne("rufsUser", {name: "admin"}).catch(() => rms.entityManager.insert("rufsUser", RufsMicroService.defaultUserAdmin))).
				then(() => Promise.resolve());
			});
		*/
		return nil
	}

	syncDb2OpenApi := func() (*OpenApi, error) {
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

			return execMigrations().
			then(() => rms.entityManager.getOpenApi({}, {requestBodyContentType: rms.config.requestBodyContentType})).
			then(openApiDb => {
				return rms.loadOpenApi().
				then(openapi => {
					console.log(`[${rms.constructor.name}.syncDb2OpenApi()] openapi after execMigrations`);
					OpenApi.fillOpenApi(openApiDb, {schemas: RufsMicroService.openApiRufs.components.schemas, requestBodyContentType: rms.config.requestBodyContentType});

					for (let [name, schemaDb] of Object.entries(openApiDb.components.schemas)) {
						openApiDb.components.schemas[name] = OpenApi.mergeSchemas(openapi.components.schemas[name], schemaDb, false, name);
					}

					for (let name in openApiDb.components.schemas) {
						if (openapi.components.schemas[name] == undefined) {
							if (openApiDb.components.schemas[name] != null) openapi.components.schemas[name] = openApiDb.components.schemas[name];
							if (openApiDb.paths["/" + name] != null) openapi.paths["/" + name] = openApiDb.paths["/" + name];
							if (openApiDb.components.parameters[name] != null) openapi.components.parameters[name] = openApiDb.components.parameters[name];
							if (openApiDb.components.requestBodies[name] != null) openapi.components.requestBodies[name] = openApiDb.components.requestBodies[name];
							if (openApiDb.components.responses[name] != null) openapi.components.responses[name] = openApiDb.components.responses[name];
						}
					}

					return rms.storeOpenApi(openapi);
				});
			});
		*/
		return rms.LoadOpenApi()
	}

	//console.log(`[${rms.constructor.name}] starting ${rms.config.appName}...`);
	if err := rms.entityManager.Connect(); err != nil {
		return err
	}

	if err := createRufsTables(); err != nil {
		return err
	}

	if openapi, err := syncDb2OpenApi(); err != nil {
		return err
	} else {
		//console.log(`[${rms.constructor.name}.listen()] openapi after syncDb2OpenApi`);
		if err := RequestFilterUpdateRufsServices(rms.entityManager, openapi); err != nil {
			return err
		}

		if err := rms.MicroServiceServer.Listen(); err != nil {
			return err
		}
		//then(() => console.log(`[${rms.constructor.name}] ... ${rms.config.appName} started.`));
		return nil
	}
}
