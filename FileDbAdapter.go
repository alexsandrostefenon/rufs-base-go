package rufsBase

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
)

type FileDbAdapter struct {
	openapi    *OpenApi
	fileTables map[string][]map[string]any
}

func (fileDbAdapter *FileDbAdapter) Connect() error {
	return nil
}

/*
	constructor(openapi) {
		this.fileTables = new Map();
		this.openapi = openapi;
	}
*/
func (fda *FileDbAdapter) Load(name string, defaultRows []map[string]any) (err error) {
	var data []byte
	var list []map[string]any

	if data, err = ioutil.ReadFile(fmt.Sprintf("%s.json", name)); err == nil {
		err = json.Unmarshal(data, &list)
	}

	if fda.fileTables == nil {
		fda.fileTables = make(map[string][]map[string]any)
	}

	if len(list) == 0 && len(defaultRows) > 0 {
		err = fda.store(name, defaultRows)
		list = defaultRows
	}

	fda.fileTables[name] = list

	return err
}

func (fda *FileDbAdapter) store(name string, list []map[string]any) error {
	/*
		const schema = this.openapi.components.schemas[tableName];
		let listOut;

		if (schema.properties.id != undefined) {
			const rufsSchema = new DataStore(tableName, schema);
			listOut = [];

			for (let i = 0; i < list.length; i++) {
				let item = list[i];

				if (item.id == undefined) {
					item = OpenApi.copyFields(rufsSchema, item);
					item.id = ++i;
				}

				listOut.push(item);
			}
		} else {
			listOut = list;
		}

	*/
	fileName := fmt.Sprintf("%s.json", name)
	log.Printf("[FileDbAdapterStore] : writing file %s ...", fileName)

	if data, err := json.Marshal(list); err != nil {
		log.Fatalf("[FileDbAdapterStore] : failt to marshal list before wrinting file %s : %s", fileName, err)
		return err
	} else if err = ioutil.WriteFile(fileName, data, fs.ModePerm); err != nil {
		log.Fatalf("[FileDbAdapterStore] : failt to write file %s : %s", fileName, err)
		return err
	}

	log.Printf("[FileDbAdapterStore] : ... writed file %s", fileName)
	fda.fileTables[name] = list
	return nil
}

func (fileDbAdapter *FileDbAdapter) Insert(tableName string, obj map[string]any) (map[string]any, error) {
	list, ok := fileDbAdapter.fileTables[tableName]

	if !ok {
		return nil, fmt.Errorf("[FileDbAdapter.Update(name = %s)] : don't find table", tableName)
	}

	if fileDbAdapter.openapi.Components.Schemas[tableName].Properties["id"] != nil {
		id := 0

		for _, item := range list {
			buffer, err := json.Marshal(item)

			if err != nil {
				return nil, err
			}

			itemMap := map[string]any{}
			json.Unmarshal(buffer, &itemMap)

			if value, ok := itemMap["id"]; ok && int(value.(float64)) > id {
				id = int(value.(float64))
			}
		}

		obj["id"] = id + 1
	}

	list = append(list, obj)
	fileDbAdapter.store(tableName, list)
	return obj, nil
}

func (fileDbAdapter *FileDbAdapter) Find(tableName string, fields map[string]any, orderBy []string) ([]map[string]any, error) {
	if list, ok := fileDbAdapter.fileTables[tableName]; ok {
		return FilterFind(list, fields)
	}

	return nil, fmt.Errorf("Don't find")
}

func (fileDbAdapter *FileDbAdapter) FindOne(tableName string, key map[string]any) (map[string]any, error) {
	list, ok := fileDbAdapter.fileTables[tableName]

	if !ok {
		return nil, fmt.Errorf("[FileDbAdapter.FindOne] missing table %s", tableName)
	}

	obj, err := FilterFindOne(list, key)

	if err != nil {
		return nil, fmt.Errorf("[FileDbAdapter.FindOne] don't found register in %s with key %s", tableName, key)
	}

	objMap := map[string]any{}
	buffer, _ := json.Marshal(obj)
	err = json.Unmarshal(buffer, &objMap)
	return objMap, err
}

func (fileDbAdapter *FileDbAdapter) Update(tableName string, key map[string]any, obj map[string]any) (map[string]any, error) {
	list, ok := fileDbAdapter.fileTables[tableName]

	if !ok {
		return nil, fmt.Errorf("[FileDbAdapter.Update(name = %s)] : don't find table", tableName)
	}

	pos, err := FilterFindIndex(list, key)

	if pos < 0 || err != nil {
		return nil, fmt.Errorf("[FileDbAdapter.update(name = %s, key = %s)] fail : %s", tableName, key, err)
	}

	list[pos] = obj
	fileDbAdapter.store(tableName, list)
	return obj, nil
}

func (fileDbAdapter *FileDbAdapter) DeleteOne(tableName string, key map[string]any) error {
	list, ok := fileDbAdapter.fileTables[tableName]

	if !ok {
		return fmt.Errorf("[FileDbAdapter.DeleteOne(name = %s)] : don't find table", tableName)
	}

	pos, err := FilterFindIndex(list, key)

	if pos < 0 || err != nil {
		return fmt.Errorf("[FileDbAdapter.DeleteOne(name = %s, key = %s)] fail : %s", tableName, key, err)
	}

	list = append(list[:pos], list[pos+1:]...)
	return fileDbAdapter.store(tableName, list)
}

func (fileDbAdapter *FileDbAdapter) UpdateOpenApi(openapi *OpenApi, options FillOpenApiOptions) error {
	return nil
}

func (fileDbAdapter *FileDbAdapter) CreateTable(name string, schema *Schema) (sql.Result, error) {
	return nil, nil
}
