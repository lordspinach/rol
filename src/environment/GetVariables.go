package environment

import (
	"os"
	"rol/app/errors"
	"strconv"
	"strings"
)

var (
	IsDebug bool
)

func GetEnvVariables() error {
	var err error
	debug := os.Getenv("DEBUG")
	if strings.ToLower(debug) != "true" {
		return nil
	}
	IsDebug, err = strconv.ParseBool(debug)
	if err != nil {
		return errors.Internal.Wrap(err, "parse string to bool failed")
	}
	return nil
}
