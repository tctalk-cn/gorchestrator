package http

import (
	"encoding/json"
	"fmt"
	"github.com/tctalk-cn/gorchestrator/go/collection"
	"net/http"

	"github.com/martini-contrib/render"

	"github.com/tctalk-cn/gorchestrator/go/inst"
)

// APIResponseCode is an OK/ERROR response code
type APIResponseCode int

const (
	ERROR APIResponseCode = iota
	OK
)

var apiSynonyms = map[string]string{
	"relocate-slaves":            "relocate-replicas",
	"regroup-slaves":             "regroup-replicas",
	"move-up-slaves":             "move-up-replicas",
	"repoint-slaves":             "repoint-replicas",
	"enslave-siblings":           "take-siblings",
	"enslave-master":             "take-master",
	"regroup-slaves-bls":         "regroup-replicas-bls",
	"move-slaves-gtid":           "move-replicas-gtid",
	"regroup-slaves-gtid":        "regroup-replicas-gtid",
	"match-slaves":               "match-replicas",
	"match-up-slaves":            "match-up-replicas",
	"regroup-slaves-pgtid":       "regroup-replicas-pgtid",
	"detach-slave":               "detach-replica",
	"reattach-slave":             "reattach-replica",
	"detach-slave-master-host":   "detach-replica-master-host",
	"reattach-slave-master-host": "reattach-replica-master-host",
	"cluster-osc-slaves":         "cluster-osc-replicas",
	"start-slave":                "start-replica",
	"restart-slave":              "restart-replica",
	"stop-slave":                 "stop-replica",
	"stop-slave-nice":            "stop-replica-nice",
	"reset-slave":                "reset-replica",
	"restart-slave-statements":   "restart-replica-statements",
}

// 注册路径
var registeredPath = []string{}

// 空节点标识
var emptyInstanceKey inst.InstanceKey

func (this *APIResponseCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.String())
}

func (this *APIResponseCode) String() string {
	switch *this {
	case ERROR:
		return "ERROR"
	case OK:
		return "OK"
	}
	return "unknown"
}

func (this *APIResponseCode) HttpStatus() int {
	switch *this {
	case ERROR:
		return http.StatusInternalServerError
	case OK:
		return http.StatusOK
	}
	return http.StatusNotImplemented
}

// APIResponse is a response returned as JSON to various requests.
type APIResponse struct {
	Code    APIResponseCode
	Message string
	Details interface{}
}

func Respond(r render.Render, apiResponse *APIResponse) {
	r.JSON(apiResponse.Code.HttpStatus(), apiResponse)
}

type HttpAPI struct {
	URLPrefix string
}

var API HttpAPI = HttpAPI{}
var discoveryMetrics = collection.CreateOrReturnCollection("DISCOVERY_METRICS")
var queryMetrics = collection.CreateOrReturnCollection("BACKED_WRITES")
var writeBufferMetrics = collection.CreateOrReturnCollection("WRITE_BUFFER")

func (this *HttpAPI) getInstanceKeyInternal(host string, port string, resolve bool) (inst.InstanceKey, error) {
	var instanceKey *inst.InstanceKey
	var err error
	if resolve {
		instanceKey, err = inst.NewResolveInstanceKeyStrings(host, port)
	} else {
		instanceKey, err = inst.NewRawInstanceKeyStrings(host, port)
	}
	if err != nil {
		return emptyInstanceKey, err
	}
	instanceKey, err = inst.FigureInstanceKey(instanceKey, nil)
	if err != nil {
		return emptyInstanceKey, err
	}
	if instanceKey == nil {
		return emptyInstanceKey, fmt.Errorf("Unexpected nil instanceKey in getInstanceKeyInternal(%+v,%+v,%+v)", host, port, resolve)
	}
	return *instanceKey, nil
}
