package rufsBase

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type OpenApiSecurity struct {
}

type OpenApiContact struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Email string `json:"email"`
}

type OpenApiInfo struct {
	Title       string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Contact     OpenApiContact `json:"contact"`
}

type OpenApiServerComponent struct {
	Url string `json:"url"`
}

type OpenApiOperationObject struct {
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	OperationId string   `json:"operationId"`
	Parameters  []OpenApiParameterObject
	RequestBody OpenApiRequestBodyObject
	Responses   map[string]OpenApiResponseObject `json:"responses"`
}

type Schema struct {
	Name               string             `json:"name"`
	PrimaryKeys        []string           `json:"x-primaryKeys"`
	UniqueKeys         any                `json:"x-uniqueKeys"`
	ForeignKeys        any                `json:"x-foreignKeys"`
	Required           []string           `json:"required"`
	Ref                string             `json:"x-$ref"`
	Type               string             `json:"type"`
	Format             string             `json:"format"`
	Nullable           bool               `json:"nullable"`
	Essential          bool               `json:"x-required"`
	Title              string             `json:"x-title"`
	Hiden              bool               `json:"x-hiden"`
	InternalName       string             `json:"x-internalName"`
	Default            string             `json:"default"`
	Enum               []string           `json:"enum"`
	EnumLabels         []string           `json:"x-enumLabels"`
	IdentityGeneration string             `json:"x-identityGeneration"`
	Updatable          bool               `json:"x-updatable"`
	Scale              int                `json:"x-scale"`
	Precision          int                `json:"x-precision"`
	MaxLength          int                `json:"maxLength"`
	Properties         map[string]*Schema `json:"properties"`
	Items              *Schema            `json:"items"`
}

type OpenApiParameterObject struct {
	Ref         string  `json:"$ref"`
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Schema      *Schema `json:"schema"`
}

type OpenApiMediaTypeObject struct {
	Schema *Schema `json:"schema"`
}

type OpenApiRequestBodyObject struct {
	Ref     string                             `json:"$ref"`
	Content map[string]*OpenApiMediaTypeObject `json:"content"`
}

type OpenApiResponseObject struct {
	Ref     string                             `json:"$ref"`
	Content map[string]*OpenApiMediaTypeObject `json:"content"`
}

type OpenApiSecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme"`
	Name         string `json:"name"`
	In           string `json:"in"`
	BearerFormat string `json:"bearerFormat"`
}

type OpenApi struct {
	Openapi    string                                       `json:"openapi"`
	Info       OpenApiInfo                                  `json:"info"`
	Servers    []*OpenApiServerComponent                    `json:"servers"`
	Paths      map[string]map[string]OpenApiOperationObject `json:"paths"`
	Components struct {
		Schemas         map[string]*Schema                  `json:"schemas"`
		Parameters      map[string]OpenApiParameterObject   `json:"parameters"`
		RequestBodies   map[string]OpenApiRequestBodyObject `json:"requestBodies"`
		Responses       map[string]OpenApiResponseObject    `json:"responses"`
		SecuritySchemes map[string]OpenApiSecurityScheme    `json:"securitySchemes"`
	}
	Security []any `json:"security"`
	Tags     []any `json:"tags"`
}

