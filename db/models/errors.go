package models

type RequestError struct {
	Err error
	StatusCode int
}