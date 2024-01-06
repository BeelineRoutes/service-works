/** ****************************************************************************************************************** **
    Handles authentication and authorization

** ****************************************************************************************************************** **/

package serviceworks 

import (
    "fmt"
    "net/http"
    "context"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type respLogin struct {
    apiStatus
    Companyid int 
    Token, Phone, PhoneCode, Email, TimeZoneName string 
}

func (this *respLogin) response () (*RespLogin, error) {
    err := this.Error()
    if err != nil { return nil, err } // our status failed

    // we're good
    return &RespLogin {
        Token: this.Token,
        Phone: this.PhoneCode + this.Phone,
        TimeZoneName: this.TimeZoneName,
        CompanyId: fmt.Sprintf("%d", this.Companyid),
    }, nil
}

type RespLogin struct {
    Token, Phone, TimeZoneName, CompanyId string
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

//----- LOGIN -------------------------------------------------------------------------------------------------------//

// Takes the passed code we got from the params of the redirect url and converts it to long-live token and refresh token
func (this *ServiceWorks) Login (ctx context.Context, username, password, apikey string) (*RespLogin, error) {
    
    header := make(map[string]string)
    header["UserName"] = username
    header["Password"] = password
    header["ApiKey"] = apikey

    resp := &respLogin{}

    // make our call
    errObj, err := this.send (ctx, http.MethodPost, "Login/LoginWithKey", header, nil, resp)
    if err != nil { return nil, err } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    return resp.response() // return what we got
}

// refreshes the token
// service works tells me i need to do this every 30 days
func (this *ServiceWorks) RefreshToken (ctx context.Context, token string) (*RespLogin, error) {
    
    header := make(map[string]string)
    header["Token"] = token

    resp := &respLogin{}

    // make our call
    errObj, err := this.send (ctx, http.MethodPost, "Login/RefreshToken", header, nil, resp)
    if err != nil { return nil, err } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    return resp.response() // return what we got
}
