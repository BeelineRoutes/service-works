
package serviceworks 

import (
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"

	"testing"
	"context"
	"time"
)

// this login should work
func TestFirstLogin1 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// login
	login, err := sw.Login (ctx, cfg.Username, cfg.Password, cfg.ApiKey)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(login.Token) > 10)
	assert.Equal (t, "373", login.CompanyId)
	assert.Equal (t, "Atlantic Standard Time", login.TimeZoneName)

	// store the login token in our local config for future calls
	cfg.Token  = login.Token 
	saveConfig (t, cfg)
}

// this should fail - bad password
func TestFirstLogin2 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// login
	_, err := sw.Login (ctx, cfg.Username, "asdf", cfg.ApiKey)
	if err == nil { t.Fatal ("expecting an error") }

	assert.Equal (t, ErrInvalidUserPassword, errors.Cause(err))
}

// this should fail - bad api key
/*
func TestFirstLogin2b (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// login
	_, err := sw.Login (ctx, cfg.Username, cfg.Password, "asdf")
	if err == nil { t.Fatal ("expecting an error") }

	// assert.Equal (t, ErrInvalidUserPassword, errors.Cause(err))
	t.Logf("%v\n", err)
	t.Logf("%s\n", errors.Cause(err))
}
*/


func TestRefreshLogin1 (t *testing.T) {
	sw, cfg := newServiceWorks (t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// refresh
	login, err := sw.RefreshToken (ctx, cfg.Token)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(login.Token) > 10)
	assert.Equal (t, "373", login.CompanyId)
	assert.Equal (t, "Atlantic Standard Time", login.TimeZoneName)

	// now make sure that it's different? not sure if it changes actually
	assert.NotEqual (t, cfg.Token, login.Token)

	// store the login token in our local config for future calls
	cfg.Token  = login.Token 
	saveConfig (t, cfg)
}
