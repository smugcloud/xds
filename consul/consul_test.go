package consul

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
	cache := NewCache()

	expected := Cache{
		Services: map[string]*Service{
			"consul": &Service{Name: "consul"},
			"probe":  &Service{Name: "probe"},
		},
	}
	result := cache.populateCache(services)
	eq := reflect.DeepEqual(expected.Services, result.Services)

	if eq == false {
		t.Errorf("populateCache producing the wrong result with Cache's: %+v and %+v", expected.Services, result.Services)
	}

}
