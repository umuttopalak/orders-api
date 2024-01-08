package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/umuttopalak/orders-api/model"
	"github.com/umuttopalak/orders-api/repository/product"
)

type Product struct {
	Repo *product.RedisRepo
}

func (h *Product) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Category     model.Category `json:"category"`
		ProductPrice int64          `json:"price"`
		ProductName  string         `json:"product_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Product := model.Product{
		ProductID:    rand.Uint64(),
		ProductName:  body.ProductName,
		ProductPrice: body.ProductPrice,
		Category:     body.Category,
	}

	err := h.Repo.Insert(r.Context(), Product)
	if err != nil {
		fmt.Println("failed to insert: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(Product)
	if err != nil {
		fmt.Println("failed to insert: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(res); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Product) List(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.Repo.FindAll(r.Context(), product.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to find all", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Products []model.Product `json:"products"`
		Next     uint64          `json:"next,omitempty"`
	}

	response.Products = res.Products
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

func (h *Product) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, err := h.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, product.ErrNotExist) {
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

func (h *Product) UpdateByID(w http.ResponseWriter, r *http.Request) {

	var body struct {
		ProductID    uint64         `json:"product_id"`
		Category     model.Category `json:"category"`
		ProductPrice int64          `json:"price"`
		ProductName  string         `json:"product_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	ProductID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theProduct, err := h.Repo.FindByID(r.Context(), ProductID)
	if errors.Is(err, product.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find Product: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	theProduct.Category = body.Category
	theProduct.ProductPrice = body.ProductPrice
	theProduct.ProductName = body.ProductName

	err = h.Repo.Update(r.Context(), theProduct)
	if err != nil {
		fmt.Println("failed to insert: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theProduct); err != nil {
		fmt.Println("failed to marshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Product) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	productID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteByID(r.Context(), productID)
	if errors.Is(err, product.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
