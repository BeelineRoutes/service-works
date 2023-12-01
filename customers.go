/** ****************************************************************************************************************** **
    Customer related calls
    
** ****************************************************************************************************************** **/

package serviceworks 

import (
    "github.com/pkg/errors"
    
    "fmt"
    "net/http"
    "context"
    "net/url"
    "strings"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Address struct {
    AddressId, Type int
    AddressLine1, AddressLine2, Zip, City, State, Lat, Long, FirstName, LastName, PrimaryPhone, Email string 
    NotifyEmail, NotifyPrimaryPhone bool 
}

type customer struct {
    FirstName, LastName, CompanyName, Email, PrimaryPhone string 
    CustomerId int 
    IsActive bool 

    Addresses []Address
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//


func (this *ServiceWorks) SearchCustomers (ctx context.Context, token, search string) ([]customer, error) {
    header := make(map[string]string)
    header["Token"] = token 

    params := url.Values{}
    params.Set("CustomerName", search)

    var resp struct {
        ApiStatus apiStatus
        Data struct {
            Customers []customer
        }
    }
    
    errObj, err := this.send (ctx, http.MethodGet, fmt.Sprintf("Job/GetCustomerSearch?%s", params.Encode()), header, nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    err = resp.ApiStatus.Error()
    if err != nil { return nil, err }

    // for whatever reason they still return a single customer to tell us they didn't find a crew member, so handle that as an empty response
    if len(resp.Data.Customers) == 1 {
        if strings.EqualFold (resp.Data.Customers[0].FirstName, "No Customer Found") {
            return nil, nil // didn't find anyone
        }
    }
    
    return resp.Data.Customers, nil // and return
}

// creates a new customer with the required info
func (this *ServiceWorks) CreateCustomer (ctx context.Context, token, firstName, lastName, email, phone, addr, addr2, zip, city, state string) (*customer, error) {
    header := make(map[string]string)
    header["Token"] = token 

    var request struct {
        FirstName, LastName, CustomerType, Email, PrimaryPhone string 
        CustomerId int 
        IsSendEmail, IsSendSms bool 

        Address Address `json:"address"`
    }

    request.FirstName = firstName
    request.LastName = lastName
    request.CustomerType = "0"
    request.Email = email
    request.PrimaryPhone = phone 
    request.IsSendEmail = len(email) > 0
    request.IsSendSms = len(phone) > 0 

    request.Address.AddressLine1 = addr
    request.Address.AddressLine2 = addr2
    request.Address.Zip = zip 
    request.Address.City = city 
    request.Address.State = state 
    request.Address.Type = 2
    request.Address.FirstName = firstName
    request.Address.LastName = lastName
    request.Address.PrimaryPhone = phone 
    request.Address.Email = email 
    request.Address.NotifyEmail = request.IsSendEmail
    request.Address.NotifyPrimaryPhone = request.IsSendSms

    var resp struct {
        ApiStatus apiStatus
        Data struct {
            Customers []customer
        }
    }
    
    errObj, err := this.send (ctx, http.MethodPost, "Job/AddEditCustomerDetail", header, &request, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    err = resp.ApiStatus.Error()
    if err != nil { return nil, err }

    if len(resp.Data.Customers) == 0 {
        return nil, errors.Errorf("New customer was not created")
    }

    return &resp.Data.Customers[0], nil // and return
}
