package transport

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestList(t *testing.T)  {
	w := httptest.NewRecorder()
	getOrders(w, nil)
	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", response.StatusCode, http.StatusOK)
	}

	jsoneString, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	items := make([]Order, 10)
	if err = json.Unmarshal(jsoneString, &items); err != nil {
		t.Errorf("Can't parce json response with error %v", err)
	}
}

func TestOrder(t *testing.T)  {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://http://localhost:8000/api/v1/getOrder/1234888", nil)
	getOrder(w, r)
	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", response.StatusCode, http.StatusOK)
	}

	jsoneString, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	items := OrderDetail{}
	if err = json.Unmarshal(jsoneString, &items); err != nil {
		t.Errorf("Can't parce json response with error %v", err)
	}
}