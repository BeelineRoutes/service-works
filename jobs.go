/** ****************************************************************************************************************** **
	Calls related to jobs

    There's a couple of filters used for requesting jobs, these have been broken out into their own functions.

    Updating jobs allows for setting a target time as well as an employee.
    Multiple employees may be assigned as well using UpdateJobDispatch
** ****************************************************************************************************************** **/

package serviceworks 

import (
    "github.com/pkg/errors"

    "fmt"
    "net/http"
    "net/url"
    "context"
    "time"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type TimeRange struct {
    Id, Text string
}

type Assignment struct {
    TripAssignmentId, TripNo, Duration, TimeRangeId int 
    AssignDateTime, TimeRange string 

    AssignmentDetails []struct {
        TripDetailsId, EmployeeId int
        EmployeeName string 
    }

    TripList []struct {
        TripDetailsId, EmployeeId int
        CompletionTime time.Time 
    }
}

type Job struct {
    TicketId int 
    IssueDescription, TicketStatus string 
    IsActive bool 

    Customer customer

    Assignments []Assignment
}

type jobTech struct {
    EmployeeId int 
    IsSupervisor bool 
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *ServiceWorks) JobsListTimeRanges (ctx context.Context, token string) ([]TimeRange, error) {
    header := make(map[string]string)
    header["Token"] = token 

    var resp struct {
        ApiStatus apiStatus
        Data []TimeRange 
    }
    
    errObj, err := this.send (ctx, http.MethodGet, "Job/GetTimeRange", header, nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    err = resp.ApiStatus.Error()
    return resp.Data, wrapErr(err, nil, resp) // and return
}

// creates a new job
func (this *ServiceWorks) JobCreate (ctx context.Context, token, issueDesc string, customerId, duration, timeRangeId int, target time.Time, 
                                    employeeIds []int) (*Job, error) {
    header := make(map[string]string)
    header["Token"] = token 

    // first request is to create the new job
    var req struct {
        CustomerId, Duration, TimeRangeId int
        IssueDescription, AssignDateTime, AssignTime string
    }

    req.CustomerId = customerId
    req.IssueDescription = issueDesc
    req.AssignDateTime = target.Format("01/02/2006 15:04:00")
    req.Duration = duration
    req.AssignTime = "TimeRange"
    req.TimeRangeId = timeRangeId

    var resp struct {
        ApiStatus apiStatus
        Jobs []Job
    }
    
    errObj, err := this.send (ctx, http.MethodPost, "Job/CreateNewJob", header, &req, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    
    err = resp.ApiStatus.Error()
    if err != nil { return nil, wrapErr(err, req, resp) }

    // make sure we got a job
    if len(resp.Jobs) == 0 { return nil, wrapErr(errors.Errorf("Didn't get any jobs back"), req, resp) }
    
    j := resp.Jobs[0] // so it's easier to reference

    if len(j.Assignments) == 0 { return nil, wrapErr(errors.Errorf("Didn't get any job assignments back"), req, resp) }
    a := j.Assignments[0]

    // now we create the trip schedule and assign it to these employees
    err = this.JobUpdate (ctx, token, j.TicketId, a.TripAssignmentId, duration, timeRangeId, a.TripNo, target, employeeIds)
    return &j, err // and return
}

// updates the arrival time or assigned crew or both for an existing job
func (this *ServiceWorks) JobUpdate (ctx context.Context, token string, ticketId, tripAssignId, duration, timeRangeId, tripNo int, target time.Time, employeeIds []int) error {
    header := make(map[string]string)
    header["Token"] = token 

    // first request is to create the new job
    var req struct {
        TicketId, TripAssignmentId, Duration, TimeRangeId, TripNo int
        IssueDescription, AssignDateTime, AssignTime string

        Technicians []jobTech
    }

    req.TicketId = ticketId
    req.TripAssignmentId = tripAssignId
    req.AssignDateTime = target.Format("01/02/2006 15:04:00")
    req.Duration = duration
    req.AssignTime = "TimeRange"
    req.TimeRangeId = timeRangeId
    req.TripNo = tripNo

    // add in the job techs
    for _, id := range employeeIds {
        req.Technicians = append (req.Technicians, jobTech { EmployeeId: id })
    }

    // make the first one a supervisor
    if len(req.Technicians) > 0 {
        req.Technicians[0].IsSupervisor = true 
    }

    var resp struct {
        ApiStatus apiStatus
    }
    
    errObj, err := this.send (ctx, http.MethodPost, "Job/SaveSchedule", header, &req, &resp)
    if err != nil { return errors.WithStack(err) } // bail
    if errObj != nil { return errObj.Err() } // something else bad

    // see if the response was what was expected
    return wrapErr(resp.ApiStatus.Error(), req, resp)
}


func (this *ServiceWorks) ListJobs (ctx context.Context, token string, start, finish time.Time) ([]Job, error) {
    params := url.Values{}
    params.Set("fromdate", start.Format("01/02/2006"))
    params.Set("todate", finish.Format("01/02/2006"))

    var resp struct {
        ApiStatus apiStatus
        Jobs []Job
    }
    
    errObj, err := this.send (ctx, http.MethodGet, fmt.Sprintf("Job/GetJob?%s", params.Encode()), this.defaultHeader(token), nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    err = wrapErr(resp.ApiStatus.Error(), nil, resp)
    return resp.Jobs, err // and return
}

