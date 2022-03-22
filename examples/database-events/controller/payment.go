package controller

import (
	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/brianvoe/gofakeit/v6"
)

func RandomPayment() payment.Payment {
	return payment.Payment{
		Id:     gofakeit.UUID(),
		From:   gofakeit.Name(),
		To:     gofakeit.Name(),
		Amount: gofakeit.Price(1000, 10000),
	}
}
