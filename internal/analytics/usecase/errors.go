package usecase

import "errors"

var (
	ErrAnalyticsInvalidFilter = errors.New("analytics invalid filter")
	ErrAnalyticsForbidden     = errors.New("analytics forbidden")
	ErrAnalyticsNoData        = errors.New("analytics no data")
)
