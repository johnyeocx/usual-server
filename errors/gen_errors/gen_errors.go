package gen_errors

import (
	"net/http"

	"github.com/johnyeocx/usual/server/db/models"
)

func NotSupportedReqErr(err error) *models.RequestError {
	return &models.RequestError{
		Err: err,
		StatusCode: http.StatusNotImplemented,
		Code: "not_supported",
	}
}

