package transpiler_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/krateoplatformops/crdgen/internal/ptr"
	"github.com/krateoplatformops/crdgen/internal/transpiler"
	"github.com/krateoplatformops/crdgen/internal/transpiler/jsonschema"
)

func TestRootRef(t *testing.T) {
	root := jsonschema.Schema{
		SchemaType: "http://json-schema.org/draft-06/schema#",
		Title:      "TestRootRef",
		Definitions: map[string]*jsonschema.Schema{
			"Address": {
				TypeValue:                "object",
				AdditionalPropertiesBool: ptr.To(false),
				Properties: map[string]*jsonschema.Schema{
					"zipcode": {
						Reference: "#/definitions/zip",
					},
				},
			},
			"zip": {TypeValue: "string"},
		},
		Reference: "#/definitions/Address",
	}
	root.Init()

	structs, err := transpiler.Transpile(&root)
	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	testField(structs["Root"].Fields["Zipcode"], "zipcode", "Zipcode", "string", false, "", t)

}

func TestFieldGeneration(t *testing.T) {
	properties := map[string]*jsonschema.Schema{
		"property1": {TypeValue: "string"},
		"property2": {Reference: "#/definitions/address"},
		// test sub-objects with properties or additionalProperties
		"property3": {
			TypeValue:   "object",
			Title:       "SubObj1",
			Description: "SubObj1 is a sub-object",
			Properties: map[string]*jsonschema.Schema{
				"name": {TypeValue: "string"},
			},
		},
		"property4": {
			TypeValue:            "object",
			Title:                "SubObj2",
			AdditionalProperties: &jsonschema.AdditionalProperties{TypeValue: "integer"},
		},
		// test sub-objects with properties composed of objects
		"property5": {
			TypeValue: "object",
			Title:     "SubObj3",
			Properties: map[string]*jsonschema.Schema{
				"SubObj3a": {
					TypeValue: "object",
					Properties: map[string]*jsonschema.Schema{
						"subproperty1": {TypeValue: "integer"},
					},
				},
			},
		},
		// test sub-objects with additionalProperties composed of objects
		"property6": {
			TypeValue: "object",
			Title:     "SubObj4",
			AdditionalProperties: &jsonschema.AdditionalProperties{
				TypeValue: "object",
				Title:     "SubObj4a",
				Properties: map[string]*jsonschema.Schema{
					"subproperty1": {TypeValue: "integer"},
				},
			},
		},
		// test sub-objects without titles
		"property7": {TypeValue: "object"},
		// test sub-objects with properties AND additionalProperties
		"property8": {
			TypeValue: "object",
			Title:     "SubObj5",
			Properties: map[string]*jsonschema.Schema{
				"name": {TypeValue: "string"},
			},
			AdditionalProperties: &jsonschema.AdditionalProperties{TypeValue: "integer"},
		},
		"property9": {
			TypeValue:                "object",
			Title:                    "SubObj6",
			AdditionalPropertiesBool: ptr.To(true),
		},
		"enumProperty": {
			TypeValue: "string",
			Enum:      []interface{}{"value1", "value2", "1", "2"},
		},
	}

	requiredFields := []string{"property2"}

	root := jsonschema.Schema{
		SchemaType: "http://localhost",
		Title:      "TestFieldGeneration",
		TypeValue:  "object",
		Properties: properties,
		Definitions: map[string]*jsonschema.Schema{
			"address": {TypeValue: "object"},
		},
		Required: requiredFields,
	}
	root.Init()

	structs, err := transpiler.Transpile(&root)
	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	b, _ := json.MarshalIndent(structs, "", "  ")
	fmt.Println(string(b))

	if len(structs) != 9 {
		t.Errorf("Expected 9 results, but got %d results", len(structs))
	}

	// b, _ := json.MarshalIndent(structs, "", "  ")
	// fmt.Println(string(b))

	testField(structs["Root"].Fields["Property1"], "property1", "Property1", "string", false, "", t)
	testField(structs["Root"].Fields["Property2"], "property2", "Property2", "*Address", true, "", t)
	testField(structs["Root"].Fields["Property3"], "property3", "Property3", "*Property3", false, "SubObj1", t)
	testField(structs["Root"].Fields["Property4"], "property4", "Property4", "map[string]int", false, "SubObj2", t)
	testField(structs["Root"].Fields["Property5"], "property5", "Property5", "*Property5", false, "SubObj3", t)
	testField(structs["Root"].Fields["Property6"], "property6", "Property6", "map[string]*Property6Item", false, "SubObj4", t)
	testField(structs["Root"].Fields["Property7"], "property7", "Property7", "*Property7", false, "", t)
	testField(structs["Root"].Fields["Property8"], "property8", "Property8", "*Property8", false, "SubObj5", t)
	testField(structs["Root"].Fields["Property9"], "property9", "Property9", "*Property9", false, "SubObj6", t)
	testField(structs["Root"].Fields["EnumProperty"], "enumProperty", "EnumProperty", "string", false, "", t)

	testField(structs["Property3"].Fields["Name"], "name", "Name", "string", false, "", t)
	testField(structs["Property5"].Fields["SubObj3a"], "SubObj3a", "SubObj3a", "*SubObj3a", false, "", t)
	testField(structs["Property6Item"].Fields["Subproperty1"], "subproperty1", "Subproperty1", "int", false, "", t)

	testField(structs["Property8"].Fields["Name"], "name", "Name", "string", false, "", t)
	//testField(structs["Property8"].Fields["AdditionalProperties"], "-", "AdditionalProperties", "map[string]int", false, t)

	if strct, ok := structs["Property7"]; !ok {
		t.Fatal("Property7 wasn't generated")
	} else {
		if len(strct.Fields) != 0 {
			t.Fatal("Property7 expected 0 fields")
		}
	}

	en := structs["Root"].Fields["EnumProperty"].Enum

	check := []string{"value1", "value2", "1", "2"}
	for _, e := range en {
		if !slices.ContainsFunc(check, func(i string) bool {
			return strconv.Quote(i) == e
		}) {
			t.Errorf("Expected enum value %s", e)
		}

	}

	// test the enum property
}

