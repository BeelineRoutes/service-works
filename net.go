/** ****************************************************************************************************************** **
	The actual sending and receiving stuff
	Reused for most of the calls to serviceworks
	
** ****************************************************************************************************************** **/

package serviceworks 

import (
    "github.com/pkg/errors"

    "fmt"
    "net/http"
    "context"
    "encoding/json"
    "io/ioutil"
    "bytes"
	"strings"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// handles making the request and reading the results from it 
// if there's an error the Error object will be set, otherwise it will be nil
func (this *ServiceWorks) finish (req *http.Request, out interface{}) (*Error, error) {
	resp, err := http.DefaultClient.Do (req)
	
	if err != nil { return nil, errors.WithStack (err) }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll (resp.Body)

	if resp.StatusCode == http.StatusGone {
		// this means that the job/estimate was deleted 
		errObj := &Error{
			StatusCode: resp.StatusCode,
		}
		return errObj, nil

	} else if resp.StatusCode > 499 { 
		// 500 level errors seem to not share the same error object
		errObj := &Error{}
		errObj.ErrMsg = string(body) // dump the whole body in here
        errObj.StatusCode = resp.StatusCode // if it didn't get an error code, set it
		
        return errObj, nil

	} else if resp.StatusCode > 399 { 
		errObj := &Error{}
		json.Unmarshal (body, errObj)

		if errObj.StatusCode == 0 {
			errObj.StatusCode = resp.StatusCode // if it didn't get an error code, set it
		}
		return errObj, nil
	}
	
	if out != nil { err = errors.WithStack (json.Unmarshal (body, out)) }
	
	return nil, err // we're good
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *ServiceWorks) send (ctx context.Context, requestType, link string, header map[string]string, 
						in, out interface{}) (*Error, error) {
	var jstr []byte 
	var err error 

	if in != nil {
		jstr, err = json.Marshal (in)
		if err != nil { return nil, errors.WithStack (err) }

		header["Content-Type"] = "application/json; charset=utf-8"
	}

	// make sure we got a url
	if len(this.Url) == 0 {
		this.Url = "https://apiapp.service.works/api" // production url
	} else if strings.HasSuffix (this.Url, "api") == false {
		if strings.HasSuffix (this.Url, "/") {
			this.Url += "api"
		} else {
			this.Url += "/api"
		}
	}
	
	req, err := http.NewRequestWithContext (ctx, requestType, fmt.Sprintf ("%s/%s", this.Url, link), bytes.NewBuffer(jstr))
	if err != nil { return nil, errors.Wrap (err, link) }

	for key, val := range header { req.Header.Set (key, val) }
	errObj, err := this.finish (req, out)
	
	return errObj, errors.Wrapf (err, " %s : %s", link, string(jstr))
}

func (this *ServiceWorks) defaultHeader (token string) (map[string]string) {
	header := make(map[string]string)
	header["Token"] = token 
	return header 
}