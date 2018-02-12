package cmd_test

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	"k8s.io/api/autoscaling/v1"

	"github.com/rebuy-de/kubernetes-hpaa/cmd"
	"github.com/rebuy-de/rebuy-golang-sdk/testutil"
)

func testReadHPA(tb testing.TB, name string) *v1.HorizontalPodAutoscaler {
	tb.Helper()

	raw, err := ioutil.ReadFile(path.Join("test-fixtures", name+".json"))
	if err != nil {
		tb.Fatal(err)
	}

	hpa := new(v1.HorizontalPodAutoscaler)

	err = json.Unmarshal(raw, hpa)
	if err != nil {
		tb.Fatal(err)
	}

	return hpa
}

func TestHandle(t *testing.T) {
	hpa := testReadHPA(t, "handler-1")
	hpa, err := cmd.Handle(hpa)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGoldenJSON(t, "test-fixtures/handler-1-golden.json", hpa)
}
