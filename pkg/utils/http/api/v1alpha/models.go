package v1alpha

import "encoding/json"

type Status int

// Response wrapper
// @Description Response wrapper to not build the API on top of outdated HTTP codes set
type Response struct {
	// Response status
	// * success - everything is OK
	// * error   - something went wrong
	// * warning - something went wrong, but it's not critical
	Status Status `enums:"success,error,warning" example:"success" json:"status"`
	// Any data
	Data any `json:"data"`
	// Error message. If status is not success, this field must be filled by a string with error message
	Error error `example:"null" json:"error"`
} // @Name Response

// Version
// @Description Application version
type Version struct {
	Version string `example:"v1.0.0" json:"version"`
} // @Name Version

const (
	StatusSuccess Status = iota
	StatusError
	StatusWarning
)

var statusName = map[Status]string{
	StatusSuccess: "success",
	StatusError:   "error",
	StatusWarning: "warning",
}

func (rs Status) String() string {
	return statusName[rs]
}

func (rs Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(rs.String()) //nolint:wrapcheck
}
