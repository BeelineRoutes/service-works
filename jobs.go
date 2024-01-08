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
    AssignDateTime, TimeRange, AssignTime string 

    AssignmentDetails []struct {
        TripDetailsId, EmployeeId int
        EmployeeName string 
    }

    TripList []struct {
        TripDetailsId, EmployeeId int
        CompletionTime string 
    }
}

type Job struct {
    TicketStatusId JobStatus
    TicketId, Duration int 
    IssueDescription, TicketStatus string

    TripAssignmentId, TripNo, TimeRangeId int 
    AssignDateTime, AssignTime, Team, TeamIds string 

    CustomerId int 
    CustomerName, CustomerAddress, ContactPhone string 
}

func (this *Job) IsUnscheduled () bool {
    switch this.TicketStatusId {
    case JobStatus_unassigned, JobStatus_scheduled:
        return true 
    }
    return false 
}

type jobCreate struct {
    TicketStatusId JobStatus
    TicketId int 
    IssueDescription, TicketStatus string 
    
    Customer Customer

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
        Jobs []jobCreate
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

    ret := &Job {
        TicketId: j.TicketId,
        TicketStatusId: j.TicketStatusId,
        IssueDescription: j.IssueDescription,
        TicketStatus: j.TicketStatus,

        CustomerId: j.Customer.CustomerId,
        CustomerName: fmt.Sprintf("%s %s", j.Customer.FirstName, j.Customer.LastName),
        ContactPhone: j.Customer.PrimaryPhone,
        
        Duration: a.Duration,
        TripAssignmentId: a.TripAssignmentId,
        TripNo: a.TripNo,
        TimeRangeId: a.TimeRangeId,
        AssignDateTime: a.AssignDateTime,
        AssignTime: a.AssignTime,
    }

    ret.TeamIds = fmt.Sprintf("%d", employeeIds[0]) // just use the first

    if len(j.Customer.Addresses) > 0 {
        ret.CustomerAddress = fmt.Sprintf ("%s %s %s, %s %s", j.Customer.Addresses[0].AddressLine1, j.Customer.Addresses[0].AddressLine2, 
                            j.Customer.Addresses[0].City, j.Customer.Addresses[0].State, j.Customer.Addresses[0].Zip)
    }

    // now we create the trip schedule and assign it to these employees
    err = this.JobUpdate (ctx, token, j.TicketId, a.TripAssignmentId, duration, timeRangeId, a.TripNo, target, employeeIds)
    return ret, err // and return
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


func (this *ServiceWorks) ListJobs (ctx context.Context, token string, start, finish time.Time) ([]*Job, error) {
    params := url.Values{}
    params.Set("fromdate", start.Format("01/02/2006"))
    params.Set("todate", finish.Format("01/02/2006"))
    params.Set("roleId", "0")
    params.Set("isDateTrue", "true")

    var resp struct {
        ApiStatus apiStatus
        Data []*Job
    }
    
    errObj, err := this.send (ctx, http.MethodGet, fmt.Sprintf("Job/GetApiJobForSearch?%s", params.Encode()), this.defaultHeader(token), nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    // see if the response was what was expected
    err = wrapErr(resp.ApiStatus.Error(), nil, resp)
    return resp.Data, err // and return
}

