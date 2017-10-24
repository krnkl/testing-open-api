package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/h2non/baloo"
)

// test stores the HTTP testing client preconfigured
var test = baloo.New("http://localhost:8080")

// assert implements an assertion function with custom validation logic.
// If the assertion fails it should return an error.
func assert(res *http.Response, req *http.Request) error {

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
		AssertFunc(assert).
		Done()
}

func TestLoadOpenAPIDef(t *testing.T) {

}
