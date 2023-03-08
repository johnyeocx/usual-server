package bus_errors

import (
	"net/http"

	"github.com/johnyeocx/usual/server/db/models"
)

type BusError string
const (
	GetBusFailed BusError = "get_business_failed"
	GetInvoicesFailed BusError = "get_business_invoices_failed"
	GetPayoutsFailed BusError = "get_business_payouts_failed"
)

func GetBusFailedReqErr(err error) *models.RequestError {
	return &models.RequestError{
		Err: err,
		StatusCode: http.StatusBadGateway,
		Code: string(GetBusFailed),
	}
}

func GetInvoicesFailedErr(err error) *models.RequestError {
	return &models.RequestError{
		Err: err,
		StatusCode: http.StatusBadGateway,
		Code: string(GetInvoicesFailed),
	}
}

func GetPayoutsFailedErr(err error) *models.RequestError {
	return &models.RequestError{
		Err: err,
		StatusCode: http.StatusBadGateway,
		Code: string(GetPayoutsFailed),
	}
}