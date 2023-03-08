package models

import (
	"log"

	"github.com/gin-gonic/gin"
)

type RequestError struct {
	Err 		error
	StatusCode 	int
	Code 		string
}

func (e *RequestError) ErrToMap() gin.H{
	return gin.H{
		"message": e.Err,
		"code": e.Code,
	}
}

func (e *RequestError) Log(){
	log.Println(e.Code, ": ", e.Err)
}