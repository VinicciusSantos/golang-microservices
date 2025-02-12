package main

import (
	"os"
	"testing"

	"github.com/VinicciusSantos/golang-microservices/authentication/data"
)

var testApp Config

func TestMain(m *testing.M) {
	testApp.Repo = data.NewPostgresTestRepository(nil)
	os.Exit(m.Run())
}
