
package serviceworks 

import (
	
	// "github.com/stretchr/testify/assert"
	//"github.com/pkg/errors"

	"testing"
	"os"
	"encoding/json"
)

// used for testing so I don't have my actual credentials in the repo
type testConfig struct {
	Username, Password, ApiKey, Token string 
}

func saveConfig (t *testing.T, cfg *testConfig) {
	jstr, err := json.Marshal(cfg)
	if err != nil { t.Fatal (err) }

	err = os.WriteFile("test.cfg", jstr, 0666)
	if err != nil { t.Fatal (err) }
}

func newServiceWorks (t *testing.T) (*ServiceWorks, *testConfig) {
	// read our local config
	config, err := os.Open("test.cfg")
	if err != nil { t.Fatal (err) }

	cfg := &testConfig{}

	jsonParser := json.NewDecoder (config)
	err = jsonParser.Decode (cfg)
	if err != nil { t.Fatal (err) }

	sw := &ServiceWorks {
		Url: "http://65.61.142.55:82", // connect to the qa server by default
	}

	return sw, cfg
}
