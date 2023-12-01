
package serviceworks 

import (
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"

	"testing"
	"context"
	"time"
)

// this login should work
func TestSecondCrew1 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of jobs, only unscheduled ones
	crew, err := sw.CrewList (ctx, cfg.Token)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(crew) > 0)
	assert.Equal (t, 1694, crew[0].EmployeeID)
	
}

// this should fail for a bad token
func TestSecondCrew2 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of jobs, only unscheduled ones
	_, err := sw.CrewList (ctx, cfg.Token + "a")
	if err == nil { t.Fatal("we were expecting an error") }

	assert.Equal (t, ErrInvalidCode, errors.Cause(err))	
}
