package api

type UserContextKey string

const (
	APIInstanceContextKey = UserContextKey("apiInstance")
	RequestIDContextKey   = UserContextKey("requestID")
)
