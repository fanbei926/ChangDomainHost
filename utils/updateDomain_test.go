package utils

import (
	"sync"
	"testing"
)

func TestCreateClient(t *testing.T) {
	_, err := CreateClient()
	if err != nil {
		t.Error(err)
	}
}

func TestExecCurl(t *testing.T) {
	err := ExecCurl()
	if err != nil {
		t.Error(err)
	}
}

func TestExecLookup(t *testing.T) {
	err := ExecLookup()
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateDomainRecord(t *testing.T) {
	var m sync.Map
	m.Store("curlRes", "127.0.0.1")
	m.Store("dnsRes", "127.0.0.1")
	err := UpdateDomainRecord()
	if err != nil {
		t.Error(err)
	}
}
