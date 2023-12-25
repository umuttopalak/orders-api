package model

import (
	"net/mail"
)

type Customer struct {
	CustomerID uint64       `json:"customer_id"`
	Name       string       `json:"name"`
	Surname    string       `json:"surname"`
	Email      mail.Address `json:"email"`
	Is_deleted bool         `json:"is_deleted"`
}
