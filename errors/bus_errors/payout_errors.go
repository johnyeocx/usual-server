package bus_errors

import (
	"net/http"

	"github.com/johnyeocx/usual/server/db/models"
)

type PayoutError string
const (
	InsertPayoutFailed PayoutError = "insert_payout_failed"
)

func InsertPayoutFailedErr(err error) *models.RequestError {
	return &models.RequestError{
		Err: err,
		StatusCode: http.StatusBadGateway,
		Code: string(InsertPayoutFailed),
	}
}
