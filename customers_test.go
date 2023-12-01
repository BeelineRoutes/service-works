
package serviceworks 

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"context"
	"time"
)

// this login should work
func TestSecondCustomer1 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of jobs, only unscheduled ones
	customers, err := sw.SearchCustomers (ctx, cfg.Token, "nate dogg")
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(customers) > 0)
	assert.Equal (t, 20144, customers[0].CustomerId)
	
}
