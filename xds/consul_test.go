package xds

import (
	"reflect"
	"testing"
)

var (
	services = map[string][]string{
		"consul": []string{},
		"probe":  []string{},
	}
)

func TestPopulateCache(t *testing.T) {
	cache := NewConsulCache()

	expected := ConsulCache{
		Services: map[string]*Service{
			"consul": &Service{Name: "consul"},
			"probe":  &Service{Name: "probe"},
		},
	}
	cache.populateCache(services)
	eq := reflect.DeepEqual(expected.Services, cache.Services)

	if eq == false {
		t.Errorf("populateCache producing the wrong result with Cache's: %+v and %+v", expected.Services, cache.Services)
	}

}
