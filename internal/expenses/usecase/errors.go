package usecase

import "errors"

var (
	ErrExpenseNotFound      = errors.New("expense not found")
	ErrExpenseForbidden     = errors.New("expense not allowed")
	ErrExpenseInvalidFilter = errors.New("expense invalid filter")
)
