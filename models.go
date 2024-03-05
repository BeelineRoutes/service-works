/** ****************************************************************************************************************** **
	Data objects
	Converted objects from the serviceworks api into go-lang equivilants

** ****************************************************************************************************************** **/

package serviceworks 

import (
	"github.com/pkg/errors"

	"net/http"
	"strings"
	"encoding/json"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type JobStatus int 

const (
	JobStatus_unassigned 	JobStatus = 1
	JobStatus_scheduled		JobStatus = 2
	JobStatus_unscheduled 	JobStatus = 7
	JobStatus_confirmed		JobStatus = 13
)

//----- ERRORS ---------------------------------------------------------------------------------------------------------//

var (
	ErrInvalidCode 		= errors.New("Token not valid")
	ErrInvalidUserPassword	= errors.New("Username or Password is invalid")
	ErrAuthExpired		= errors.New("Token expired")
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type apiStatus struct {
	Status int 
	Message string 
	Errors interface{}
}

func (this *apiStatus) Error () error {
	if this == nil { return errors.Errorf ("ApiStatus not found") }

	jstr, _ := json.Marshal(this.Errors) // i think this should give me the error object as a string

	// check the response
	if this.Status == 1 { return nil } // we're good

	// this means it failed
	// see if we know why
	if this.Status == 3 && strings.Contains(this.Message, "Username or Password is not valid") {
		return errors.WithStack(ErrInvalidUserPassword)
	}

	if this.Status == 0 && strings.Contains(this.Message, "Api Exception") {
		return errors.WithStack(ErrInvalidCode)
	}

	if this.Status == 0 && strings.EqualFold(this.Message, "No Jobs Found") {
		return nil // this is cool, just means we had no jobs from the search
	}

	if this.Status == 2 && strings.Contains(this.Message, "invalid token") {
		return errors.WithStack(ErrInvalidCode)
	}

	// we don't know what happened, create a generic error message
	return errors.Errorf ("Bad response. Got status %d :: message '%s' :: %s", this.Status, this.Message, string(jstr))
}

//----- ERRORS ---------------------------------------------------------------------------------------------------------//
type Error struct {
	ErrMsg, Description string 
	StatusCode int 
}

func (this *Error) UnmarshalJSON (b []byte) error {

	// try this way
	var one struct {
		Error struct {
			Message string 
		}
	}

	err := json.Unmarshal(b, &one)
	if err == nil && len(one.Error.Message) > 0 {
		this.ErrMsg = one.Error.Message

		if strings.Contains(this.ErrMsg, "archived job") {
			this.StatusCode = http.StatusGone // it's gone 410
		}

	} else {
		// that didn't work, try another format
		var two struct {
			Error string 
			Description string `json:"error_description"`
			StatusCode int
		}

		err = json.Unmarshal (b, &two)
		if err == nil {
			this.ErrMsg = two.Error 
			this.Description = two.Description
			this.StatusCode = two.StatusCode
		} else {
			this.ErrMsg = err.Error()
			this.Description = string(b)
		}
	}

	if len(this.ErrMsg) == 0 {
		// this didn't work
		this.ErrMsg = "Unkown struct type"
		this.Description = string(b)
	}
	return nil 
}

func (this *Error) Err () error {
	if this == nil { return nil } // no error
	
	if this.ErrMsg == "invalid_grant" { // this is for granting access based on the passed code
		return errors.Wrap (ErrInvalidCode, this.Description)
	}

	switch this.StatusCode {
	case http.StatusUnauthorized:
		return errors.Wrap (ErrAuthExpired, this.Description) // invalid for another reason, most likely the oauth has been revoked
	
	}
	// just a default
	return errors.Errorf ("ServiceWorks Error : %d : %s : %s", this.StatusCode, this.ErrMsg, this.Description)
}

func wrapErr (err error, req interface{}, resp interface{}) error {
	if err == nil { return nil }

	// otherwiset create a more useful error
	jReq, _ := json.Marshal(req)
    jResp, _ := json.Marshal(resp)

	return errors.Wrapf(err, "%s :: %s", string(jReq), string(jResp))
}

//----- PUBLIC ---------------------------------------------------------------------------------------------------------//

type ServiceWorks struct {
	Url string // this is so we can switch between production and a qa endpoint
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

