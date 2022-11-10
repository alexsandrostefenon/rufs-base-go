package rufsBase

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type OpenApiSecurity struct {
}

type ContactObject struct {
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type InfoObject struct {
	Title       string        `json:"title"`
	Version     string        `json:"version"`
	Description string        `json:"description,omitempty"`
	Contact     ContactObject `json:"contact,omitempty"`
}

type ServerObject struct {
	Url string `json:"url"`
}

type ParameterObject struct {
	Ref         string  `json:"$ref,omitempty"`
	Name        string  `json:"name,omitempty"`
	In          string  `json:"in,omitempty"`
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

type OperationObject struct {
	Tags        []string                    `json:"tags,omitempty"`
	Description string                      `json:"description,omitempty"`
	OperationId string                      `json:"operationId,omitempty"`
	Parameters  []ParameterObject           `json:"parameters,omitempty"`
	RequestBody *RequestBodyObject          `json:"requestBody,omitempty"`
	Responses   map[string]*ResponseObject  `json:"responses,omitempty"`
	Security    []SecurityRequirementObject `json:"security,omitempty"`
}

type ForeignKey struct {
	TableRef  string   `json:"tableRef"`
	Fields    []string `json:"fields"`
	FieldsRef []string `json:"fieldsRef"`
}

type Schema struct {
	Name               string                `json:"-"`
	PrimaryKeys        []string              `json:"x-primaryKeys,omitempty"`
	UniqueKeys         map[string][]string   `json:"x-uniqueKeys,omitempty"`
	ForeignKeys        map[string]ForeignKey `json:"x-foreignKeys,omitempty"`
	Required           []string              `json:"required,omitempty"`
	Ref                string                `json:"$ref,omitempty"`
	Type               string                `json:"type,omitempty"`
	Format             string                `json:"format,omitempty"`
	Description        string                `json:"description,omitempty"`
	Nullable           bool                  `json:"nullable,omitempty"`
	Essential          bool                  `json:"x-required,omitempty"`
	Title              string                `json:"x-title,omitempty"`
	Hiden              bool                  `json:"x-hiden,omitempty"`
	InternalName       string                `json:"x-internalName,omitempty"`
	Default            string                `json:"default,omitempty"`
	Enum               []any                 `json:"enum,omitempty"`
	EnumLabels         []string              `json:"x-enumLabels,omitempty"`
	IdentityGeneration string                `json:"x-identityGeneration,omitempty"`
	Updatable          bool                  `json:"x-updatable,omitempty"`
	Scale              int                   `json:"x-scale,omitempty"`
	Precision          int                   `json:"x-precision,omitempty"`
	MaxLength          int                   `json:"maxLength,omitempty"`
	Properties         map[string]*Schema    `json:"properties,omitempty"`
	Items              *Schema               `json:"items,omitempty"`
}

type MediaTypeObject struct {
	Schema *Schema `json:"schema"`
}

type RequestBodyObject struct {
	Required bool                        `json:"required,omitempty"`
	Ref      string                      `json:"$ref,omitempty"`
	Content  map[string]*MediaTypeObject `json:"content,omitempty"`
}

type ResponseObject struct {
	Description string                      `json:"description,omitempty"`
	Ref         string                      `json:"$ref,omitempty"`
	Content     map[string]*MediaTypeObject `json:"content,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

type TagObject struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PathItemObject map[string]*OperationObject

type SecurityRequirementObject map[string][]string

type OpenApi struct {
	Openapi    string                    `json:"openapi"`
	Info       *InfoObject               `json:"info"`
	Servers    []*ServerObject           `json:"servers,omitempty"`
	Paths      map[string]PathItemObject `json:"paths"`
	Components struct {
		Schemas         map[string]*Schema           `json:"schemas,omitempty"`
		Parameters      map[string]*ParameterObject  `json:"parameters,omitempty"`
		RequestBodies   map[string]RequestBodyObject `json:"requestBodies,omitempty"`
		Responses       map[string]ResponseObject    `json:"responses,omitempty"`
		SecuritySchemes map[string]SecurityScheme    `json:"securitySchemes,omitempty"`
	} `json:"components,omitempty"`
	Security []SecurityRequirementObject `json:"security,omitempty"`
	Tags     []TagObject                 `json:"tags,omitempty"`
}

func OpenApiCreate(openapi *OpenApi, security string) {
	if openapi.Openapi == "" {
		openapi.Openapi = "3.0.3"
	}

	if openapi.Info == nil {
		openapi.Info = &InfoObject{Title: "rufs-base-es6 openapi genetator", Version: "0.0.0", Description: "CRUD operations", Contact: ContactObject{Name: "API Support", Url: "http://www.example.com/support", Email: "support@example.com"}}
	}

	if openapi.Paths == nil {
		openapi.Paths = map[string]PathItemObject{}
	}

	if openapi.Components.Schemas == nil {
		openapi.Components.Schemas = map[string]*Schema{}
	}

	if openapi.Components.Parameters == nil {
		openapi.Components.Parameters = map[string]*ParameterObject{}
	}

	if openapi.Components.RequestBodies == nil {
		openapi.Components.RequestBodies = map[string]RequestBodyObject{}
	}

	if openapi.Components.Responses == nil {
		openapi.Components.Responses = map[string]ResponseObject{}
	}

	if openapi.Components.SecuritySchemes == nil {
		openapi.Components.SecuritySchemes = map[string]SecurityScheme{
			"jwt":    {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
			"apiKey": {Type: "apiKey", In: "header", Name: "X-API-KEY"},
			"basic":  {Type: "http", Scheme: "basic"},
		}
	}

	if openapi.Security == nil && len(security) > 0 {
		openapi.Security = []SecurityRequirementObject{{security: []string{}}}
	}
}

func (source *OpenApi) copy(paths []string) *OpenApi {
	dest := &OpenApi{}
	/*
		dest.openapi = source.openapi;
		dest.info = source.info;
		dest.servers = source.servers;
		dest.components.securitySchemes = source.components.securitySchemes;
		dest.security = source.security;
		dest.tags = source.tags;

		for (let [schemaName, role] of Object.entries(paths)) {
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
	*/
	return dest
}

/*
func MergeSchemas(schemaOld *Schema, schemaNew *Schema, keepOld bool, schemaName string) *Schema {
	mergeArrayString := func(oldArray []string, newArray []string) []string {
		if len(newArray) == 0 {
			return oldArray
		}

		if len(oldArray) == 0 {
			return newArray
		}

		for _, itemNew := range newArray {
			if idx := slices.Index(oldArray, itemNew); idx < 0 {
				oldArray = append(oldArray, itemNew)
			}
		}

		return oldArray
	}

	mergeArray := func(oldArray []any, newArray []any) []any {
		if len(newArray) == 0 {
			return oldArray
		}

		if len(oldArray) == 0 {
			return newArray
		}

		for _, itemNew := range newArray {
			found := false

			for _, itemOld := range oldArray {
				if itemOld == itemNew {
					found = true
					break
				}
			}

			if !found {
				oldArray = append(oldArray, itemNew)
			}
		}

		return oldArray
	}

	if schemaOld == nil {
		schemaOld = &Schema{}
	}
	//		console.log(`[${this.constructor.name}.updateJsonSchema(schemaName: ${schemaName}, schemaNew.properties: ${schemaNew.properties}, schemaOld.properties: ${schemaOld.properties})]`);
	jsonSchemaTypes := []string{"boolean", "string", "integer", "number", "date-time", "date", "object", "array"}
	jsonBuilder := map[string]*Schema{}

	if keepOld {
		jsonBuilder = schemaOld.Properties
	}

	if schemaOld.Properties == nil {
		schemaOld.Properties = map[string]*Schema{}
	}

	if schemaNew.Properties == nil {
		schemaNew.Properties = map[string]*Schema{}
	}

	oldFields := schemaOld.Properties
	newFields := schemaNew.Properties

	for fieldName, field := range schemaNew.Properties {
		if field.Type == "" {
			field.Type = "string"
		}

		if field.IdentityGeneration != "" {
			field.Hiden = true
			//			field.ReadOnly = true
		}

		if slices.Index(jsonSchemaTypes, field.Type) < 0 {
			//				console.error(`${schemaName} : ${fieldName} : Unknow type : ${field.type}`);
			continue
		}
		// type (columnDefinition), readOnly, hiden, primaryKey, essential (insertable), updatable, default, length, precision, scale
		jsonBuilderValue := &Schema{}
		// registra conflitos dos valores antigos com os valores detectados do banco de dados
		jsonBuilderValue.Type = field.Type
		jsonBuilderValue.Format = field.Format

		if field.Updatable == false {
			jsonBuilderValue.Updatable = false
		}

		if field.MaxLength > 0 {
			jsonBuilderValue.MaxLength = field.MaxLength
		}

		if field.Precision > 0 {
			jsonBuilderValue.Precision = field.Precision
		}

		if field.Scale > 0 {
			jsonBuilderValue.Scale = field.Scale
		}

		jsonBuilderValue.Nullable = field.Nullable
		//
		if field.Ref != "" {
			jsonBuilderValue.Ref = field.Ref
		}

		if len(field.Properties) > 0 {
			jsonBuilderValue.Properties = field.Properties
		}

		if field.Items != nil {
			jsonBuilderValue.Items = field.Items
		}

		if field.InternalName != "" {
			jsonBuilderValue.InternalName = field.InternalName
		}

		if field.Essential {
			jsonBuilderValue.Essential = field.Essential
		}

		if field.Default != "" {
			jsonBuilderValue.Default = field.Default
		}
		// if field.Unique {
		// 	jsonBuilderValue.Unique = field.Unique
		// }
		if field.IdentityGeneration != "" {
			jsonBuilderValue.IdentityGeneration = field.IdentityGeneration
		}
		// if field.IsClonable == false {
		// 	jsonBuilderValue.IsClonable = field.IsClonable
		// }
		if field.Hiden {
			jsonBuilderValue.Hiden = field.Hiden
		}
		// if field.ReadOnly {
		// 	jsonBuilderValue.ReadOnly = field.ReadOnly
		// }
		if field.Description != "" {
			jsonBuilderValue.Description = field.Description
		}
		// oculta tipos incompatíveis
		if jsonBuilderValue.Type != "string" {
			jsonBuilderValue.MaxLength = 0
		}

		if jsonBuilderValue.Type != "number" {
			jsonBuilderValue.Precision = 0
			jsonBuilderValue.Scale = 0
		}

		if jsonBuilderValue.Type != "object" {
			jsonBuilderValue.Properties = map[string]*Schema{}
		}

		if jsonBuilderValue.Type != "array" {
			jsonBuilderValue.Items = nil
		}
		// habilita os campos PLENAMENTE não SQL
		if field.Title != "" {
			jsonBuilderValue.Title = field.Title
		}
		// if field.Document {
		// 	jsonBuilderValue.Document = field.Document
		// }

		// if field.SortType != "" {
		// 	jsonBuilderValue.SortType = field.SortType
		// }

		// if field.OrderIndex > 0 {
		// 	jsonBuilderValue.OrderIndex = field.OrderIndex
		// }

		// if field.TableVisible == false {
		// 	jsonBuilderValue.TableVisible = field.TableVisible
		// }

		// if field.ShortDescription != "" {
		// 	jsonBuilderValue.ShortDescription = field.ShortDescription
		// }
		if len(field.Enum) > 0 {
			jsonBuilderValue.Enum = mergeArray(jsonBuilderValue.Enum, field.Enum)
		}

		if len(field.EnumLabels) > 0 {
			jsonBuilderValue.EnumLabels = mergeArrayString(jsonBuilderValue.EnumLabels, field.EnumLabels)
		}
		// exceções
		if oldFields[fieldName] != nil {
			fieldOriginal := oldFields[fieldName]
			// copia do original os campos PLENAMENTE não SQL
			jsonBuilderValue.Title = fieldOriginal.Title
			// jsonBuilderValue.Document = fieldOriginal.Document
			// jsonBuilderValue.SortType = fieldOriginal.SortType
			// jsonBuilderValue.OrderIndex = fieldOriginal.OrderIndex
			// jsonBuilderValue.TableVisible = fieldOriginal.TableVisible
			// jsonBuilderValue.ShortDescription = fieldOriginal.ShortDescription
			jsonBuilderValue.Enum = mergeArray(jsonBuilderValue.Enum, fieldOriginal.Enum)
			jsonBuilderValue.EnumLabels = mergeArrayString(jsonBuilderValue.EnumLabels, fieldOriginal.EnumLabels)
			// registra conflitos dos valores antigos com os valores detectados do banco de dados
			// exceptions := []string{"service", "isClonable", "hiden", "$ref"}

			// for (let subFieldName in fieldOriginal) {
			// 	if (exceptions.indexOf(subFieldName) < 0 && fieldOriginal[subFieldName] != jsonBuilderValue[subFieldName]) {
			// 		console.warn(`generateJsonSchema() : table [${schemaName}], field [${fieldName}], property [${subFieldName}] conflict previous declared [${fieldOriginal[subFieldName]}] new [${jsonBuilderValue[subFieldName]}]\nold:\n`, fieldOriginal, "\nnew:\n", jsonBuilderValue);
			// 	}
			// }
			// copia do original os campos PARCIALMENTE não SQL
			// if fieldOriginal.IsClonable == false {
			// 	jsonBuilderValue.IsClonable = fieldOriginal.IsClonable
			// }

			// if fieldOriginal.ReadOnly {
			// 	jsonBuilderValue.ReadOnly = fieldOriginal.ReadOnly
			// }
			if fieldOriginal.Hiden == false {
				jsonBuilderValue.Hiden = fieldOriginal.Hiden
			}
		}

		if old, ok := oldFields[fieldName]; jsonBuilderValue.Type == "array" && ok {
			jsonBuilder[fieldName] = &Schema{}
			jsonBuilder[fieldName].Items = MergeSchemas(old.Items, newFields[fieldName].Items, keepOld, schemaName)
		} else if jsonBuilderValue.Type == "object" && ok {
			jsonBuilder[fieldName] = MergeSchemas(old, newFields[fieldName], keepOld, schemaName)
		} else {
			jsonBuilder[fieldName] = jsonBuilderValue
		}
	}

	schema := &Schema{}

	if schemaOld.Title != "" {
		schema.Title = schemaOld.Title
	} else {
		schema.Title = schemaNew.Title
	}

	schema.Type = "object"
	schema.Required = []string{}
	schema.PrimaryKeys = schemaNew.PrimaryKeys
	schema.UniqueKeys = schemaNew.UniqueKeys
	schema.ForeignKeys = schemaNew.ForeignKeys
	schema.Properties = jsonBuilder
	for fieldName, field := range schema.Properties {
		if field.Essential {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	return schema
}
*/
/*
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

	dataType := field.Type

	if field.Format != "" {
		dataType = field.Format
	}

	if dataType == "" || dataType == "string" {
		switch value.(type) {
		case string:
			if len(value.(string)) > field.MaxLength {
				ret = value.(string)[:field.MaxLength]
			} else {
				ret = value
			}
		default:
			ret = value
		}
	} else if dataType == "integer" {
		switch value.(type) {
		case string:
			ret, err = strconv.Atoi(value.(string))
		default:
			ret = value
		}
	} else if dataType == "number" {
		switch value.(type) {
		case string:
			ret, err = strconv.ParseFloat(value.(string), 64)
		default:
			ret = value
		}
	} else if dataType == "boolean" {
		switch value.(type) {
		case bool:
			ret = value.(bool)
		case string:
			ret = (value == "true")
		}
	} else if dataType == "date-time" {
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
	} else if dataType == "date" {
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

func (openapi *OpenApi) copyFields(schema *Schema, dataIn map[string]any, ignorenil bool, ignoreHiden bool, onlyPrimaryKeys bool) (map[string]any, error) {
	ret := map[string]any{}
	var err error

	for fieldName, field := range schema.Properties {
		if ignoreHiden == true && field.Hiden == true {
			continue
		}

		if _, ok := dataIn[fieldName]; !ok && ignorenil {
			continue
		}

		if onlyPrimaryKeys && slices.Index(schema.PrimaryKeys, fieldName) < 0 {
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
*/
type FillOpenApiOptions struct {
	forceGenerateSchemas   bool
	requestBodyContentType string
	responseContentType    string
	methods                []string
	parameterSchemas       map[string]*Schema
	requestSchemas         map[string]*Schema
	responseSchemas        map[string]*Schema
	disableResponseList    map[string]bool
	schemas                map[string]*Schema
	security               SecurityRequirementObject
}

func (openapi *OpenApi) FillOpenApi(options FillOpenApiOptions) {
	OpenApiCreate(openapi, "jwt")
	forceGeneratePath := options.requestSchemas == nil && options.parameterSchemas == nil

	if options.requestBodyContentType == "" {
		options.requestBodyContentType = "application/json"
	}

	if options.responseContentType == "" {
		options.responseContentType = "application/json"
	}

	if len(options.methods) == 0 {
		options.methods = []string{"get", "put", "post", "delete", "patch"}
	}

	if options.requestSchemas["login"] == nil {
		requestSchema := &Schema{}
		json.Unmarshal([]byte(`{"type": "object", "properties": {"user": {type: "string"}, "password": {type: "string"}}, "required": ["user", "password"]}`), requestSchema)
		responseSchema := &Schema{}
		json.Unmarshal([]byte(`{"type": "object", "properties": {"tokenPayload": {type: "string"}}, "required": ["tokenPayload"]}`), responseSchema)
		loginOptions := FillOpenApiOptions{methods: []string{"post"}, requestSchemas: map[string]*Schema{"login": requestSchema}, schemas: map[string]*Schema{"login": responseSchema}}
		openapi.FillOpenApi(loginOptions)
	}

	if len(options.schemas) == 0 {
		options.schemas = openapi.Components.Schemas
	} else {
		for schemaName, schema := range options.schemas {
			openapi.Components.Schemas[schemaName] = schema
		}
	}
	// add components/responses with error schema
	schemaError := &Schema{}
	json.Unmarshal([]byte(`{"type": "object", "properties": {"code": {"type": "integer"}, "description": {"type": "string"}}, "required": ["code", "description"]}`), schemaError)

	openapi.Components.Responses["Error"] = ResponseObject{Description: "Error response", Content: map[string]*MediaTypeObject{"application/json": {Schema: schemaError}}}

	for schemaName, schema := range options.schemas {
		parameterSchema := options.parameterSchemas[schemaName]
		requestSchema := options.requestSchemas[schemaName]
		responseSchema := options.responseSchemas[schemaName]
		disableResponseList := options.disableResponseList[schemaName]

		if !options.forceGenerateSchemas && !forceGeneratePath && requestSchema == nil && parameterSchema == nil {
			continue
		}

		if slices.IndexFunc(openapi.Tags, func(item TagObject) bool { return item.Name == schemaName }) < 0 {
			openapi.Tags = append(openapi.Tags, TagObject{Name: schemaName})
		}

		referenceToSchema := &Schema{Ref: fmt.Sprintf("#/components/schemas/%s", schemaName)}
		// fill components/requestBody with schemas
		openapi.Components.RequestBodies[schemaName] = RequestBodyObject{Required: true, Content: map[string]*MediaTypeObject{}}

		if requestSchema != nil && requestSchema.Type != "" {
			openapi.Components.RequestBodies[schemaName].Content[options.requestBodyContentType] = &MediaTypeObject{Schema: requestSchema}
		} else {
			openapi.Components.RequestBodies[schemaName].Content[options.requestBodyContentType] = &MediaTypeObject{Schema: referenceToSchema}
		}
		// fill components/responses with schemas
		openapi.Components.Responses[schemaName] = ResponseObject{Description: "response", Content: map[string]*MediaTypeObject{"application/json": {Schema: referenceToSchema}}}

		if !disableResponseList {
			openapi.Components.Responses[schemaName+"List"] = ResponseObject{Description: "response list", Content: map[string]*MediaTypeObject{"application/json": {Schema: &Schema{Type: "array", Items: referenceToSchema}}}}
		}

		if requestSchema != nil && requestSchema.Type != "" {
			openapi.Components.RequestBodies[schemaName].Content[options.responseContentType] = &MediaTypeObject{Schema: responseSchema}
		} else {
			openapi.Components.RequestBodies[schemaName].Content[options.responseContentType] = &MediaTypeObject{Schema: referenceToSchema}
		}
		// fill components/parameters with primaryKeys
		if parameterSchema != nil {
			openapi.Components.Parameters[schemaName] = &ParameterObject{Name: "main", In: "query", Required: true, Schema: parameterSchema}
		} else if len(schema.PrimaryKeys) > 0 {
			schemaPrimaryKey := Schema{Type: "object", Required: schema.PrimaryKeys, Properties: map[string]*Schema{}}

			for _, primaryKey := range schema.PrimaryKeys {
				schemaPrimaryKey.Properties[primaryKey] = schema.Properties[primaryKey]
			}

			parameterObject := &ParameterObject{Name: "primaryKey", In: "query", Required: true, Schema: &schemaPrimaryKey}
			openapi.Components.Parameters[schemaName] = parameterObject
		}
		// path
		pathName := fmt.Sprintf("/%s", CamelToUnderscore(schemaName))
		pathItemObject := PathItemObject{}
		openapi.Paths[pathName] = pathItemObject
		responsesRefOk := &ResponseObject{Ref: fmt.Sprintf("#/components/responses/%s", schemaName)}
		responsesRefOkList := &ResponseObject{Ref: fmt.Sprintf("#/components/responses/%sList", schemaName)}
		responsesRefError := &ResponseObject{Ref: `#/components/responses/Error`}
		parametersRef := []ParameterObject{{Ref: fmt.Sprintf(`#/components/parameters/%s`, schemaName)}}
		requestBodyRef := RequestBodyObject{Ref: fmt.Sprintf(`#/components/requestBodies/%s`, schemaName)}

		methods := []string{"get", "put", "post", "delete", "patch"}
		methodsHaveParameters := []bool{true, true, false, true, true}
		methodsHaveRequestBody := []bool{false, true, true, false, true}
		methodsHaveResponseList := []bool{true, false, false, false, false}

		for i, method := range methods {
			if slices.Index(options.methods, method) < 0 {
				continue
			}

			operationObject := &OperationObject{}

			if len(options.methods) > 1 {
				operationObject.OperationId = fmt.Sprintf("zzz_%s_%s", method, schemaName)
			} else {
				operationObject.OperationId = schemaName
			}

			if methodsHaveParameters[i] && openapi.Components.Parameters[schemaName] != nil {
				operationObject.Parameters = parametersRef
			}

			if methodsHaveRequestBody[i] {
				operationObject.RequestBody = &requestBodyRef
			}

			if methodsHaveResponseList[i] && !disableResponseList {
				operationObject.Responses = map[string]*ResponseObject{"200": responsesRefOkList, "default": responsesRefError}
			} else {
				operationObject.Responses = map[string]*ResponseObject{"200": responsesRefOk, "default": responsesRefError}
			}

			operationObject.Tags = []string{schemaName}
			operationObject.Description = fmt.Sprintf(`CRUD %s operation over %s`, method, schemaName)
			operationObject.Security = []SecurityRequirementObject{}

			if options.security != nil {
				operationObject.Security = append(operationObject.Security, options.security)
			}

			if !methodsHaveParameters[i] || operationObject.Parameters != nil {
				pathItemObject[method] = operationObject
			}
		}
	}
}

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

func (openapi *OpenApi) getSchemaFromResponses(schemaName string) (schema *Schema, ok bool) {
	schemaName = OpenApiGetSchemaName(schemaName)
	responseObject, ok := openapi.Components.Responses[schemaName]

	if !ok {
		return nil, ok
	}

	for _, mediaTypeObject := range responseObject.Content {
		if mediaTypeObject.Schema.Ref != "" {
			schema, _ = openapi.getSchemaFromRef(mediaTypeObject.Schema.Ref)
			break
		} else if mediaTypeObject.Schema.Properties != nil || mediaTypeObject.Schema.Items != nil {
			schema = mediaTypeObject.Schema
			break
		}
	}

	return schema, schema != nil
}

func (openapi *OpenApi) getSchemaFromRef(ref string) (schema *Schema, err error) {
	schemaName := OpenApiGetSchemaName(ref)

	if strings.HasPrefix(ref, "#/components/parameters/") {
		if parameterObject, ok := openapi.Components.Parameters[schemaName]; ok {
			schema = parameterObject.Schema
		}
	} else if strings.HasPrefix(ref, "#/components/schemas/") {
		schema = openapi.Components.Schemas[schemaName]
	} else if strings.HasPrefix(ref, "#/components/responses/") {
		if response, ok := openapi.Components.Responses[schemaName]; ok {
			for _, content := range response.Content {
				schema = content.Schema
				break
			}
		}
	}

	if schema == nil {
		err = fmt.Errorf("[OpenApi.getSchemaFromParameters] don't find schema from %s", ref)
	} else if schema.Name == "" {
		schema.Name = schemaName
	}

	return schema, err
}

func (openapi *OpenApi) getPathParams(uri string, params map[string]any) (path string, err error) {
	uriSegments := strings.Split(uri, "/")

	for pattern := range openapi.Paths {
		pathSegments := strings.Split(pattern, "/")

		if len(uriSegments) == len(pathSegments) {
			match := true

			for idx, pathSegment := range pathSegments {
				if strings.HasPrefix(pathSegment, "{") && strings.HasSuffix(pathSegment, "}") {
					name := pathSegment[1 : len(pathSegment)-1]
					params[name] = uriSegments[idx]
				} else if pathSegment != uriSegments[idx] {
					match = false
					break
				}
			}

			if match {
				path = pattern
				break
			}
		}
	}

	return path, err
}

func (openapi *OpenApi) getSchemaFromParameters(path string, method string) (*Schema, error) {
	for pattern, pathItemObject := range openapi.Paths {
		if pattern == path {
			if operationObject, ok := pathItemObject[method]; ok {
				for _, parameterObject := range operationObject.Parameters {
					if parameterObject.Ref != "" {
						return openapi.getSchemaFromRef(parameterObject.Ref)
					} else if parameterObject.Schema != nil {
						if parameterObject.Schema.Ref != "" {
							return openapi.getSchemaFromRef(parameterObject.Schema.Ref)
						} else {
							return parameterObject.Schema, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("[OpenApi.getSchemaFromParameters] don't find schema parameter from %s", path)
}

func (openapi *OpenApi) getSchema(path string, method string, _type string) (schema *Schema, err error) {
	getSchemaFromContent := func(content map[string]*MediaTypeObject) (schema *Schema, err error) {
		for _, mediaTypeObject := range content {
			if mediaTypeObject.Schema.Ref != "" {
				schema, err = openapi.getSchemaFromRef(mediaTypeObject.Schema.Ref)
			} else {
				schema = mediaTypeObject.Schema
			}
		}

		return schema, err
	}

	if pathItemObject, ok := openapi.Paths[path]; ok {
		method = strings.ToLower(method)

		if operationObject, ok := pathItemObject[method]; ok {
			if _type == "responseObject" {
				if responseObject, ok := operationObject.Responses["200"]; ok {
					if responseObject.Ref != "" {
						if schema, ok := openapi.getSchemaFromResponses(responseObject.Ref); ok {
							return schema, nil
						}
					}

					return getSchemaFromContent(responseObject.Content)
				}
			}
		} else {
			err = fmt.Errorf("[OpenApi.getResponseSchema] missing OperationObject %s.%s", path, method)
		}
	} else {
		err = fmt.Errorf("[OpenApi.getResponseSchema] missing PathItemObject %s", path)
	}

	return schema, err
}

func (openapi *OpenApi) getSchemaName(path string, method string) (ret string, err error) {
	if pathItemObject, ok := openapi.Paths[path]; ok {
		method = strings.ToLower(method)

		if operationObject, ok := pathItemObject[method]; ok {
			if method == "post" {
				ret = OpenApiGetSchemaName(operationObject.RequestBody.Ref)
			} else {
				if responseObject, ok := operationObject.Responses["200"]; ok {
					schema, err := openapi.getSchemaFromRef(responseObject.Ref)

					if err == nil && schema.Type == "array" && schema.Items != nil && schema.Items.Ref != "" {
						ret = OpenApiGetSchemaName(schema.Items.Ref)
					} else {
						ret = OpenApiGetSchemaName(responseObject.Ref)
					}
				}
			}
		} else {
			err = fmt.Errorf("[OpenApi.getSchemaName] missing OperationObject %s.%s", path, method)
		}
	} else {
		err = fmt.Errorf("[OpenApi.getSchemaName] missing PathItemObject %s", path)
	}

	return ret, err
}

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

		if err != nil || foreignKeyDescription == nil {
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
