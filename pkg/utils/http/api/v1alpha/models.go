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
	// Any structured data or null
	Data any `json:"data" swaggertype:"object,object"`
	// Error message. If status is not success, this field must be filled by a string with error message
	Error error `example:"null" json:"error" swaggertype:"string"`
} // @Name Response

// Version
// @Description Application version
type Version struct {
	Version string `example:"v1.0.0" json:"version"`
} // @Name VersionResponse

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

func (r Response) MarshalJSON() ([]byte, error) {
	type Alias Response

	var errMsg *string

	if r.Error != nil {
		msg := r.Error.Error()
		errMsg = &msg
	}

	return json.Marshal(&struct {
		Error *string `json:"error"`
		*Alias
	}{
		Error: errMsg,
		Alias: (*Alias)(&r),
	})
}
