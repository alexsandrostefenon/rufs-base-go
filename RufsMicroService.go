package rufsBase

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type Role struct {
	get    bool
	post   bool
	patch  bool
	put    bool
	delete bool
}

type RufsGroupOwner struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Id             int             `json:"id"`
	Name           string          `json:"name"`
	Password       string          `json:"password"`
	RufsGroupOwner int             `json:"rufsGroupOwner"`
	Routes         string          `json:"routes"`
	Path           string          `json:"path"`
	Menu           string          `json:"menu"`
	Roles          map[string]Role `json:"roles"`
}

type TokenPayload struct {
	Ip             string         `json:"ip"`
	Name           string         `json:"name"`
	RufsGroupOwner int            `json:"rufsGroupOwner"`
	Groups         []string       `json:"groups"`
	Roles          map[string]int `json:"roles"`
}

type LoginResponse struct {
	TokenPayload
	JwtHeader string `json:"jwtHeader"`
	Title     string `json:"title"`
	Routes    string `json:"routes"`
	Path      string `json:"path"`
	Menu      string `json:"menu"`
	Openapi   string `json:"openapi"`
}

type RufsClaims struct {
	*jwt.StandardClaims
	TokenPayload
}

type DataStoreManager struct {
	openapi string
}

type EntityManager struct {
	dataStoreManager DataStoreManager
}

func EntityManagerFind[T any](em *EntityManager, name string, tmp T) ([]T, error) {
	return nil, errors.New("don't implemented")
}

type FileDbAdapter struct {
	EntityManager
}

func FileDbAdapterLoad[T any](fda *FileDbAdapter, name string, tmp T) (list []T, err error) {
	var data []byte
	fileName := fmt.Sprintf("%s.json", name)
	log.Printf("[FileDbAdapter.Load(%s)]", name)

	if data, err = ioutil.ReadFile(fileName); err == nil {
		err = json.Unmarshal(data, &list)
	}

	log.Printf("[FileDbAdapter.Load(%s)] : err = %s", name, err)
	return list, err
}

func FileDbAdapterStore[T any](fda *FileDbAdapter, name string, list []T) error {
	return nil
}

type RufsMicroService struct {
	MicroServiceServer
	appName string
	//	checkRufsTables bool
	migrationPath  string
	entityManager  EntityManager
	fileDbAdapter  FileDbAdapter
	listGroup      []interface{}
	listGroupUser  []interface{}
	listGroupOwner []RufsGroupOwner
	listUser       []User
}

var defaultGroupOwnerAdmin []RufsGroupOwner = []RufsGroupOwner{{Name: "ADMIN"}}

var defaultUserAdmin []User = []User{{
	Name: "admin", RufsGroupOwner: 1,
	//password: HttpRestRequest.MD5("admin"), path: "rufs_user/search",
	//roles:  {"rufsGroupOwner": {"post": true, "put": true, "delete": true}, "rufsUser": {"post": true, "put": true, "delete": true}, "rufsGroup": {"post": true, "put": true, "delete": true}, "rufsGroupUser": {"post": true, "put": true, "delete": true}},
	Routes: `[{"path": "/app/rufs_service/:action", "controller": "OpenApiOperationObjectController"}, {"path": "/app/rufs_user/:action", "controller": "UserController"}]`,
}}

func (rms *RufsMicroService) Init(imss IMicroServiceServer) {
	rms.MicroServiceServer.Init(imss)

	if rms.appName == "" {
		rms.appName = "base"
	}

	if rms.migrationPath == "" {
		rms.migrationPath = `./rufs-${config.appName}-es6/sql`
	}

	//	rms.entityManager = new DbClientPostgres(rms.config.dbConfig, {missingPrimaryKeys: rms.config.dbMissingPrimaryKeys, missingForeignKeys: rms.config.dbMissingForeignKeys, aliasMap: rms.config.aliasMap});
}

