package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/umuttopalak/orders-api/handler"
	"github.com/umuttopalak/orders-api/repository/customer"
	"github.com/umuttopalak/orders-api/repository/order"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/customer", a.loadCustomerRoutes)
	router.Route("/order", a.loadOrderRoutes)

	a.router = router
}

func (a *App) loadOrderRoutes(router chi.Router) {
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: a.rdb,
		},
	}

	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateByID)
	router.Delete("/{id}", orderHandler.DeleteByID)
}

func (a *App) loadCustomerRoutes(router chi.Router) {
	customerHandler := &handler.Customer{
		Repo: &customer.RedisRepo{
			Client: a.rdb,
		},
	}

	router.Post("/", customerHandler.Create)
	router.Get("/", customerHandler.List)
	router.Get("/{id}", customerHandler.GetByID)
	//router.Put("/{id}", customerHandler.UpdateByID)
	router.Delete("/{id}", customerHandler.DeleteByID)
}
