package main

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/util"
	"log"
	"os"
	"testing"
)

func TestRaceFail(t *testing.T) {
	testingDir := t.TempDir()
	f, _ := os.CreateTemp(testingDir, "policies")
	a := fileadapter.NewAdapter(f.Name())

	e, err := casbin.NewSyncedEnforcer("./auth_model.conf", a)
	if err != nil {
		log.Fatal("failed to init enforcer", err)
	}

	e.EnableAutoSave(true)
	_ = e.AddNamedMatchingFunc("g2", "g2", util.KeyMatch4)

	var gs [][]string
	gs = append(gs, []string{"admin@metrika.co", "registered_user", "metrika"})
	for i := 0; i < 1000; i++ {
		// Not strictly required, but makes it significantly more consistent
		gs = append(gs, []string{
			fmt.Sprintf("%d@metrika.co", i),
			"registered_user",
			"metrika",
		})
	}
	_, err = e.AddGroupingPolicies(gs)
	if err != nil {
		panic(err)
	}

	var g2s [][]string
	g2s = append(g2s, []string{"/example/*", "metrika_resources"})
	for i := 0; i < 1000; i++ {
		// Not strictly required, but makes it significantly more consistent
		g2s = append(g2s, []string{fmt.Sprintf("/ex%d/*", i), "metrika_resources"})
	}
	_, err = e.AddNamedGroupingPolicies("g2", g2s)
	if err != nil {
		panic(err)
	}

	_, err = e.AddPolicies([][]string{{"registered_user", "metrika", "metrika_resources", "GET", "allow"}})
	if err != nil {
		panic(err)
	}
	_ = e.SavePolicy()

	// Make some concurrent requests
	n := 100
	doneCh := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(i int) {
			var res bool
			res, err = e.Enforce("admin@metrika.co", "metrika", "/example/1", "GET")
			if err != nil {
				panic(err)
			}
			if !res {
				panic(fmt.Sprintf("result failure: %d", i))
			}

			doneCh <- struct{}{}
		}(i)
	}
	for i := 0; i < n; i++ {
		<-doneCh
	}

	println("Done!")

}
