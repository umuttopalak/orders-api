package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/umuttopalak/orders-api/model"
	"github.com/umuttopalak/orders-api/repository/customer"
	"github.com/umuttopalak/orders-api/repository/order"
)

type Customer struct {
	Repo *customer.RedisRepo
}

func (c *Customer) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string       `json:"name"`
		Surname    string       `json:"surname"`
		Email      mail.Address `json:"email"`
		Is_deleted bool         `json:"is_deleted"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	customer := model.Customer{
		Name:       body.Name,
		Surname:    body.Surname,
		Email:      body.Email,
		Is_deleted: false,
	}

	err := c.Repo.Insert(r.Context(), customer)
	if err != nil {
		fmt.Println("failed to insert: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(customer)
	if err != nil {
		fmt.Println("failed to encode customer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *Customer) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := c.Repo.FindAll(r.Context(), customer.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to find all", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Customers []model.Customer `json:"customers"`
		Next      uint64           `json:"next,omitempty"`
	}

	response.Customers = res.Customers
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Customer) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	customerID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, err := c.Repo.FindByID(r.Context(), customerID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("failed to marshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Customer) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	customerID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.Repo.DeleteByID(r.Context(), customerID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