func OpenApiCreate(openapi *OpenApi, security any) {
	if openapi.Openapi == "" {
		openapi.Openapi = "3.0.3"
	}

	openapi.Info = OpenApiInfo{Title: "rufs-base-es6 openapi genetator", Version: "0.0.0", Description: "CRUD operations", Contact: OpenApiContact{Name: "API Support", Url: "http://www.example.com/support", Email: "support@example.com"}}

	openapi.Components.SecuritySchemes = map[string]OpenApiSecurityScheme{
		"jwt":    {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
		"apiKey": {Type: "apiKey", In: "header", Name: "X-API-KEY"},
		"basic":  {Type: "http", Scheme: "basic"},
	}
}

/*
	static copy(dest, source, roles) {
		dest.openapi = source.openapi;
		dest.info = source.info;
		dest.servers = source.servers;
		dest.components.securitySchemes = source.components.securitySchemes;
		dest.security = source.security;
		dest.tags = source.tags;

		for (let [schemaName, role] of Object.entries(roles)) {
			if (source.components.schemas[schemaName] != nil) dest.components.schemas[schemaName] = source.components.schemas[schemaName];
			if (source.components.responses[schemaName] != nil) dest.components.responses[schemaName] = source.components.responses[schemaName];
			if (source.components.parameters[schemaName] != nil) dest.components.parameters[schemaName] = source.components.parameters[schemaName];
			if (source.components.requestBodies[schemaName] != nil) dest.components.requestBodies[schemaName] = source.components.requestBodies[schemaName];

			const pathIn = source.paths["/"+schemaName];
			if (pathIn == nil) continue;
			const pathOut = dest.paths["/"+schemaName] = {};
			// TODO : alterar UserController para não usar valores default
			const defaultAccess = {get: true, post: false, patch: false, put: false, delete: false};

			for (const [method, value] of Object.entries(defaultAccess)) {
				if (role[method] == nil) role[method] = value;
			}

			for (let [method, value] of Object.entries(role)) {
				if (value == true) pathOut[method] = pathIn[method];
			}
		}

		if (dest.components.responses.Error == nil) dest.components.responses.Error = source.components.responses.Error;
	}

	static mergeSchemas(schemaOld, schemaNew, keepOld, schemaName) {
		const mergeArray = (oldArray, newArray) => {
			if (newArray == nil) return oldArray;
			if (oldArray == nil) return newArray;
			for (const item of newArray) if (oldArray.includes(item) == false) oldArray.push(item);
			return oldArray;
		};

		schemaOld = schemaOld != nil && schemaOld != null ? schemaOld : {};
//		console.log(`[${this.constructor.name}.updateJsonSchema(schemaName: ${schemaName}, schemaNew.properties: ${schemaNew.properties}, schemaOld.properties: ${schemaOld.properties})]`);
		const jsonSchemaTypes = ["boolean", "string", "integer", "number", "date-time", "date", "object", "array"];
		if (schemaNew.properties == nil) schemaNew.properties = {};
		if (schemaOld.properties == nil) schemaOld.properties = {};
		let newFields = schemaNew.properties || {};
		let oldFields = schemaOld.properties || {};
		if (typeof(newFields) == "string") newFields = Object.entries(JSON.parse(newFields));
		if (typeof(oldFields) == "string") oldFields = JSON.parse(oldFields);
		const newFieldsIterator = newFields instanceof Map == true ? newFields : Object.entries(newFields);
		let jsonBuilder = {};
		if (keepOld == true) jsonBuilder = oldFields;

		for (let [fieldName, field] of newFieldsIterator) {
			if (field == null) field = {};
			if (field.type == nil) field.type = "string";

			if (field.hiden == nil && field.identityGeneration != nil) {
				field.hiden = true;
			}

			if (field.readOnly == nil && field.identityGeneration != nil) field.readOnly = true;

			if (jsonSchemaTypes.indexOf(field.type) < 0) {
				console.error(`${schemaName} : ${fieldName} : Unknow type : ${field.type}`);
				continue;
			}
			// type (columnDefinition), readOnly, hiden, primaryKey, essential (insertable), updatable, default, length, precision, scale
			let jsonBuilderValue = {};
			// registra conflitos dos valores antigos com os valores detectados do banco de dados
			jsonBuilderValue["type"] = field.type;
			jsonBuilderValue["format"] = field.format;

			if (field.updatable == false) {
				jsonBuilderValue["updatable"] = false;
			}

			if (field.maxLength > 0) {
				jsonBuilderValue["maxLength"] = field.maxLength;
			}

			if (field.precision > 0) {
				jsonBuilderValue["precision"] = field.precision;
			}

			if (field.scale > 0) {
				jsonBuilderValue["scale"] = field.scale;
			}

			if (field.nullable == true) {
				jsonBuilderValue["nullable"] = true;
			} else {
				jsonBuilderValue["nullable"] = field.nullable;
			}
			//
			if (field.$ref != nil) {
				jsonBuilderValue["$ref"] = field.$ref;
			}

			if (field.properties != nil) {
				jsonBuilderValue["properties"] = field.properties;
			}

			if (field.items != nil) {
				jsonBuilderValue["items"] = field.items;
			}

			if (field.internalName != null) jsonBuilderValue["internalName"] = field.internalName;
			if (field.essential != nil) jsonBuilderValue["essential"] = field.essential;
			if (field.default != nil) jsonBuilderValue["default"] = field.default;
			if (field.unique != nil) jsonBuilderValue["unique"] = field.unique;
			if (field.identityGeneration != nil) jsonBuilderValue["identityGeneration"] = field.identityGeneration;
			if (field.isClonable != nil) jsonBuilderValue["isClonable"] = field.isClonable;
			if (field.hiden != nil) jsonBuilderValue["hiden"] = field.hiden;
			if (field.readOnly != nil) jsonBuilderValue["readOnly"] = field.readOnly;
			if (field.description != nil) jsonBuilderValue["description"] = field.description;
			// oculta tipos incompatíveis
			if (jsonBuilderValue["type"] != "string") {
				delete jsonBuilderValue["length"];
			}

			if (jsonBuilderValue["type"] != "number") {
				delete jsonBuilderValue["precision"];
				delete jsonBuilderValue["scale"];
			}

			if (jsonBuilderValue["type"] != "object") {
				delete jsonBuilderValue["properties"];
			}

			if (jsonBuilderValue["type"] != "array") {
				delete jsonBuilderValue["items"];
			}
			// habilita os campos PLENAMENTE não SQL
			if (field.title != nil) jsonBuilderValue.title = field.title;
			if (field.document != nil) jsonBuilderValue.document = field.document;
			if (field.sortType != nil) jsonBuilderValue.sortType = field.sortType;
			if (field.orderIndex != nil) jsonBuilderValue.orderIndex = field.orderIndex;
			if (field.tableVisible != nil) jsonBuilderValue.tableVisible = field.tableVisible;
			if (field.shortDescription != nil) jsonBuilderValue.shortDescription = field.shortDescription;

			if (field.enum != nil) jsonBuilderValue.enum = mergeArray(jsonBuilderValue.enum, field.enum);
			if (field.enumLabels != nil) jsonBuilderValue.enumLabels = mergeArray(jsonBuilderValue.enumLabels, field.enumLabels);
			// exceções
			if (oldFields[fieldName] != null) {
				let fieldOriginal = oldFields[fieldName];
				// copia do original os campos PLENAMENTE não SQL
				jsonBuilderValue.title = fieldOriginal.title;
				jsonBuilderValue.document = fieldOriginal.document;
				jsonBuilderValue.sortType = fieldOriginal.sortType;
				jsonBuilderValue.orderIndex = fieldOriginal.orderIndex;
				jsonBuilderValue.tableVisible = fieldOriginal.tableVisible;
				jsonBuilderValue.shortDescription = fieldOriginal.shortDescription;

				jsonBuilderValue.enum = mergeArray(jsonBuilderValue.enum, fieldOriginal.enum);
				jsonBuilderValue.enumLabels = mergeArray(jsonBuilderValue.enumLabels, fieldOriginal.enumLabels);
				// registra conflitos dos valores antigos com os valores detectados do banco de dados
				const exceptions = ["service", "isClonable", "hiden", "$ref"];

				for (let subFieldName in fieldOriginal) {
					if (exceptions.indexOf(subFieldName) < 0 && fieldOriginal[subFieldName] != jsonBuilderValue[subFieldName]) {
						console.warn(`rufsServiceDbSync.generateJsonSchema() : table [${schemaName}], field [${fieldName}], property [${subFieldName}] conflict previous declared [${fieldOriginal[subFieldName]}] new [${jsonBuilderValue[subFieldName]}]\nold:\n`, fieldOriginal, "\nnew:\n", jsonBuilderValue);
					}
				}
				// copia do original os campos PARCIALMENTE não SQL
				if (fieldOriginal.isClonable != nil) jsonBuilderValue.isClonable = fieldOriginal.isClonable;
				if (fieldOriginal.readOnly != nil) jsonBuilderValue.readOnly = fieldOriginal.readOnly;
				if (fieldOriginal.hiden != nil) jsonBuilderValue.hiden = fieldOriginal.hiden;
			}
			// oculta os valores dafault
			const defaultValues = {updatable: true, maxLength: 255, precision: 9, scale: 3, hiden: false, primaryKey: false, essential: false};

			for (let subFieldName in defaultValues) {
				if (jsonBuilderValue[subFieldName] == defaultValues[subFieldName]) {
					delete jsonBuilderValue[subFieldName];
				}
			}
			// troca todos os valores null por nil
			for (let [key, value] of Object.entries(jsonBuilderValue)) {
				if (value == null) delete jsonBuilderValue[key];
			}

			if (jsonBuilderValue["type"] == "array" && oldFields[fieldName] != null)
				jsonBuilder[fieldName].items = this.mergeSchemas(oldFields[fieldName].items, newFields[fieldName].items, keepOld, schemaName);
			else if (jsonBuilderValue["type"] == "object" && oldFields[fieldName] != null)
				jsonBuilder[fieldName] = this.mergeSchemas(oldFields[fieldName], newFields[fieldName], keepOld, schemaName);
			else
				jsonBuilder[fieldName] = jsonBuilderValue;
		}

		const schema = {};
		schema.title = schemaOld.title || schemaNew.title;
		schema.type = "object";
		schema.required = [];
		schema.primaryKeys = schemaNew.primaryKeys;
		schema.uniqueKeys = schemaNew.uniqueKeys;
		schema.foreignKeys = schemaNew.foreignKeys;
		schema.properties = jsonBuilder;
		for (const [fieldName, field] of Object.entries(schema.properties)) if (field.essential == true) schema.required.push(fieldName);
		return schema;
	}

	static convertRufsToStandartSchema(schema, onlyClientUsage) {
		const standartSchema = {};
		standartSchema.type = schema.type || "object";
		standartSchema.required = schema.required || [];
		if (schema.primaryKeys && schema.primaryKeys.length > 0) standartSchema["x-primaryKeys"] = schema.primaryKeys;

		if (onlyClientUsage != true) {
			standartSchema["x-uniqueKeys"] = schema.uniqueKeys;
			standartSchema["x-foreignKeys"] = schema.foreignKeys;
		}

		standartSchema.description = schema.description;
		standartSchema.properties = {};

		for (let [fieldName, field] of Object.entries(schema.properties)) {
			if (onlyClientUsage == true && field.hiden == true) continue;
			let property = {};
			let type = field.type;

			if (type == "date-time" || type == "date") {
				property.type = "string";
				property.format = type;
			} else if (type == null && field.properties != null) {
				type = "object";
			} else {
				property.type = type;
			}

			if (field.description) property.description = field.description;
			if (field.default) property.default = field.default;
			if (field.enum) property.enum = field.enum;

			if (type == "object") {
				if (field.$ref) {
					property.$ref = field.$ref;
				} else {
					if (field.properties != nil) {
						property = this.convertRufsToStandartSchema(field, onlyClientUsage);
					} else {
						console.error(`[${this.constructor.name}.convertRufsToStandartSchema()] : missing "properties" in field ${fieldName} from schema :`, schema);
					}
				}
			} else if (type == "array") {
				if (field.items) {
					if (field.minItems != null) property.minItems = field.minItems;
					if (field.maxItems != null) property.maxItems = field.maxItems;

					if (field.items.type == "object") {
						if (field.items.$ref) {
							property.items = {};
							property.items.$ref = field.items.$ref;
						} else {
							if (field.items.properties != nil) {
								property.items = this.convertRufsToStandartSchema(field.items, onlyClientUsage);
							} else {
								console.error(`[${this.constructor.name}.convertRufsToStandartSchema()] : missing "properties" in field ${fieldName} from schema :`, schema);
							}
						}
					} else {
						property.items = field.items;
					}
				}

				if (field.hiden) property["x-hiden"] = field.hiden;
				if (field.internalName && onlyClientUsage != true) property["x-internalName"] = field.internalName;
				if (field.enumLabels && onlyClientUsage != true) property["x-enumLabels"] = field.enumLabels;
			} else {
				if (field.example) property.example = field.example;
				if (field.nullable) property.nullable = field.nullable;
				if (field.updatable) property["x-updatable"] = field.updatable;
				if (field.scale) property["x-scale"] = field.scale;
				if (field.precision) property["x-precision"] = field.precision;
				if (field.maxLength) property.maxLength = field.maxLength;
				if (field.pattern) property.pattern = field.pattern;
				if (field.format) property.format = field.format;
				if (field.$ref) property["x-$ref"] = field.$ref;
				if (field.title) property["x-title"] = field.title;

				if (onlyClientUsage != true) {
					if (field.essential) property["x-required"] = field.essential;
					if (field.hiden) property["x-hiden"] = field.hiden;
					if (field.internalName) property["x-internalName"] = field.internalName;
					if (field.enumLabels) property["x-enumLabels"] = field.enumLabels;
					if (field.identityGeneration) property["x-identityGeneration"] = field.identityGeneration;
				}
			}

			if (field.essential == true && standartSchema.required.indexOf(fieldName) < 0)
				standartSchema.required.push(fieldName);

			standartSchema.properties[fieldName] = property;
		}

		if (standartSchema.required.length == 0) delete standartSchema.required;
		return standartSchema;
	}

	static convertRufsToStandart(openapi, onlyClientUsage) {
		const standartOpenApi = {};
		standartOpenApi.openapi = openapi.openapi;
		standartOpenApi.info = openapi.info;
		standartOpenApi.servers = openapi.servers;

		standartOpenApi.paths = openapi.paths;
standartOpenApi.components = {};
const standartSchemas = {};

for (let [name, schema] of Object.entries(openapi.components.schemas)) {
	if (schema == nil) {
		console.error(`[${this.constructor.name}.convertRufsToStandart(openapi)] : openapi.components.schemas[${name}] is nil !`);
		continue;
	}

	standartSchemas[name] = this.convertRufsToStandartSchema(schema, onlyClientUsage);
}

standartOpenApi.components.schemas = standartSchemas;
standartOpenApi.components.parameters = openapi.components.parameters;
standartOpenApi.components.requestBodies = {};

for (let [name, requestBodyObject] of Object.entries(openapi.components.requestBodies)) {
	const standartRequestBodyObject = standartOpenApi.components.requestBodies[name] = {"required": true, "content": {}};

	for (let [mediaTypeName, mediaTypeObject] of Object.entries(requestBodyObject.content)) {
		standartRequestBodyObject.content[mediaTypeName] = {};

		if (mediaTypeObject.schema.properties != nil) {
			standartRequestBodyObject.content[mediaTypeName].schema = this.convertRufsToStandartSchema(mediaTypeObject.schema, onlyClientUsage);
		} else {
			standartRequestBodyObject.content[mediaTypeName].schema = mediaTypeObject.schema;
		}
	}
}

standartOpenApi.components.responses = openapi.components.responses;
standartOpenApi.components.securitySchemes = openapi.components.securitySchemes;
standartOpenApi.security = openapi.security;
standartOpenApi.tags = openapi.tags;
return standartOpenApi;
}
*/
func (openapi *OpenApi) convertStandartToRufs() {
	var convertSchema func(schema *Schema)

	convertSchema = func(schema *Schema) {
		for fieldName, field := range schema.Properties {
			for _, value := range schema.Required {
				if value == fieldName {
					field.Essential = true
					break
				}
			}

			if field.Format == "date-time" || field.Format == "date" {
				field.Type = field.Format
			}

			if field.Type == "object" && field.Properties != nil {
				convertSchema(field)
			} else if field.Type == "array" && field.Items != nil && field.Items.Type == "object" && field.Items.Properties != nil {
				convertSchema(field.Items)
			}
		}
	}

	for _, schema := range openapi.Components.Schemas {
		convertSchema(schema)
	}

	for _, requestBodyObject := range openapi.Components.RequestBodies {
		for _, mediaTypeObject := range requestBodyObject.Content {
			if mediaTypeObject.Schema.Properties != nil {
				convertSchema(mediaTypeObject.Schema)
			}
		}
	}
}

/*
static getMaxFieldSize(schema, fieldName) {
let ret = 0;
const field = schema.properties[fieldName];
const type = field["type"];

if (type == nil || type == "string") {
	if (field.maxLength != nil) {
		ret = field.maxLength;
	} else {
		ret = 100;
	}
} else if (type == "integer") {
	ret = 9;
} else if (type == "number") {
	if (field.precision != nil) {
		ret = field.precision;
	} else {
		ret = 15;
	}
} else if (type == "boolean") {
	ret = 5;
} else if (type == "date" || type == "date-time") {
	ret = 30;
}

return ret;
}
*/
func (openapi *OpenApi) copyValue(field *Schema, value any) (ret any, err error) {
	if value == nil && field.Essential && !field.Nullable {
		if field.Enum != nil && len(field.Enum) == 1 {
			value = field.Enum[0]
		} else if field.Default != "" {
			value = field.Default
		}
	}

	if field.Type == "" || field.Type == "string" {
		switch value.(type) {
		case string:
			if value != "" && field.MaxLength > 0 {
				ret = value.(string)[:field.MaxLength]
			} else {
				ret = value
			}
		default:
			ret = value
		}
	} else if field.Type == "integer" {
		switch value.(type) {
		case string:
			ret, err = strconv.Atoi(value.(string))
		default:
			ret = value
		}
	} else if field.Type == "number" {
		switch value.(type) {
		case string:
			ret, err = strconv.ParseFloat(value.(string), 64)
		default:
			ret = value
		}
	} else if field.Type == "boolean" {
		switch value.(type) {
		case bool:
			ret = value.(bool)
		case string:
			ret = (value == "true")
		}
	} else if field.Type == "date-time" {
		switch value.(type) {
		case string:
			if value != "" && field.MaxLength > 0 {
				ret = value.(string)[:field.MaxLength]
			} else {
				ret, _ = time.Parse(time.RFC3339, value.(string))
			}
		default:
			ret, _ = time.Parse(time.RFC3339, value.(string))
		}
	} else if field.Type == "date" {
		switch value.(type) {
		case string:
			if value != "" && field.MaxLength > 0 {
				ret = value.(string)[:10]
			} else {
				ret, _ = time.Parse(time.RFC3339, value.(string))
			}
		default:
			ret, _ = time.Parse(time.RFC3339, value.(string))
		}
	} else {
		ret = value
	}

	return ret, err
}

/*
static copyToInternalName(schema, dataIn) {
const copy = (property, valueIn) => {
	if (property.type == "object" && property.properties != nil) {
		return this.copyToInternalName(property, valueIn);
	} else if (property.type == "array" && property.items != nil && property.items.properties != nil) {
		const valueOut = [];
		for (const val of valueIn) valueOut.push(this.copyToInternalName(property.items, val));
		return valueOut;
	} else {
		return this.copyValue(property, valueIn);
	}
}

const dataOut = {};

for (let [name, property] of Object.entries(schema.properties)) {
	if (property.internalName != nil) {
		dataOut[property.internalName] = copy(property, dataIn[name]);
	} else {
		dataOut[name] = copy(property, dataIn[name]);
	}
}

return dataOut;
}

static copyFromInternalName(schema, dataIn, caseInsensitive) {
const copy = (property, valueIn) => {
	if (property.type == "object" && property.properties != nil) {
		return this.copyFromInternalName(property, valueIn, caseInsensitive);
	} else if (property.type == "array" && property.items != nil && property.items.properties != nil && Array.isArray(valueIn)) {
		const valueOut = [];
		for (const val of valueIn) valueOut.push(this.copyFromInternalName(property.items, val, caseInsensitive));
		return valueOut;
	} else {
		return this.copyValue(property, valueIn);
	}
}

const dataOut = {};
console.log(`[${this.constructor.name}.copyFromInternalName] dataIn :`, dataIn);

for (let [name, property] of Object.entries(schema.properties)) {
	if (property.internalName != nil) {
		if (caseInsensitive == true) {
			for (let fieldName in dataIn) {
				if (fieldName.toLowerCase() == property.internalName.toLowerCase()) {
					dataOut[name] = copy(property, dataIn[fieldName]);
				}
			}
		} else {
			dataOut[name] = copy(property, dataIn[property.internalName]);
		}
	} else {
		if (caseInsensitive == true) {
			for (let fieldName in dataIn) if (fieldName.toLowerCase() == name.toLowerCase()) dataOut[name] = copy(property, dataIn[fieldName]);
		} else {
			dataOut[name] = copy(property, dataIn[name]);
		}
	}
}

console.log(`[${this.constructor.name}.copyFromInternalName] dataOut :`, dataOut);
return dataOut;
}
*/
func (openapi *OpenApi) getValueFromSchema(schema *Schema, propertyName string, obj map[string]any) any {
	property, ok := schema.Properties[propertyName]
	var ret any

	if ok {
		if value, ok := obj[propertyName]; ok {
			ret = value
		} else if property.InternalName != "" && obj[property.InternalName] != nil {
			ret = obj[property.InternalName]
		} else if property.Nullable {
			if obj[propertyName] == nil {
				return nil
			}

			if property.InternalName != "" && obj[property.InternalName] == nil {
				return nil
			}
		}
	}

	if ret == nil {
		for fieldName, field := range schema.Properties {
			if field.InternalName == propertyName {
				property = field
				ret = obj[fieldName]
				break
			}
		}
	}

	if ret != nil {
		switch ret.(type) {
		case float64:
			if property.Type == "integer" {
				ret = int(ret.(float64))
			}
		}
	}
	/*
		if ret != nil && ret instanceof Date && !isNaN(ret) && property != nil && property.type == "date" {
			const str = ret.toISOString();
			ret = str.substring(0, 10);
		}
	*/
	return ret
}

func (openapi *OpenApi) copyFields(schema *Schema, dataIn map[string]any, ignorenil bool, ignoreHiden bool) (map[string]any, error) {
	ret := map[string]any{}
	var err error

	for fieldName, field := range schema.Properties {
		if ignoreHiden == true && field.Hiden == true {
			continue
		}

		if _, ok := dataIn[fieldName]; !ok && ignorenil {
			continue
		}

		value := openapi.getValueFromSchema(schema, fieldName, dataIn)
		/*
			if (field.type == "array" && field.items.type == "object") {
				if (Array.isArray(value) == true) {
					const list = ret[fieldName] = [];

					for (const item of value) {
						list.push(this.copyFields(field.items, item, ignorenil, ignoreHiden));
					}
				}
			} else if (field.type == "object") {
				ret[fieldName] = this.copyFields(field, value, ignorenil, ignoreHiden);
			} else {
		*/
		if value == nil && field.Nullable {
			ret[fieldName] = nil
		} else if value != nil {
			if ret[fieldName], err = openapi.copyValue(field, value); err != nil {
				return nil, err
			}
		}
		/*
			}
		*/
	}

	return ret, nil
}

/*
static getList(Qs, openapi, onlyClientUsage, roles) {
const process = properties => {
	for (let [fieldName, property] of Object.entries(properties)) {
		const $ref = property["x-$ref"];

		if ($ref != null) {
			let pos = $ref.indexOf("?");
			const queryObj = {"filter": {}};

			if (pos >= 0 && Qs != null) {
				const params = Qs.parse($ref.substring(pos), {ignoreQueryPrefix: true, allowDots: true});

				for (let [name, value] of Object.entries(params)) {
					if (value != null && value.startsWith("*") == true) queryObj.filter[name] = value.substring(1);
				}
			}

			const schemaName = OpenApi.getSchemaName($ref);
			const href = "#!/app/" + schemaName + "/search?" + Qs.stringify(queryObj, {allowDots: true});
			property["x-$ref"] = href;
		}
	}
}

const fillPropertiesRequired = schema => {
	if (schema.required == nil) return schema;

	for (const fieldName of schema.required) {
		if (schema.properties && schema.properties[fieldName] != nil) {
			schema.properties[fieldName]["x-required"] = true;
		}
	}

	return schema;
};

if (openapi == nil || openapi.components == nil || openapi.components.schemas == nil) return [];
const list = [];

for (const [schemaName, methods] of Object.entries(roles)) {
	for (const method in methods) {
		if (methods[method] == false) continue;
		const operationObject = this.getOperationObject(openapi, schemaName, method);
		if (operationObject == nil) continue;
		if (onlyClientUsage == true && operationObject.operationId.startsWith("zzz") == true) continue;
		const item = {operationId: operationObject.operationId, path: "/" + schemaName, method: method};
		const parameterSchema = OpenApi.getSchemaFromParameters(openapi, schemaName);
		const requestBodySchema = OpenApi.getSchemaFromRequestBodies(openapi, schemaName);
		const responseSchema = OpenApi.getSchemaFromSchemas(openapi, schemaName);
		if (parameterSchema != nil) item.parameter = parameterSchema.properties;

		if (requestBodySchema != nil) {
			item.requestBody = requestBodySchema.properties;
			process(item.requestBody);
			fillPropertiesRequired(requestBodySchema);
		}

		if (responseSchema != nil) {
			item.response = responseSchema.properties;
			process(item.response);
			fillPropertiesRequired(responseSchema);
		}

		list.push(item);
	}
}

return list;
}

static objToSchemaAdd(obj, schema, stringMayBeNumber) {
if (schema.properties == nil) schema.properties = {};
schema.count = schema.count == nil ? 1 : schema.count + 1;

for (let fieldName in obj) {
	let value = obj[fieldName];
	if (typeof value == "string") value = value.trim();
	let property = schema.properties[fieldName];

	if (property == nil) {
		property = schema.properties[fieldName] = {};
		property.mayBeNumber = true;
		property.mayBeInteger = true;
		property.mayBeDate = true;
		property.mayBeEmpty = false;
		property.nullable = false;
		property.maxLength = 0;
		property.default = value;
		property.count = 0;

		if (fieldName.startsWith("compet")) {
			property.pattern = "^20\\d\\d[01]\\d$";
			property.description = `${fieldName} deve estar no formato yyyymm`;
		}
	}

	property.count++;

	if (value == nil || value == null) {
		if (property.nullable == false) {
			property.nullable = true;

			if (["chv"].includes(fieldName) == true) {
				console.log(`${this.constructor.name}.objToSchemaAdd() : field ${fieldName} nullable`, obj);
			}
		}
	} else if (typeof value == "string" && value.length == 0) {
		if (property.mayBeEmpty == false) {
			property.mayBeEmpty = true;

			if (["chv"].includes(fieldName) == true) {
				console.log(`${this.constructor.name}.objToSchemaAdd() : field ${fieldName} mayBeEmpty`, obj);
			}
		}
	} else if (typeof value == "string" || typeof value == "number") {
		if (typeof value == "string") {
			if (property.maxLength < value.length) property.maxLength = value.length;
			if (property.mayBeDate == true && ((Date.parse(value) > 0) == false || value.includes("-") == false)) property.mayBeDate = false;

			if (property.mayBeNumber == true) {
				if (stringMayBeNumber != true || Number.isNaN(Number(value)) == true) {
					property.mayBeNumber = false;
					property.mayBeInteger = false;
				} else {
					if (property.mayBeInteger == true && value.includes(".") == true) property.mayBeInteger = false;
				}
			}
		} else if (typeof value == "number") {
			const strLen = value.toString().length;
			if (property.maxLength < strLen) property.maxLength = strLen;
			if (property.mayBeInteger == true && Number.isInteger(value) == false) property.mayBeInteger = false;
		}

		if (property.enum == nil) {
			property.enum = [];
			property.enumCount = [];
		}

		if (property.enum.length < 10) {
			const pos = property.enum.indexOf(value);

			if (pos < 0) {
				property.enum.push(value);
				property.enumCount.push(1);
			} else {
				property.enumCount[pos] = property.enumCount[pos] + 1;
			}
		}
	} else if (Array.isArray(value) == true) {
		property.type = "array";
		if (property.items == nil) property.items = {type:"object", properties:{}};
		for (const item of value) this.objToSchemaAdd(item, property.items, stringMayBeNumber);
	} else {
		property.type = "object";
		if (property.properties == nil) property.properties = {};
		this.objToSchemaAdd(value, property, stringMayBeNumber);
	}
}
}

static objToSchemaFinalize(schema, options) {
const adjustSchemaType = (schema) => {
	for (let [fieldName, property] of Object.entries(schema.properties)) {
		if (property.type == "object" && property.properties != nil) {
			adjustSchemaType(property);
			continue;
		}

		if (property.type == "array" && property.items != nil && property.items.properties != nil) {
			adjustSchemaType(property.items);
			continue;
		}

		if (property.type == nil) {
			if (property.mayBeInteger && property.maxLength > 0)
				property.type = "integer";
			else if (property.mayBeNumber && property.maxLength > 0)
				property.type = "number";
			else if (property.mayBeDate && property.maxLength > 0)
				property.type = "date-time";
			else
				property.type = "string";
		}
	}
}

const adjustRequired = (schema) => {
	if (schema.required == nil) schema.required = [];

	for (let [fieldName, property] of Object.entries(schema.properties)) {
		if (property.type == "object" && property.properties != nil) {
			adjustRequired(property);
			continue;
		}

		if (property.type == "array" && property.items != nil && property.items.properties != nil) {
			adjustRequired(property.items);
			property.required = property.items.required;
			continue;
		}

		if (property.count == schema.count) {
			if (property.essential == null) {
				property.essential = true;
				if (schema.required.includes(fieldName) == false) schema.required.push(fieldName);
			}

			if (property.nullable == false && property.mayBeEmpty == true) property.nullable = true;
		}
	}
}

const adjustSchemaEnumExampleDefault = (schema, enumMaxLength) => {
	for (let [fieldName, property] of Object.entries(schema.properties)) {
		if (property.type == "array" && property.items != nil && property.items.properties != nil) {
			adjustSchemaEnumExampleDefault(property.items, enumMaxLength);
			continue;
		}

		if (property.type == "object") {
			adjustSchemaEnumExampleDefault(property, enumMaxLength);
			continue;
		}

		if (property.enumCount != nil) {
			if (property.enumCount.length < enumMaxLength) {
				let posOfMax = 0;
				let countMax = -1;

				for (let i = 0; i < property.enumCount.length; i++) {
					const count = property.enumCount[i];

					if (count > countMax) {
						countMax = count;
						posOfMax = i;
					}
				}

				for (let i = 0; i < property.enum.length; i++) property.enum[i] = OpenApi.copyValue(property, property.enum[i]);
				property.default = property.enum[posOfMax];
			} else {
				property.example = property.enum.join(",");
				delete property.enum;
				delete property.enumCount;
			}
		}

		if (property.default != nil) property.default = OpenApi.copyValue(property, property.default);
	}
}

adjustSchemaType(schema);
adjustRequired(schema);
options = options || {};
options.enumMaxLength = options.enumMaxLength || 10
adjustSchemaEnumExampleDefault(schema, options.enumMaxLength);
schema.primaryKeys = options.primaryKeys || [];

for (const fieldName of schema.primaryKeys) {
	if (schema.properties[fieldName] == null)
		console.error(`${this.constructor.name}.objToSchemaFinalize() : invalid primaryKey : ${fieldName}, allowed values : `, Object.keys(schema.properties));
}
}

static genSchemaFromExamples(list, options) {
const schema = {type: "object", properties: {}};
for (let obj of list) this.objToSchemaAdd(obj, schema);
this.objToSchemaFinalize(schema, options);
return schema;
}
//{"methods": ["get", "post"], "schemas": responseSchemas, parameterSchemas, requestSchemas}
static fillOpenApi(openapi, options) {
const forceGeneratePath = options.requestSchemas == null && options.parameterSchemas == null;
if (openapi.paths == nil) openapi.paths = {};
if (openapi.components == nil) openapi.components = {};
if (openapi.components.schemas == nil) openapi.components.schemas = {};
if (openapi.components.responses == nil) openapi.components.responses = {};
if (openapi.components.parameters == nil) openapi.components.parameters = {};
if (openapi.components.requestBodies == nil) openapi.components.requestBodies = {};
if (openapi.tags == nil) openapi.tags = [];
//
if (options == nil) options = {};
if (options.requestBodyContentType == nil) options.requestBodyContentType = "application/json"
if (options.methods == nil) options.methods = ["get", "put", "post", "delete", "patch"];
if (options.parameterSchemas == nil) options.parameterSchemas = {};
if (options.requestSchemas == nil) options.requestSchemas = {};
if (options.responseSchemas == nil) options.responseSchemas = {};
if (options.security == nil) options.security = {};

if (options.requestSchemas["login"] == nil) {
	const requestSchema = {"type": "object", "properties": {"user": {type: "string"}, "password": {type: "string"}}, "required": ["user", "password"]};
	const responseSchema = {"type": "object", "properties": {"tokenPayload": {type: "string"}}, "required": ["tokenPayload"]};
	this.fillOpenApi(openapi, {methods: ["post"], requestSchemas: {"login": requestSchema}, schemas: {"login": responseSchema}, security: {"login": [{"basic": []}]}});
}

if (options.schemas == nil) {
	options.schemas = openapi.components.schemas;
} else {
	for (let [schemaName, schema] of Object.entries(options.schemas)) {
		openapi.components.schemas[schemaName] = schema;
	}
}
// add components/responses with error schema
const schemaError = {"type": "object", "properties": {"code": {"type": "integer"}, "description": {"type": "string"}}, "required": ["code", "description"]};
openapi.components.responses["Error"] = {"description": "Error response", "content": {"application/json": {"schema": schemaError}}};

for (const schemaName in options.schemas) {
	const requestSchema = options.requestSchemas[schemaName];
	const parameterSchema = options.parameterSchemas[schemaName];
	if (options.forceGenerateSchemas != true && forceGeneratePath == false && requestSchema == null && parameterSchema == null) continue;
	const schema = options.schemas[schemaName];
	if (schema.primaryKeys == nil) schema.primaryKeys = [];
	if (openapi.tags.find(item => item.name == schemaName) == nil) openapi.tags.push({"name": schemaName});
	const referenceToSchema = {"$ref": `#/components/schemas/${schemaName}`};
//			if (forceGeneratePath == false && requestSchema == null && parameterSchema == null) continue;
	// fill components/requestBody with schemas
	openapi.components.requestBodies[schemaName] = {"required": true, "content": {}};
	openapi.components.requestBodies[schemaName].content[options.requestBodyContentType] = {"schema": options.requestSchemas[schemaName] || referenceToSchema};
	// fill components/responses with schemas
	openapi.components.responses[schemaName] = {"description": "response", "content": {}};
	openapi.components.responses[schemaName].content[options.responseContentType] = {"schema": options.responseSchemas[schemaName] || referenceToSchema};
	// fill components/parameters with primaryKeys
	if (parameterSchema != null) {
		openapi.components.parameters[schemaName] = {"name": "main", "in": "query", "required": true, "schema": OpenApi.convertRufsToStandartSchema(parameterSchema)};
	} else if (schema.primaryKeys.length > 0) {
		const schemaPrimaryKey = {"type": "object", "properties": {}, "required": schema.primaryKeys};

		for (const primaryKey of schema.primaryKeys) {
			schemaPrimaryKey.properties[primaryKey] = OpenApi.getPropertyFromSchema(schema, primaryKey);
		}

		openapi.components.parameters[schemaName] = {"name": "primaryKey", "in": "query", "required": true, "schema": OpenApi.convertRufsToStandartSchema(schemaPrimaryKey)};
	}
	// path
	const pathName = `/${schemaName}`;
	const pathItemObject = openapi.paths[pathName] = {};
	const responsesRef = {"200": {"$ref": `#/components/responses/${schemaName}`}, "default": {"$ref": `#/components/responses/Error`}};
	const parametersRef = [{"$ref": `#/components/parameters/${schemaName}`}];
	const requestBodyRef = {"$ref": `#/components/requestBodies/${schemaName}`};

	const methods =                ["get", "put", "post", "delete", "patch"];
	const methodsHaveParameters =  [true , true , false , true    , true   ];
	const methodsHaveRequestBody = [false, true , true  , false   , true   ];

	for (let i = 0; i < methods.length; i++) {
		const method = methods[i];
		if (options.methods.includes(method) == false) continue;
		const operationObject = {};

		if (options.methods.length > 1) {
			operationObject.operationId = `zzz_${method}_${schemaName}`;
		} else {
			operationObject.operationId = schemaName;
		}

		if (methodsHaveParameters[i] == true && openapi.components.parameters[schemaName] != nil) operationObject.parameters = parametersRef;
		if (methodsHaveRequestBody[i] == true) operationObject.requestBody = requestBodyRef;
		operationObject.responses = responsesRef;
		operationObject.tags = [schemaName];
		operationObject.description = `CRUD ${method} operation over ${schemaName}`;
		if (options.security[schemaName] != nil) operationObject.security = options.security[schemaName];

		if (methodsHaveParameters[i] == false || operationObject.parameters != nil) {
			pathItemObject[method] = operationObject;
		}
	}
}

return openapi;
}
*/
func OpenApiGetSchemaName(ref string) string {
	ret := ref

	if pos := strings.LastIndex(ret, "/"); pos >= 0 {
		ret = ret[pos+1:]
	}

	if pos := strings.Index(ret, "?"); pos >= 0 {
		ret = ret[:pos]
	}

	return ret
}

func (openapi *OpenApi) getSchemaFromSchemas(ref string) (*Schema, bool) {
	schemaName := OpenApiGetSchemaName(ref)
	schema, ok := openapi.Components.Schemas[schemaName]
	return schema, ok
}

func (openapi *OpenApi) getSchemaFromRequestBodies(schemaName string) (schema *Schema, ok bool) {
	schemaName = OpenApiGetSchemaName(schemaName)
	requestBodyObject, ok := openapi.Components.RequestBodies[schemaName]

	if !ok {
		return nil, ok
	}

	for _, mediaTypeObject := range requestBodyObject.Content {
		if mediaTypeObject.Schema.Properties != nil {
			schema = mediaTypeObject.Schema
			break
		}
	}

	return schema, schema != nil
}

func (openapi *OpenApi) getSchemaFromParameters(schemaName string) (*Schema, error) {
	schemaName = OpenApiGetSchemaName(schemaName)
	parameterObject := openapi.Components.Parameters[schemaName]

	if parameterObject.Schema == nil {
		return nil, fmt.Errorf("[OpenApi.getSchemaFromParameters] don't find schema from parameter %s", schemaName)
	}

	return parameterObject.Schema, nil
}

/*
static getOperationObject(openapi, resource, method) {
let operationObject = nil;
const pathItemObject = openapi.paths["/" + resource];

if (pathItemObject != nil) {
	operationObject = pathItemObject[method.toLowerCase()];
}

return operationObject;
}
*/
func (openapi *OpenApi) getPropertyFromSchema(schema *Schema, propertyName string) (ret *Schema, ok bool) {
	if value, ok := schema.Properties[propertyName]; ok {
		return value, ok
	}

	for _, field := range schema.Properties {
		if field.InternalName == propertyName {
			ret = field
			break
		}
	}

	return ret, ret != nil
}

func (openapi *OpenApi) getPropertyFromSchemas(schemaName string, propertyName string) (field *Schema, ok bool) {
	schema, ok := openapi.getSchemaFromSchemas(schemaName)

	if ok {
		field, ok = openapi.getPropertyFromSchema(schema, propertyName)
	}

	return field, ok
}

func (openapi *OpenApi) getPropertyFromRequestBodies(schemaName string, propertyName string) (field *Schema, ok bool) {
	schema, ok := openapi.getSchemaFromRequestBodies(schemaName)

	if ok {
		field, ok = openapi.getPropertyFromSchema(schema, propertyName)
	}

	return field, ok
}

func (openapi *OpenApi) getProperty(schemaName string, propertyName string) (*Schema, bool) {
	schemaName = OpenApiGetSchemaName(schemaName)
	field, ok := openapi.getPropertyFromSchemas(schemaName, propertyName)

	if !ok {
		field, ok = openapi.getPropertyFromRequestBodies(schemaName, propertyName)
	}

	return field, ok
}

func (openapi *OpenApi) getPropertiesWithRef(schemaName string, ref string) (list []map[string]any, err error) {
	schemaName = OpenApiGetSchemaName(schemaName)

	processSchema := func(schema *Schema) {
		if schema == nil || schema.Properties == nil {
			return
		}

		for fieldName, field := range schema.Properties {
			if field.Ref != "" {
				if field.Ref == ref {
					found := false

					for _, item := range list {
						if item["fieldName"] == fieldName {
							found = true
							break
						}
					}

					if !found {
						list = append(list, map[string]any{"fieldName": fieldName, "field": field})
					}
				}
			}
		}
	}

	schema, ok := openapi.getSchemaFromSchemas(schemaName)

	if ok {
		processSchema(schema)
	}

	schema, ok = openapi.getSchemaFromRequestBodies(schemaName)

	if ok {
		processSchema(schema)
	}

	return list, err
}

/*
static getDependencies(openapi, schemaName, list, localSchemas) {
const processDependency = (schemaName, list) => {
	if (list.includes(schemaName) == false) {
		list.unshift(schemaName);
		this.getDependencies(openapi, schemaName, list, localSchemas);
	}
}

const processDependenciesFromSchema = (schema, list) => {
	if (schema == nil || schema.properties == nil) return;

	for (let [fieldName, field] of Object.entries(schema.properties)) {
		if (field.type == "array") {
			processDependenciesFromSchema(field.items, list);
		} else if (field.type == "object") {
			processDependenciesFromSchema(field, list);
		} else if (field.$ref != nil) {
			processDependency(this.getSchemaName(field.$ref), list);
		}
	}
}

schemaName = this.getSchemaName(schemaName);

if (list == nil)
	list = [];

if (localSchemas != nil && localSchemas[schemaName] != nil)
	processDependenciesFromSchema(localSchemas[schemaName], list);

let schema = this.getSchemaFromRequestBodies(openapi, schemaName);

if (schema != nil && schema.properties != nil)
	processDependenciesFromSchema(schema, list);

schema = this.getSchemaFromSchemas(openapi, schemaName);

if (schema != nil && schema.properties != nil)
	processDependenciesFromSchema(schema, list);

return list;
}

static getDependenciesSchemas(openapi, schema) {
const list = [];
// TODO : varrer todos os schema.properties e adicionar na lista os property.properties que não se repetem
return list;
}

static getDependents(openapi, schemaNameTarget, onlyInDocument, localSchemas) {
const processSchema = (schema, schemaName, schemaNameTarget, onlyInDocument, list) => {
	if (schema == null || schema.properties == null) return;

	for (let [fieldName, field] of Object.entries(schema.properties)) {
		if (field.$ref != null) {
			let found = false;
			if (field.$ref == schemaNameTarget || this.getSchemaName(field.$ref) == schemaNameTarget) found = true;

			if (found == true && (onlyInDocument != true || field.type == "object")) {
				if (list.find(item => item.table == schemaName && item.field == fieldName) == null) {
					list.push({"table": schemaName, "field": fieldName})
				}
			}
		}
	}
}

schemaNameTarget = this.getSchemaName(schemaNameTarget);
const list = [];

if (localSchemas) {
	for (let [schemaName, schema] of Object.entries(localSchemas)) {
		processSchema(schema, schemaName, schemaNameTarget, onlyInDocument, list);
	}
}

for (let [schemaName, requestBodyObject] of Object.entries(openapi.components.requestBodies)) {
	for (const [mediaTypeName, mediaTypeObject] of Object.entries(requestBodyObject.content)) {
		processSchema(mediaTypeObject.schema, schemaName, schemaNameTarget, onlyInDocument, list);
	}
}

return list;
}

static resolveSchema(propertyName, schema, openapi, localSchemas) {
if (typeof(schema) == "string") {
	const schemaName = this.getSchemaName(schema);
	let field;

	if (localSchemas && localSchemas[schemaName] && localSchemas[schemaName].properties && OpenApi.getPropertyFromSchema(localSchemas[schemaName], propertyName) != nil)
		return localSchemas[schemaName];

	if (OpenApi.getPropertyFromSchemas(openapi, schemaName, propertyName) != nil)
		return this.getSchemaFromSchemas(openapi, schemaName);

	return OpenApi.getSchemaFromRequestBodies(openapi, schemaName);
} else if (schema.properties != nil) {
	return schema;
}

return schema;
}
*/
type ForeignKeyDescription struct {
	TableRef    string
	FieldsRef   map[string]any
	IsUniqueKey bool
}

// (service, (service.field|foreignTableName)
func (openapi *OpenApi) getForeignKeyDescription(schema string, fieldName string) (*ForeignKeyDescription, error) {
	field, ok := openapi.getProperty(schema, fieldName)

	if !ok {
		return nil, fmt.Errorf("[OpenApi.getForeignKeyDescription] : Missing property %s from schema %s", fieldName, schema)
	}

	if field.Ref == "" {
		return nil, nil
	}

	serviceRef, ok := openapi.getSchemaFromSchemas(field.Ref)

	if !ok {
		return nil, fmt.Errorf("[OpenApi.getForeignKeyDescription] : Missing schema %s", field.Ref)
	}
	/*
		pos = field.ref.indexOf("?");

		if (pos >= 0 && Qs != nil) {
			const fieldsRef = Qs.parse(field.$ref.substring(pos), {ignoreQueryPrefix: true, allowDots: true});
			const entries = Object.entries(fieldsRef);
			let isUniqueKey = entries.length == serviceRef.primaryKeys.length;

			if (isUniqueKey == true) {
				for (let [fieldName, fieldNameMap] of entries) {
					if (serviceRef.primaryKeys.indexOf(fieldName) < 0) {
						isUniqueKey = false;
						break;
					}
				}
			}

			const ret = {tableRef: field.$ref, fieldsRef: fieldsRef, isUniqueKey: isUniqueKey};
			return ret;
		}
	*/
	fieldsRef := map[string]any{}

	for _, primaryKey := range serviceRef.PrimaryKeys {
		fieldsRef[primaryKey] = nil
	}

	if len(fieldsRef) == 1 {
		for primaryKey := range fieldsRef {
			fieldsRef[primaryKey] = fieldName
		}
	} else if len(fieldsRef) > 1 {
		for fieldRef := range fieldsRef {
			if _, ok := openapi.getProperty(field.Ref, fieldRef); fieldsRef[fieldRef] == nil && ok {
				fieldsRef[fieldRef] = fieldRef
			}
		}

		for fieldRef := range fieldsRef {
			if fieldsRef[fieldRef] == "id" {
				fieldsRef[fieldRef] = fieldName
			}
		}
	}

	for fieldRef, value := range fieldsRef {
		if value == nil {
			return nil, fmt.Errorf("[OpenApi.getForeignKeyDescription(%s, %s)] : don't pair with key %s : %s", schema, fieldName, fieldRef, fieldsRef)
		}
	}

	ret := ForeignKeyDescription{TableRef: field.Ref, FieldsRef: fieldsRef, IsUniqueKey: true}
	return &ret, nil
}

func (openapi *OpenApi) getForeignKeyEntries(serviceName string, ref string) (list []map[string]any, err error) {
	return openapi.getPropertiesWithRef(serviceName, ref)
}

/*
static getForeignKey(openapi, schema, fieldName, obj, localSchemas) {
if (fieldName == "CpfCnpj" && obj.cpfCnpj != null)
	console.log(`[${this.constructor.name}.getPrimaryKeyForeign(${fieldName})] : obj :`, obj);

const foreignKeyDescription = OpenApi.getForeignKeyDescription(openapi, schema, fieldName, localSchemas);

if (foreignKeyDescription == nil)
	return nil;

let key = {};

for (let [fieldRef, field] of Object.entries(foreignKeyDescription.fieldsRef)) {
	key[field] = obj[fieldRef];
}

schema = this.resolveSchema(fieldName, schema, openapi, localSchemas);
key = this.copyFields(schema, key);
console.log(`[${this.constructor.name}.getForeignKey(${fieldName})] : obj :`, obj, "key :", key);
return key;
}
*/
// (service, (service.field|foreignTableName), service.obj) => [{name: constraintName, table: foreignTableName, foreignKey: {}}]
type PrimaryKeyForeign struct {
	Table       string
	PrimaryKey  map[string]any
	Valid       bool
	IsUniqueKey bool
}

func (openapi *OpenApi) getPrimaryKeyForeign(schemaName string, fieldName string, obj map[string]any) (ret *PrimaryKeyForeign, err error) {
	process := func(schema *Schema, fieldName string, obj map[string]any) (*PrimaryKeyForeign, error) {
		if schema == nil {
			return nil, nil
		}

		if schema.Properties == nil {
			return nil, nil
		}

		foreignKeyDescription, err := openapi.getForeignKeyDescription(schemaName, fieldName)

		if err != nil {
			return nil, err
		}

		key := map[string]any{}
		ret := PrimaryKeyForeign{Table: foreignKeyDescription.TableRef, PrimaryKey: key, Valid: false, IsUniqueKey: foreignKeyDescription.IsUniqueKey}

		if obj == nil {
			return &ret, nil
		}

		valid := true

		for fieldRef, fieldNameMap := range foreignKeyDescription.FieldsRef {
			switch fieldNameMap.(type) {
			case string:
				var value any

				if strings.HasPrefix(fieldNameMap.(string), "*") {
					value = fieldNameMap.(string)[1:]
				} else {
					value = openapi.getValueFromSchema(schema, fieldNameMap.(string), obj)
				}

				key[fieldRef] = value

				if value == "" {
					valid = false
				}
			default:
				valid = false
			}
		}

		ret.Valid = valid
		return &ret, nil
	}

	schemaName = OpenApiGetSchemaName(schemaName)

	if schema, ok := openapi.getSchemaFromRequestBodies(schemaName); ok {
		ret, err = process(schema, fieldName, obj)
	}

	if schema, ok := openapi.getSchemaFromSchemas(schemaName); ok {
		ret, err = process(schema, fieldName, obj)
	}

	return ret, err
}
