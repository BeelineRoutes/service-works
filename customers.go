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

type addressGet struct {
    Address, AddressLine2, CityName, StateName, Zip string 
    IsActive bool
}

func (this *addressGet) human () string {
    ret := this.Address
    if len(this.AddressLine2) > 0 {
        ret = ret + " " + this.AddressLine2
    }

    if len(this.CityName) > 0 {
        ret = ret + " " + this.CityName
    }

    if len(this.StateName) > 0 {
        ret = ret + ", " + this.StateName
    }

    if len(this.Zip) > 0 {
        ret = ret + " " + this.Zip
    }

    return ret 
}

type Address struct {
    AddressId, Type int
    AddressLine1, AddressLine2, Zip, City, State, Lat, Long, FirstName, LastName, PrimaryPhone, Email string 
    NotifyEmail, NotifyPrimaryPhone bool 
}

type Customer struct {
    FirstName, LastName, CompanyName, Email, PrimaryPhone string 
    CustomerId int 
    IsActive bool 

    Addresses []Address
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//


func (this *ServiceWorks) SearchCustomers (ctx context.Context, token, search string) ([]*Customer, error) {
    header := make(map[string]string)
    header["Token"] = token 

    params := url.Values{}
    params.Set("CustomerName", search)

    var resp struct {
        ApiStatus apiStatus
        Data struct {
            Customers []*Customer
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

// gets the customer's address from their id
func (this *ServiceWorks) GetCustomerAddress (ctx context.Context, token string, customerId int) (string, error) {
    header := make(map[string]string)
    header["Token"] = token 

    params := url.Values{}
    params.Set("customerId", fmt.Sprintf("%d", customerId))

    var resp struct {
        ApiStatus apiStatus
        Data []addressGet
    }
    
    errObj, err := this.send (ctx, http.MethodGet, fmt.Sprintf("Job/GetCustomerAddress?%s", params.Encode()), header, nil, &resp)
    if err != nil { return "", errors.WithStack(err) } // bail
    if errObj != nil { return "", errObj.Err() } // something else bad

    // see if the response was what was expected
    err = resp.ApiStatus.Error()
    if err != nil { return "", err }

    // find the "best" address for this person
    if len(resp.Data) == 0 { return "", nil } // no address found

    // look for the first one that's active
    for _, addr := range resp.Data {
        if addr.IsActive {
            ret := addr.human() 
            if len(ret) > 3 {
                return ret, nil 
            }
        }
    }

    // we couldn't find an active one, i'm just going to return the first
    return resp.Data[0].human(), nil 
}


// creates a new customer with the required info
func (this *ServiceWorks) CreateCustomer (ctx context.Context, token, firstName, lastName, email, phone, addr, addr2, zip, city, state string) (*Customer, error) {
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
            Customers []Customer
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