func RootWithArrayReferences(t *testing.T) {
	properties := map[string]*jsonschema.Schema{
		"property1": {TypeValue: "string"},
		"property2": {
			TypeValue: "array",
			Items: &jsonschema.Schema{
				Reference: "#/definitions/address",
			},
		},
		"property3": {
			TypeValue: "array",
			Items: &jsonschema.Schema{
				TypeValue:            "object",
				AdditionalProperties: (*jsonschema.AdditionalProperties)(&jsonschema.Schema{TypeValue: "integer"}),
			},
		},
		"property4": {
			TypeValue: "array",
			Items: &jsonschema.Schema{
				Reference: "#/definitions/outer",
			},
		},
	}

	requiredFields := []string{"property2"}

	root := jsonschema.Schema{
		SchemaType: "http://localhost",
		Title:      "RootWithArrayReferences",
		TypeValue:  "object",
		Properties: properties,
		Definitions: map[string]*jsonschema.Schema{
			"address": {TypeValue: "object"},
			"outer":   {TypeValue: "array", Items: &jsonschema.Schema{Reference: "#/definitions/inner"}},
			"inner":   {TypeValue: "object"},
		},
		Required: requiredFields,
	}
	root.Init()

	structs, err := transpiler.Transpile(&root)
	if err != nil {
		t.Error("Failed to get the fields: ", err)
	}

	if len(structs) != 3 {
		t.Errorf("Expected 3 results, but got %d results", len(structs))
	}

	testField(structs["Root"].Fields["Property1"], "property1", "Property1", "string", false, "", t)
	testField(structs["Root"].Fields["Property2"], "property2", "Property2", "[]*Address", true, "", t)
	testField(structs["Root"].Fields["Property3"], "property3", "Property3", "[]map[string]int", false, "", t)
	testField(structs["Root"].Fields["Property4"], "property4", "Property4", "[][]*Inner", false, "", t)
}

func testField(actual transpiler.Field, expectedJSONName string, expectedName string, expectedType string, expectedToBeRequired bool, expectedTitle string, t *testing.T) {
	if actual.JSONName != expectedJSONName {
		t.Errorf("JSONName - expected \"%s\", got \"%s\"", expectedJSONName, actual.JSONName)
	}

	if actual.Title != expectedTitle {
		t.Errorf("Title - expected \"%s\", got \"%s\"", expectedTitle, actual.Title)
	}

	if actual.Name != expectedName {
		t.Errorf("Name - expected \"%s\", got \"%s\"", expectedName, actual.Name)
	}
	if actual.Type != expectedType {
		t.Errorf("Type - expected \"%s\", got \"%s\"", expectedType, actual.Type)
	}
	if actual.Required != expectedToBeRequired {
		t.Errorf("Required - expected \"%v\", got \"%v\"", expectedToBeRequired, actual.Required)
	}
}