func (rms *RufsMicroService) AuthenticateUser(userName string, userPassword string, loginResponse *LoginResponse) (map[string]Role, error) {
	time.Sleep(100 * time.Millisecond)
	rms.loadRufsTables()
	var user *User

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
				if role.get {
					mask |= 1 << 0
				}
				if role.post {
					mask |= 1 << 1
				}
				if role.patch {
					mask |= 1 << 2
				}
				if role.put {
					mask |= 1 << 3
				}
				if role.delete {
					mask |= 1 << 4
				}

				ret[schemaName] = mask
			}

			return ret
		}

		var user User
		err := json.NewDecoder(req.Body).Decode(&user)

		if err != nil {
			return ResponseUnauthorized(fmt.Sprint(err))
		}

		loginResponse := LoginResponse{TokenPayload: TokenPayload{Ip: req.RemoteAddr, Name: user.Name}}

		if roles, err := rms.AuthenticateUser(user.Name, user.Password, &loginResponse); err == nil {
			if user.Name == "admin" {
				loginResponse.Openapi = rms.entityManager.dataStoreManager.openapi
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
		return ResponseOk("TODO")
		/*
			access := RequestFilter.checkAuthorization(req, resource, action);
			if (access != true) return Response.unauthorized("Explicit Unauthorized");

			if (resource == "rufsService" && req.method == "GET" && action == "query") {
				const list = OpenApi.getList(Qs, OpenApi.convertRufsToStandart(rms.openapi, true), true, req.tokenPayload.roles);
				return Promise.resolve(Response.ok(list));
			}

			return RequestFilter.processRequest(req, res, next, rms.entityManager, rms, resource, action);
		*/
	}
}

func loadTable[T any](rms *RufsMicroService, name string, defaultRows []T) (list []T, err error) {
	var tmp T

	if list, err = EntityManagerFind(&rms.entityManager, name, tmp); err != nil {
		if list, err = FileDbAdapterLoad(&rms.fileDbAdapter, name, tmp); err == nil {
			if len(list) == 0 && len(defaultRows) > 0 {
				FileDbAdapterStore(&rms.fileDbAdapter, name, defaultRows)
				list = defaultRows
			}
		} else {
			FileDbAdapterStore(&rms.fileDbAdapter, name, defaultRows)
		}
	}

	return list, err
}

func (rms *RufsMicroService) loadRufsTables() {
	/*
		return rms.loadOpenApi().
		then(openapi => {
			rms.fileDbAdapter = new FileDbAdapter(openapi);
			return RequestFilter.updateRufsServices(rms.fileDbAdapter, openapi);
		}).
	*/
	var emptyList []interface{}
	loadTable(rms, "rufsService", emptyList)

	if list, err := loadTable(rms, "rufsGroup", emptyList); err == nil {
		rms.listGroup = list
	}

	if list, err := loadTable(rms, "rufsGroupUser", emptyList); err == nil {
		rms.listGroupUser = list
	}

	if list, err := loadTable(rms, "rufsGroupOwner", defaultGroupOwnerAdmin); err == nil {
		rms.listGroupOwner = list
	}

	if list, err := loadTable(rms, "rufsUser", defaultUserAdmin); err == nil {
		rms.listUser = list
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

func (rms *RufsMicroService) listen() {
	const createRufsTables = () => {
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
	}

	const syncDb2OpenApi = () => {
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
	}

	console.log(`[${rms.constructor.name}] starting ${rms.config.appName}...`);
	return rms.entityManager.connect().
	then(() => createRufsTables()).
	then(() => syncDb2OpenApi()).
	then(openapi => {
		console.log(`[${rms.constructor.name}.listen()] openapi after syncDb2OpenApi`);
		return Promise.resolve().
		then(() => {
			return RequestFilter.updateRufsServices(rms.entityManager, openapi);
		}).
		then(() => super.listen()).
		then(() => console.log(`[${rms.constructor.name}] ... ${rms.config.appName} started.`));
	});
}
*/
