
package serviceworks 

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"context"
	"time"
)

// time ranges
func TestSecondJobs1 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of jobs, only unscheduled ones
	ranges, err := sw.JobsListTimeRanges (ctx, cfg.Token)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(ranges) > 0)
	assert.Equal (t, true, len(ranges[0].Id) > 0)
	assert.Equal (t, true, len(ranges[0].Text) > 0)
	
}

// list the jobs
func TestSecondJobs2 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	start, err := time.Parse("2006-01-02", "2023-11-30")
	if err != nil { t.Fatal (err) }

	end, err := time.Parse("2006-01-02", "2023-12-01")
	if err != nil { t.Fatal (err) }

	// get our list of jobs, only unscheduled ones
	jobs, err := sw.ListJobs (ctx, cfg.Token, start, end)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(jobs) > 0)
	assert.Equal (t, true, jobs[0].TicketId > 0)
	assert.Equal (t, true, len(jobs[0].CustomerAddress) > 0)
	assert.Equal (t, true, jobs[0].TripAssignmentId > 0)
	
}
