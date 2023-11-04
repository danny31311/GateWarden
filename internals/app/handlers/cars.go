package handlers

import (
	"GateWarden/internals/app/models"
	"GateWarden/internals/app/processors"
	"encoding/json"
	"errors"

	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type CarsHandler struct {
	processor *processors.CarsProcessor
}

func NewCarsHandler(processor *processors.CarsProcessor) *CarsHandler {
	handler := new(CarsHandler)
	handler.processor = processor
	return handler
}

func (handler *CarsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var newCar models.Car
	err := json.NewDecoder(r.Body).Decode(&newCar)
	if err != nil {
		WrapError(w, err)
		return
	}
	err = handler.processor.CreateCar(newCar)
	if err != nil {
		WrapError(w, err)
		return
	}
	var m = map[string]interface{}{
		"result": "OK",
		"data":   "",
	}
	WrapOk(w, m)
}

func (handler *CarsHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	var userIdFilter int64 = 0
	if vars.Get("userid") != "" {
		var err error
		userIdFilter, err = strconv.ParseInt(vars.Get("userid"), 10, 64)
		if err != nil {
			WrapError(w, err)
			return
		}
	}
	list, err := handler.processor.ListCars(userIdFilter, strings.Trim(vars.Get("brand"), "\""), strings.Trim(vars.Get("colour"), "\""), strings.Trim(vars.Get("license_plate"), "\""))
	if err != nil {
		WrapError(w, err)
	}
	var m = map[string]interface{}{
		"result": "OK",
		"data":   list,
	}
	WrapOk(w, m)
}

func (handler *CarsHandler) Find(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		WrapError(w, errors.New("missing id"))
		return
	}

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WrapError(w, err)
		return
	}
	car, err := handler.processor.FindCar(id)

	if err != nil {
		WrapError(w, err)
		return
	}
	var m = map[string]interface{}{
		"result": "OK",
		"data":   car,
	}
	WrapOk(w, m)
}