func TestNestedStructGeneration(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Example"
	root.Properties = map[string]*jsonschema.Schema{
		"property1": {
			TypeValue: "object",
			Properties: map[string]*jsonschema.Schema{
				"subproperty1": {TypeValue: "string"},
			},
		},
	}

	root.Init()

	results, err := transpiler.Transpile(root)
	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Errorf("2 results should have been created, a root type and a type for the object 'property1' but %d structs were made", len(results))
	}

	if _, contains := results["Root"]; !contains {
		t.Errorf("The Example type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if _, contains := results["Property1"]; !contains {
		t.Errorf("The Property1 type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if results["Root"].Fields["Property1"].Type != "*Property1" {
		t.Errorf("Expected that the nested type property1 is generated as a struct, so the property type should be *Property1, but was %s.", results["Root"].Fields["Property1"].Type)
	}
}

func TestEmptyNestedStructGeneration(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Example"
	root.Properties = map[string]*jsonschema.Schema{
		"property1": {
			TypeValue: "object",
			Properties: map[string]*jsonschema.Schema{
				"nestedproperty1": {TypeValue: "string"},
			},
		},
	}

	root.Init()

	results, err := transpiler.Transpile(root)
	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Errorf("2 results should have been created, a root type and a type for the object 'property1' but %d structs were made", len(results))
	}

	if _, contains := results["Root"]; !contains {
		t.Errorf("The Example type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if _, contains := results["Property1"]; !contains {
		t.Errorf("The Property1 type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if results["Root"].Fields["Property1"].Type != "*Property1" {
		t.Errorf("Expected that the nested type property1 is generated as a struct, so the property type should be *Property1, but was %s.", results["Root"].Fields["Property1"].Type)
	}
}

func TestStructNameExtractor(t *testing.T) {
	m := make(map[string]transpiler.Struct)
	m["name1"] = transpiler.Struct{}
	m["name2"] = transpiler.Struct{}

	names := getStructNamesFromMap(m)
	if len(names) != 2 {
		t.Error("Didn't extract all names from the map.")
	}

	if !contains(names, "name1") {
		t.Error("name1 was not extracted")
	}

	if !contains(names, "name2") {
		t.Error("name2 was not extracted")
	}
}

func getStructNamesFromMap(m map[string]transpiler.Struct) []string {
	sn := make([]string, len(m))
	i := 0
	for k := range m {
		sn[i] = k
		i++
	}
	return sn
}

func TestStructGeneration(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "RootElement"
	root.Definitions = make(map[string]*jsonschema.Schema)
	root.Definitions["address"] = &jsonschema.Schema{
		Properties: map[string]*jsonschema.Schema{
			"address1": {TypeValue: "string"},
			"zip":      {TypeValue: "number"},
		},
	}
	root.Properties = map[string]*jsonschema.Schema{
		"property1": {TypeValue: "string"},
		"property2": {Reference: "#/definitions/address"},
	}

	root.Init()

	results, err := transpiler.Transpile(root)
	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Error("2 results should have been created, a root type and an address")
	}
}

func TestArrayGeneration(t *testing.T) {
	root := &jsonschema.Schema{
		Title:     "Array of Artists Example",
		TypeValue: "array",
		Items: &jsonschema.Schema{
			Title:     "Artist",
			TypeValue: "object",
			Properties: map[string]*jsonschema.Schema{
				"name":      {TypeValue: "string"},
				"birthyear": {TypeValue: "number"},
			},
		},
	}

	root.Init()

	results, err := transpiler.Transpile(root)
	if err != nil {
		t.Fatal("Failed to create structs: ", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected one struct should have been generated, but %d have been generated.", len(results))
	}

	artistStruct, ok := results["RootItems"]
	if !ok {
		t.Errorf("Expected Name to be Artist, that wasn't found, but the struct contains \"%+v\"", results)
	}

	if len(artistStruct.Fields) != 2 {
		t.Errorf("Expected the fields to be birtyear and name, but %d fields were found.", len(artistStruct.Fields))
	}

	if _, ok := artistStruct.Fields["Name"]; !ok {
		t.Errorf("Expected to find a Name field, but one was not found.")
	}

	if _, ok := artistStruct.Fields["Birthyear"]; !ok {
		t.Errorf("Expected to find a Birthyear field, but one was not found.")
	}
}

func TestNestedArrayGeneration(t *testing.T) {
	root := &jsonschema.Schema{
		Title:     "Favourite Bars",
		TypeValue: "object",
		Properties: map[string]*jsonschema.Schema{
			"barName": {TypeValue: "string"},
			"cities": {
				TypeValue: "array",
				Items: &jsonschema.Schema{
					Title:     "City",
					TypeValue: "object",
					Properties: map[string]*jsonschema.Schema{
						"name":    {TypeValue: "string"},
						"country": {TypeValue: "string"},
					},
				},
			},
			"tags": {
				TypeValue: "array",
				Items:     &jsonschema.Schema{TypeValue: "string"},
			},
		},
	}

	root.Init()

	results, err := transpiler.Transpile(root)
	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected two structs to be generated - 'Favourite Bars' and 'City', but %d have been generated.", len(results))
	}

	fbStruct, ok := results["Root"]
	if !ok {
		t.Errorf("FavouriteBars struct was not found. The results were %+v", results)
	}

	if _, ok := fbStruct.Fields["BarName"]; !ok {
		t.Errorf("Expected to find the BarName field, but didn't. The struct is %+v", fbStruct)
	}

	f, ok := fbStruct.Fields["Cities"]
	if !ok {
		t.Errorf("Expected to find the Cities field on the FavouriteBars, but didn't. The struct is %+v", fbStruct)
	}
	if f.Type != "[]*CitiesItems" {
		t.Errorf("Expected to find that the Cities array was of type *City, but it was of %s", f.Type)
	}

	f, ok = fbStruct.Fields["Tags"]
	if !ok {
		t.Errorf("Expected to find the Tags field on the FavouriteBars, but didn't. The struct is %+v", fbStruct)
	}

	if f.Type != "[]string" {
		t.Errorf("Expected to find that the Tags array was of type string, but it was of %s", f.Type)
	}

	cityStruct, ok := results["CitiesItems"]
	if !ok {
		t.Error("City struct was not found.")
	}

	if _, ok := cityStruct.Fields["Name"]; !ok {
		t.Errorf("Expected to find the Name field on the City struct, but didn't. The struct is %+v", cityStruct)
	}

	if _, ok := cityStruct.Fields["Country"]; !ok {
		t.Errorf("Expected to find the Country field on the City struct, but didn't. The struct is %+v", cityStruct)
	}
}

func TestMultipleSchemaStructGeneration(t *testing.T) {
	root1 := &jsonschema.Schema{
		Title: "Root1Element",
		ID:    "http://example.com/schema/root1",
		Properties: map[string]*jsonschema.Schema{
			"property1": {Reference: "root2#/definitions/address"},
		},
	}

	root2 := &jsonschema.Schema{
		Title: "Root2Element",
		ID:    "http://example.com/schema/root2",
		Properties: map[string]*jsonschema.Schema{
			"property1": {Reference: "#/definitions/address"},
		},
		Definitions: map[string]*jsonschema.Schema{
			"address": {
				Properties: map[string]*jsonschema.Schema{
					"address1": {TypeValue: "string"},
					"zip":      {TypeValue: "number"},
				},
			},
			"person": {
				Properties: map[string]*jsonschema.Schema{
					"name": {TypeValue: "string"},
					"age":  {TypeValue: "number"},
				},
			},
		},
	}

	root1.Init()
	root2.Init()

	results, err := transpiler.Transpile(root1, root2)
	if err != nil {
		t.Error("Failed to create structs: ", err)
	}

	if len(results) != 3 {
		t.Errorf("3 results should have been created, 2 root types and an address, but got %v", getStructNamesFromMap(results))
	}
}

func TestThatArraysWithoutDefinedItemTypesAreGeneratedAsEmptyInterfaces(t *testing.T) {
	root := &jsonschema.Schema{}
	root.Title = "Array without defined item"
	root.Properties = map[string]*jsonschema.Schema{
		"name": {TypeValue: "string"},
		"repositories": {
			TypeValue: "array",
		},
	}

	root.Init()

	results, err := transpiler.Transpile(root)
	if err != nil {
		t.Errorf("Error generating structs: %v", err)
	}

	if _, contains := results["Root"]; !contains {
		t.Errorf("The ArrayWithoutDefinedItem type should have been made, but only types %s were made.", strings.Join(getStructNamesFromMap(results), ", "))
	}

	if o, ok := results["ArrayWithoutDefinedItem"]; ok {
		if f, ok := o.Fields["Repositories"]; ok {
			if f.Type != "[]any" {
				t.Errorf("Since the schema doesn't include a type for the array items, the property type should be []any, but was %s.", f.Type)
			}
		} else {
			t.Errorf("Expected the ArrayWithoutDefinedItem type to have a Repostitories field, but none was found.")
		}
	}
}

// Root is an example of a generated type.
type Root struct {
	Name any `json:"name,omitempty"`
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
