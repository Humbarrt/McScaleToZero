package queries_test

import (
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/model"
	"path"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestLocalInvoiceService(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	/* test these functions:
	 	GetLocalInvoiceServiceConfig
		GetLocalInvoicePlayers
		GetLocalInvoicePlayerById
		CreateLocalInvoicePlayer
		GetLocalInvoiceServicePayments
		GetLocalInvoiceServicePaymentById
		GetLocalInvoiceServicePaymentsByPlayerId
		CreateLocalInvoiceServicePayment
		UpdateLocalInvoiceServicePayment
		GetLocalInvoiceServiceInvoices
		GetLocalInvoiceServiceInvoiceById
		CreateLocalInvoiceServiceInvoice
		GetLocalInvoiceServiceInvoiceItemsByInvoiceId
		CreateLocalInvoiceServiceInvoiceItem
	*/

	/** TEST Invoice Service Config **/
	config, err := queries.GetLocalInvoiceServiceConfig(DB)
	assert.NoError(t, err)
	assert.Equal(t, "USD", config.InvoiceCurrency)
	assert.Nil(t, config.PayingPlayerId)

	payingPlayer := int64(1)
	err = queries.UpdateLocalInvoiceServiceConfig(DB, &model.LocalInvoiceServiceConfig{
		InvoiceCurrency: "EUR",
		PayingPlayerId:  &payingPlayer,
	})
	assert.NoError(t, err)

	config, err = queries.GetLocalInvoiceServiceConfig(DB)
	assert.NoError(t, err)
	assert.Equal(t, "EUR", config.InvoiceCurrency)
	assert.Equal(t, &payingPlayer, config.PayingPlayerId)

	uuid1, err := model.NewUUIDFromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *uuid1, true, false, nil, []string{"Test Player"}, nil)
	assert.NoError(t, err)

	uuid2, err := model.NewUUIDFromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2})
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *uuid2, true, false, nil, []string{"Test Player2"}, nil)
	assert.NoError(t, err)

	/** TEST Invoice Service Invoices **/
	invoices, err := queries.GetLocalInvoiceServiceInvoices(DB)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(invoices))

	// Add an invoice
	amount := decimal.NewFromFloat(100.50)
	player1Amount := decimal.NewFromFloat(30.25)
	player2Amount := decimal.NewFromFloat(70.25)
	invoiceItems := []*model.LocalInvoiceServiceInvoiceItem{
		{
			PlayerId: 1,
			Amount:   player1Amount,
		},
		{
			PlayerId: 2,
			Amount:   player2Amount,
		},
	}
	invoiceId, err := queries.AddLocalInvoiceServiceInvoice(DB, amount, "Test Invoice", invoiceItems)
	assert.NoError(t, err)

	invoices, err = queries.GetLocalInvoiceServiceInvoices(DB)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(invoices))
	assert.Equal(t, invoiceId, invoices[0].Id)
	assert.True(t, amount.Equal(invoices[0].Amount))
	assert.Equal(t, "Test Invoice", invoices[0].Description)

	invoice, err := queries.GetLocalInvoiceServiceInvoiceById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), invoice.Id)
	assert.True(t, amount.Equal(invoice.Amount))
	assert.Equal(t, "Test Invoice", invoice.Description)

	invoiceItemsFetched, err := queries.GetLocalInvoiceServiceInvoiceItemsByInvoiceId(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(invoiceItemsFetched))
	assert.Equal(t, int64(1), invoiceItemsFetched[0].PlayerId)
	assert.True(t, player1Amount.Equal(invoiceItemsFetched[0].Amount))
	assert.Equal(t, int64(2), invoiceItemsFetched[1].PlayerId)
	assert.True(t, player2Amount.Equal(invoiceItemsFetched[1].Amount))

	/** TEST Payments **/
	payments, err := queries.GetLocalInvoiceServicePayments(DB)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(payments))

	// Add a payment
	paymentAmount := decimal.NewFromFloat(50.75)
	err = queries.AddLocalInvoiceServicePayment(DB, 1, paymentAmount, "Test Payment")
	assert.NoError(t, err)

	payments, err = queries.GetLocalInvoiceServicePayments(DB)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(payments))
	assert.Equal(t, int64(1), payments[0].Id)
	assert.Equal(t, int64(1), payments[0].PlayerId)
	assert.True(t, paymentAmount.Equal(payments[0].Amount))
	assert.Equal(t, "Test Payment", payments[0].Description)

	payment, err := queries.GetLocalInvoiceServicePaymentById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), payment.Id)
	assert.Equal(t, int64(1), payment.PlayerId)
	assert.True(t, paymentAmount.Equal(payment.Amount))
	assert.Equal(t, "Test Payment", payment.Description)

	paymentsByPlayer, err := queries.GetLocalInvoiceServicePaymentsByPlayerId(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paymentsByPlayer))
	assert.Equal(t, int64(1), paymentsByPlayer[0].Id)
	assert.Equal(t, int64(1), paymentsByPlayer[0].PlayerId)
	assert.True(t, paymentAmount.Equal(paymentsByPlayer[0].Amount))
	assert.Equal(t, "Test Payment", paymentsByPlayer[0].Description)

	// Update the payment
	updatedPaymentAmount := decimal.NewFromFloat(75.25)
	err = queries.UpdateLocalInvoiceServicePayment(DB, &model.LocalInvoiceServicePayment{
		Id:          1,
		PlayerId:    1,
		Amount:      updatedPaymentAmount,
		Description: "Updated Test Payment",
	})
	assert.NoError(t, err)

	updatedPayment, err := queries.GetLocalInvoiceServicePaymentById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), updatedPayment.Id)
	assert.Equal(t, int64(1), updatedPayment.PlayerId)
	assert.True(t, updatedPaymentAmount.Equal(updatedPayment.Amount))
	assert.Equal(t, "Updated Test Payment", updatedPayment.Description)
}
