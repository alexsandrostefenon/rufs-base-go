package rufsBase

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

type RufsGroupOwner struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Route struct {
	Path        string `json:"path"`
	Controller  string `json:"controller"`
	TemplateUrl string `json:"templateUrl"`
}

type MenuItem struct {
	Menu  string `json:"menu"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

type Role struct {
	Path string `json:"path"`
	Mask int    `json:"mask"`
}

type RufsUserProteced struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	RufsGroupOwner int    `json:"rufsGroupOwner"`
	Groups         []int  `json:"groups"`
	Roles          []Role `json:"roles"`
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
	JwtHeader string   `json:"jwtHeader"`
	Title     string   `json:"title"`
	Openapi   *OpenApi `json:"openapi"`
}

type RufsClaims struct {
	*jwt.StandardClaims
	TokenPayload
}

type EntityManager interface {
	Connect() error
	Find(tableName string, fields map[string]any, orderBy []string) ([]map[string]any, error)
	FindOne(tableName string, fields map[string]any) (map[string]any, error)
	Insert(tableName string, obj map[string]any) (map[string]any, error)
	Update(tableName string, key map[string]any, obj map[string]any) (map[string]any, error)
	DeleteOne(tableName string, key map[string]any) error
	UpdateOpenApi(openapi *OpenApi, options FillOpenApiOptions) error
	CreateTable(name string, schema *Schema) (sql.Result, error)
}

type IRufsMicroService interface {
	IMicroServiceServer
	LoadFileTables() error
}

type RufsMicroService struct {
	MicroServiceServer
	dbConfig                  *DbConfig
	checkRufsTables           bool
	migrationPath             string
	Irms                      IRufsMicroService
	wsServerConnectionsTokens map[string]*RufsClaims
	//dataStoreManager          *DataStoreManager
	entityManager EntityManager
	fileDbAdapter *FileDbAdapter
}

func (rms *RufsMicroService) authenticateUser(userName string, userPassword string, remoteAddr string) (*LoginResponse, error) {
	var entityManager EntityManager

	if _, ok := rms.fileDbAdapter.fileTables["rufsUser"]; ok {
		entityManager = rms.fileDbAdapter
	} else {
		entityManager = rms.entityManager
	}

	time.Sleep(100 * time.Millisecond)
	user := &RufsUser{}

	if userMap, err := entityManager.FindOne("rufsUser", map[string]any{"name": userName}); err == nil {
		data, _ := json.Marshal(userMap)

		if err := json.Unmarshal(data, user); err != nil {
			UtilsShowJsonUnmarshalError(string(data), err)
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("[RufsMicroService.authenticateUser] internal error : %s", err)
	}

	if len(user.Password) > 0 && user.Password != userPassword {
		return nil, errors.New("Don't match user and password.")
	}

	loginResponse := &LoginResponse{TokenPayload: TokenPayload{Ip: remoteAddr, RufsUserProteced: RufsUserProteced{Name: userName}}}
	loginResponse.Title = user.Name
	loginResponse.Id = user.Id
	loginResponse.RufsGroupOwner = user.RufsGroupOwner
	loginResponse.Roles = user.Roles
	loginResponse.Routes = user.Routes
	loginResponse.Path = user.Path
	loginResponse.Menu = user.Menu

	if loginResponse.RufsGroupOwner > 0 {
		/*
			const item = OpenApi.getPrimaryKeyForeign(this.openapi, "rufsUser", "rufsGroupOwner", user)
			return entityManager.findOne("rufsGroupOwner", item.primaryKey).then(rufsGroupOwner => {
				if (rufsGroupOwner != null) loginResponse.title = rufsGroupOwner.name + " - " + userName;
				return loginResponse
			})
		*/
	}

	if list, err := entityManager.Find("rufsGroupUser", map[string]any{"rufsUser": loginResponse.Id}, []string{}); err == nil {
		for _, item := range list {
			loginResponse.Groups = append(loginResponse.Groups, int(item["rufsGroup"].(int64)))
		}
	} else {
		return nil, fmt.Errorf("[RufsMicroService.authenticateUser] internal error : %s", err)
	}

	return loginResponse, nil
}

func (rms *RufsMicroService) OnRequest(req *http.Request) Response {
	if strings.HasSuffix(req.URL.Path, "/login") {
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
				loginResponse.Openapi = rms.openapi
			} else {
				//loginResponse.Openapi = rms.openapi.copy(loginResponse.Roles)
				loginResponse.Openapi = rms.openapi
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
		rf, err := RequestFilterInitialize(req, rms)

		if err != nil {
			return ResponseBadRequest(fmt.Sprintf("[RufsMicroService.OnRequest] : %s", err))
		}

		if access, err := rf.CheckAuthorization(req); err != nil {
			return ResponseBadRequest(fmt.Sprintf("[RufsMicroService.OnRequest.CheckAuthorization] : %s", err))
		} else if !access {
			return ResponseUnauthorized("Explicit Unauthorized")
		}

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
		rms.wsServerConnectionsTokens[tokenString] = claims
		log.Printf("[MicroServiceServer.onWsMessageFromClient] Ok")
	} else {
		fmt.Println(err)
	}
}

func (rms *RufsMicroService) LoadFileTables() error {
	loadTable := func(name string, defaultRows []map[string]any) error {
		var err error

		if _, err = rms.entityManager.Find(name, map[string]any{}, []string{}); err != nil {
			err = rms.fileDbAdapter.Load(name, defaultRows)
		}

		return err
	}

	var emptyList []map[string]any

	if rms.openapi == nil {
		if err := rms.Imss.LoadOpenApi(); err != nil {
			return err
		}
	}

	rms.fileDbAdapter = &FileDbAdapter{fileTables: make(map[string][]map[string]any), openapi: rms.openapi}
	RequestFilterUpdateRufsServices(rms.fileDbAdapter, rms.openapi)

	if err := loadTable("rufsGroup", emptyList); err != nil {
		return err
	}

	if err := loadTable("rufsGroupUser", emptyList); err != nil {
		return err
	}

	if err := loadTable("rufsGroupOwner", []map[string]any{defaultGroupOwnerAdmin}); err != nil {
		return err
	}

	if err := loadTable("rufsUser", []map[string]any{defaultUserAdmin}); err != nil {
		return err
	}

	return nil
}

func UtilsShowJsonUnmarshalError(str string, err error) {
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

	if jsonError, ok := err.(*json.SyntaxError); ok {
		line, character, lcErr := lineAndCharacter(str, int(jsonError.Offset))
		fmt.Fprintf(os.Stderr, "Cannot parse JSON schema due to a syntax error at line %d, character %d: %v\n", line, character, jsonError.Error())

		if lcErr != nil {
			fmt.Fprintf(os.Stderr, "Couldn't find the line and character position of the error due to error %v\n", lcErr)
		}
	} else if jsonError, ok := err.(*json.InvalidUnmarshalError); ok {
		fmt.Fprintf(os.Stderr, "Cannot parse JSON schema due to a InvalidUnmarshalError : %v\n", jsonError.Error())
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
func (rms *RufsMicroService) Listen() error {
	createRufsTables := func(openapiRufs *OpenApi) error {
		if !rms.checkRufsTables {
			return nil
		}

		for _, name := range []string{"rufsGroupOwner", "rufsUser", "rufsGroup", "rufsGroupUser"} {
			if _, ok := rms.openapi.Components.Schemas[name]; !ok {
				schema := openapiRufs.Components.Schemas[name]

				if _, err := rms.entityManager.CreateTable(name, schema); err != nil {
					return err
				}
			}
		}

		if response, _ := rms.entityManager.FindOne("rufsGroupOwner", map[string]any{"name": "ADMIN"}); response == nil {
			if _, err := rms.entityManager.Insert("rufsGroupOwner", defaultGroupOwnerAdmin); err != nil {
				return err
			}
		}

		if response, _ := rms.entityManager.FindOne("rufsUser", map[string]any{"name": "admin"}); response == nil {
			if _, err := rms.entityManager.Insert("rufsUser", defaultUserAdmin); err != nil {
				return err
			}
		}

		return nil
	}

	execMigrations := func() error {
		getVersion := func(name string) (int, error) {
			regExp := regexp.MustCompile(`(\d{1,3})\.(\d{1,3})\.(\d{1,3})`)
			regExpResult := regExp.FindStringSubmatch(name)

			if len(regExpResult) != 4 {
				return 0, fmt.Errorf(`Missing valid version in name %s`, name)
			}

			version, _ := strconv.Atoi(fmt.Sprintf(`%03s%03s%03s`, regExpResult[1], regExpResult[2], regExpResult[3]))
			return version, nil
		}

		migrate := func(fileName string) error {
			file, err := os.Open(filepath.Join(rms.migrationPath, fileName)) //`${this.config.migrationPath}/${fileName}`, "utf8"

			if err != nil {
				return err
			}

			defer file.Close()
			fileData, err := ioutil.ReadAll(file)

			if err != nil {
				return err
			}

			text := string(fileData)
			list := strings.Split(text, "--split")

			for _, sql := range list {
				_, err := rms.entityManager.(*DbClientSql).client.Exec(sql)

				if err != nil {
					return err
				}
			}

			newVersion, err := getVersion(fileName)

			if err != nil {
				return err
			}

			rms.openapi.Info.Version = fmt.Sprintf(`%d.%d.%d`, ((newVersion/1000)/1000)%1000, (newVersion/1000)%1000, newVersion%1000)
			return err
		}

		if rms.migrationPath == "" {
			rms.migrationPath = fmt.Sprintf(`./rufs-%s-es6/sql`, rms.appName)
		}

		if _, err := os.Stat(rms.migrationPath); errors.Is(err, os.ErrNotExist) {
			return nil
		}

		oldVersion, err := getVersion(rms.openapi.Info.Version)

		if err != nil {
			return err
		}

		files, err := ioutil.ReadDir(rms.migrationPath)

		if err != nil {
			return err
		}

		list := []string{}

		for _, fileInfo := range files {
			version, err := getVersion(fileInfo.Name())

			if err != nil {
				return err
			}

			if version > oldVersion {
				list = append(list, fileInfo.Name())
			}
		}

		sort.Slice(list, func(i, j int) bool {
			versionI, _ := getVersion(list[i])
			versionJ, _ := getVersion(list[j])
			return versionI < versionJ
		})

		for _, fileName := range list {
			if err := migrate(fileName); err != nil {
				return err
			}
		}

		rms.entityManager.UpdateOpenApi(rms.openapi, FillOpenApiOptions{requestBodyContentType: rms.requestBodyContentType})
		return rms.StoreOpenApi("")
	}

	if err := json.Unmarshal([]byte(defaultGroupOwnerAdminStr), &defaultGroupOwnerAdmin); err != nil {
		UtilsShowJsonUnmarshalError(defaultGroupOwnerAdminStr, err)
		return err
	}

	if err := json.Unmarshal([]byte(defaultUserAdminStr), &defaultUserAdmin); err != nil {
		UtilsShowJsonUnmarshalError(defaultUserAdminStr, err)
		return err
	}

	rms.wsServerConnectionsTokens = make(map[string]*RufsClaims)

	if rms.appName == "" {
		rms.appName = "base"
	}

	if rms.Irms == nil {
		rms.Irms = rms
	}

	if rms.Imss == nil {
		rms.Imss = rms
	}

	openapiRufs := &OpenApi{}

	if err := json.Unmarshal([]byte(rufsMicroServiceOpenApiStr), openapiRufs); err != nil {
		UtilsShowJsonUnmarshalError(rufsMicroServiceOpenApiStr, err)
		return err
	}

	if rms.openapi == nil {
		if err := rms.Imss.LoadOpenApi(); err != nil {
			return err
		}
	}

	rms.entityManager = &DbClientSql{dbConfig: rms.dbConfig}

	//console.log(`[${rms.constructor.name}] starting ${rms.config.appName}...`);
	if err := rms.entityManager.Connect(); err != nil {
		return err
	}

	if err := rms.entityManager.UpdateOpenApi(rms.openapi, FillOpenApiOptions{requestBodyContentType: rms.requestBodyContentType}); err != nil {
		return err
	}

	if err := createRufsTables(openapiRufs); err != nil {
		return err
	}

	rms.openapi.FillOpenApi(FillOpenApiOptions{schemas: openapiRufs.Components.Schemas, requestBodyContentType: rms.requestBodyContentType, security: map[string][]string{"jwt": {}}})

	if err := execMigrations(); err != nil {
		return err
	}

	rms.Irms.LoadFileTables()

	if err := RequestFilterUpdateRufsServices(rms.entityManager, rms.openapi); err != nil {
		return err
	}

	if err := rms.MicroServiceServer.Listen(); err != nil {
		return err
	}

	return nil
}

var rufsMicroServiceOpenApiStr string = `{
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
					"rufsGroupOwner": {"type": "integer", "nullable": false, "$ref": "#/components/schemas/rufsGroupOwner"},
					"name":           {"maxLength": 32, "nullable": false, "unique": true},
					"password":       {"nullable": false},
					"path":           {},
					"roles":          {"type": "array", "items": {"properties": {"name": {"type": "string"}, "mask": {"type": "integer"}}}},
					"routes":         {"type": "array", "items": {"properties": {"path": {"type": "string"}, "controller": {"type": "string"}, "templateUrl": {"type": "string"}}}},
					"menu":           {"type": "object", "properties": {"menu": {"type": "string"}, "label": {"type": "string"}, "path": {"type": "string"}}}
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
					"rufsUser":  {"type": "integer", "nullable": false, "$ref": "#/components/schemas/rufsUser"},
					"rufsGroup": {"type": "integer", "nullable": false, "$ref": "#/components/schemas/rufsGroup"}
				},
				"x-primaryKeys": ["rufsUser", "rufsGroup"],
				"x-uniqueKeys":  {}
			}
		}
	}
}
`
var defaultGroupOwnerAdminStr string = `{"name": "admin"}`
var defaultGroupOwnerAdmin map[string]any = map[string]any{}

var defaultUserAdminStr string = `{
		"name": "admin",
		"rufsGroupOwner": 1,
		"password": "21232f297a57a5a743894a0e4a801fc3",
		"path": "rufs_user/search",
		"menu": {},
		"roles": [
			{
				"mask": 31,
				"path": "/rufs_group_owner"
			},
			{
				"mask": 31,
				"path": "/rufs_user"
			},
			{
				"mask": 31,
				"path": "/rufs_group"
			},
			{
				"mask": 31,
				"path": "/rufs_group_user"
			}
		],
		"routes": [
			{
				"controller": "OpenApiOperationObjectController",
				"path": "/app/rufs_service/:action"
			},
			{
				"controller": "UserController",
				"path": "/app/rufs_user/:action"
			}
		]
	}`
var defaultUserAdmin map[string]any = map[string]any{}
