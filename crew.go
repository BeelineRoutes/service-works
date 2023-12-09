/** ****************************************************************************************************************** **
	Calls related to jobs

    There's a couple of filters used for requesting jobs, these have been broken out into their own functions.

    Updating jobs allows for setting a target time as well as an employee.
    Multiple employees may be assigned as well using UpdateJobDispatch
** ****************************************************************************************************************** **/

package serviceworks 

import (
    "github.com/pkg/errors"
    
    "net/http"
    "context"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Employee struct {
    EmployeeID int
    FirstName, LastName, Address, Zip, CityName, State, Phone, Email, UserId, Color string 
    IsTechnician, IsActive bool 
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *ServiceWorks) CrewList (ctx context.Context, token string) ([]*Employee, error) {
    header := make(map[string]string)
    header["Token"] = token 

    var resp struct {
        ApiStatus apiStatus
        Data struct {
            EmployeeList []*Employee
        }
    }
    
    errObj, err := this.send (ctx, http.MethodGet, "Configuration/GetUserLists", header, nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    err = resp.ApiStatus.Error()
    return resp.Data.EmployeeList, err // and return
}
