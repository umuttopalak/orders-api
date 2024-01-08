package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/umuttopalak/orders-api/handler"
	"github.com/umuttopalak/orders-api/repository/category"
	"github.com/umuttopalak/orders-api/repository/customer"
	"github.com/umuttopalak/orders-api/repository/order"
	"github.com/umuttopalak/orders-api/repository/product"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/customer", a.loadCustomerRoutes)
	router.Route("/order", a.loadOrderRoutes)
	router.Route("/product", a.loadProductRoutes)
	router.Route("/category", a.loadCategoryRoutes)
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

func (a *App) loadProductRoutes(router chi.Router) {
	productHandler := &handler.Product{
		Repo: &product.RedisRepo{
			Client: a.rdb,
		},
	}
	router.Post("/", productHandler.Create)
	router.Get("/", productHandler.List)
	router.Get("/{id}", productHandler.GetByID)
	router.Put("/{id}", productHandler.UpdateByID)
	router.Delete("/{id}", productHandler.DeleteByID)

}

func (a *App) loadCategoryRoutes(router chi.Router) {
	categoryHandler := &handler.Category{
		Repo: &category.RedisRepo{
			Client: a.rdb,
		},
	}
	router.Post("/", categoryHandler.Create)
	router.Get("/", categoryHandler.List)
	router.Get("/{id}", categoryHandler.GetByID)
	router.Put("/{id}", categoryHandler.UpdateByID)
	router.Delete("/{id}", categoryHandler.DeleteByID)

}
