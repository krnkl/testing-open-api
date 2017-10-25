package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/h2non/baloo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test stores the HTTP testing client preconfigured
var test = baloo.New("http://localhost:8080")

// assert implements an assertion function with custom validation logic.
// If the assertion fails it should return an error.
func assertS(res *http.Response, req *http.Request) error {

	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("%s  %s\t==> %s", req.Method, req.URL.Path, string(response))
	if res.StatusCode >= 400 {
		return errors.New("Invalid server response (> 400)")
	}
	return nil
}

func TestSystemPet(t *testing.T) {
	test.Get("/v2/user/user1").
		Expect(t).
		Status(200).
		Type("json").
		BodyMatchString("UserName").
		AssertFunc(assertS).
		Done()
}

func loadDocument(t *testing.T, file string) (*loads.Document, []byte) {
	d, err := fmts.YAMLDoc(file)
	require.NoError(t, err)
	// t.Log(string(d))
	doc, err := loads.Analyzed(json.RawMessage(d), "2.0")
	require.NoError(t, err)
	return doc, d
}

func TestLoadOpenAPIDef(t *testing.T) {

	// bytes, err := ioutil.ReadFile("api/swagger.yaml")
	// require.NoError(t, err)
	// var data yaml.MapSlice
	// err = yaml.Unmarshal(bytes, &data)
	// require.NoError(t, err)
	doc, _ := loadDocument(t, "api/swagger.yaml")
	t.Log(doc.BasePath())
	val, err := doc.Spec().Paths.JSONLookup("/user/{username}")
	assert.NoError(t, err)
	t.Logf("path: %s", val)
	ser, err := json.Marshal(doc.Spec().Definitions["User"].SchemaProps)
	require.NoError(t, err)
	t.Log("schema from definitions:", string(ser))
	// doc.Spec().Paths.Paths["/user/{username}"].Get.Responses.ResponsesProps.StatusCodeResponses[200].
	ref, err := jsonreference.New("file:///Users/mkrapyvchenk/go/src/github.com/krnkl/testing-open-api#/definitions/User")
	require.NoError(t, err)
	t.Log(ref.GetPointer())
	// val, err := doc.Spec().JSONLookup("/definitions/User")
	// TODO: How to Load schema by its ref
	userSchema := doc.Spec().Definitions["User"]
	var input map[string]interface{}

	require.NoError(t, json.Unmarshal([]byte(`{"id":123,"username1":"User1 Lastname"}`), &input))
	// input["place"] = json.Number("10")

	// TODO: How to get schema using ref or pointer

	err = validate.AgainstSchema(userSchema.AsReadOnly(), input, strfmt.Default)
	t.Logf("%+v", err)
	assert.NoError(t, err, "schema validation failed")
	schema, err := json.Marshal(doc.Spec().Paths.Paths["/user/{username}"].Get.Responses.ResponsesProps.StatusCodeResponses[200].Schema)
	require.NoError(t, err)
	// t.Logf("%+v", doc.Spec().Paths.Paths["/user/{username}"].Get.Responses.ResponsesProps.StatusCodeResponses[200].Schema)
	t.Log("Schema from paths:", string(schema))
}

func TestGetSchemaByRefAndValidate(t *testing.T) {
	doc, jsonValue := loadDocument(t, "api/swagger.yaml")
	t.Log(doc.BasePath())
	val, err := doc.Spec().Paths.JSONLookup("/user/{username}")

	assert.NoError(t, err)
	path, ok := val.(*spec.PathItem)
	assert.True(t, ok)
	response, err := path.Get.OperationProps.Responses.JSONLookup("200")

	assert.NoError(t, err)
	responseOK, ok := response.(spec.Response)
	assert.True(t, ok)

	t.Logf("%+v", responseOK.Schema)
	t.Logf("%+v", responseOK.Schema.SchemaProps.Ref.GetPointer())

	userpointer := responseOK.Schema.SchemaProps.Ref.GetPointer()

	parsedjson := make(map[string]interface{})
	err = json.Unmarshal(jsonValue, &parsedjson)
	assert.NoError(t, err)

	def, kind, err := userpointer.Get(parsedjson)
	assert.NoError(t, err)
	t.Logf("%+v", kind)

	serialized, err := json.Marshal(def)
	assert.NoError(t, err)
	// assert.True(t, ok)
	t.Logf("%s", string(serialized))

	targetSchema := &spec.Schema{}

	err = json.Unmarshal(serialized, targetSchema)
	assert.NoError(t, err)

	t.Logf("%+v", targetSchema)

	var input map[string]interface{}

	require.NoError(t, json.Unmarshal([]byte(`{"id":123,"username":"User1 Lastname"}`), &input))

	err = validate.AgainstSchema(targetSchema.AsReadOnly(), input, strfmt.Default)
	t.Logf("%+v", err)
	assert.NoError(t, err, "schema validation failed")

}
