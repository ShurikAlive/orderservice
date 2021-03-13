package transport

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type OrderItem struct {
	Id string `json:"id"`
	Quantity int `json:"quantity"`
}

type OrderDetail struct {
	Id string `json:"id"`
	MenuItems []OrderItem `json:"menuItems"`
	OrderedAtTimestamp string `json:"orderedAtTimestamp"`
	Cost float32 `json:"cost"`
}

type Order struct {
	Id        string      `json:"id"`
	MenuItems []OrderItem `json:"menuItems"`
}

type DeleteOrder struct {
	Id        string      `json:"id"`
}

func Router() *mux.Router {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()
	s.HandleFunc("/hello-world", helloWorld).Methods(http.MethodGet)
	s.HandleFunc("/getOrders", getOrders).Methods(http.MethodGet)
	s.HandleFunc("/getOrder/{ID}", getOrder).Methods(http.MethodGet)

	s.HandleFunc("/order", createOrder).Methods(http.MethodPost)
	s.HandleFunc("/deleteOrder", deleteOrder).Methods(http.MethodPost)
	return r
}

func deleteOrder(w http.ResponseWriter, r *http.Request ) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var msg DeleteOrder

	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderId := msg.Id

	tx, err := GetDBInstance().Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("DELETE FROM cafe_test.orders WHERE hesh_id_order = ?;", orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	_, err = tx.Exec("DELETE FROM cafe_test.orders_items WHERE id_order = ?;", id)  // Не хочет работать =( Скорее всего id не видит =(
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	orderIdJSON := []byte(`{"id":` + orderId + `}`)

	JSONResponse(w, orderIdJSON)
}

func createOrder(w http.ResponseWriter, r *http.Request ) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var msg Order
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k := range msg.MenuItems {
		if (0 >= msg.MenuItems[k].Quantity) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	u, err := uuid.NewV4()
	orderId := u.String()

	tx, err := GetDBInstance().Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("INSERT INTO cafe_test.orders (hesh_id_order) VALUES (?);", orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	for k := range msg.MenuItems {

		_, err := tx.Exec("INSERT INTO cafe_test.orders_items (id_order, id_menu_item, quantity) SELECT DISTINCT ?, id_menu_item,? FROM cafe_test.menu_items WHERE hash_id_menu_item = ? ;", id, msg.MenuItems[k].Quantity, msg.MenuItems[k].Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}
	orderIdJSON := []byte(`{"id":` + orderId + `}`)

	JSONResponse(w, orderIdJSON)
}

func helloWorld(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "Hello World! ")
}

func getOrder(w http.ResponseWriter, r *http.Request ) {
	vars := mux.Vars(r)
	id := vars["ID"] // Для запросов вида "order/{ID}"
	// some := r.URL.Query().Get("some") // ДЛя обычных Get параметров

	order := OrderDetail{id, []OrderItem {{"1213234", 2},{"435756",6}, {"876868", 45}}, time.Now().String(),  33.99}
	b, err := json.Marshal(order)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, b)
}

func getOrders(w http.ResponseWriter, _ *http.Request ) {

	orders :=  []Order {
		{"1", []OrderItem {{"1213234", 2},{"435756",6}, {"876868", 45}}},
		{"2", []OrderItem {{"902384245",4}}},
		{"3", []OrderItem {{"456730992", 31}, {"435756", 2}}},
	}
	b, err := json.Marshal(orders)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, b)
}

func JSONResponse(w http.ResponseWriter, json []byte) {
	w.Header().Set("Content-Type","application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, string(json))
	if err != nil {
		log.WithField("err", err).Error("write response error")
	}
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"method":     r.Method,
			"url":        r.URL,
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.UserAgent(),
		}).Info("got a new request")

		startTime := time.Now()
		h.ServeHTTP(w, r)
		endTime := time.Now()

		log.WithFields(log.Fields{
			"workTimeReqest": endTime.Sub(startTime),
		}).Info("end work request")
	})
}