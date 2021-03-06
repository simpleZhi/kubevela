package task

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"cuelang.org/go/cue"
	cueJson "cuelang.org/go/pkg/encoding/json"
	"github.com/bmizerany/assert"
)

const TaskTemplate = `
parameter: {
  serviceURL: string
}

processing: {
  output: {
    token ?: string
  }
  http: {
    method: *"GET" | string
    url: parameter.serviceURL
    request: {
        body ?: bytes
        header: {}
        trailer: {}
    }
  }
}

patch: {
  data: token: processing.output.token
}

output: {
  data: processing.output.token
}
`

func TestProcess(t *testing.T) {
	s := NewMock()
	defer s.Close()

	r := cue.Runtime{}
	taskTemplate, err := r.Compile("", TaskTemplate)
	if err != nil {
		t.Fatal(err)
	}
	taskTemplate, _ = taskTemplate.Fill(map[string]interface{}{
		"serviceURL": "http://127.0.0.1:8090/api/v1/token?val=test-token",
	}, "parameter")

	inst, err := Process(taskTemplate)
	if err != nil {
		t.Fatal(err)
	}
	output := inst.Lookup("output")
	data, _ := cueJson.Marshal(output)
	assert.Equal(t, "{\"data\":\"test-token\"}", data)
}

func NewMock() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			fmt.Printf("Expected 'GET' request, got '%s'", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v1/token" {
			fmt.Printf("Expected request to '/person', got '%s'", r.URL.EscapedPath())
		}
		r.ParseForm()
		token := r.Form.Get("val")
		tokenBytes, _ := json.Marshal(map[string]interface{}{"token": token})

		w.WriteHeader(http.StatusOK)
		w.Write(tokenBytes)
	}))
	l, _ := net.Listen("tcp", "127.0.0.1:8090")
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return ts
}
