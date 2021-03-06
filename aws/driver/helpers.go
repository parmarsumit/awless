package awsdriver

import (
	"fmt"
	"reflect"
	"time"

	"github.com/wallix/awless/logger"
	"github.com/wallix/awless/template"
	"github.com/wallix/awless/template/driver"
)

func DefaultTemplateEnv() *template.Env {
	env := template.NewEnv()
	env.DefLookupFunc = AWSLookupDefinitions
	return env
}

func AWSLookupDefinitions(key string) (t template.Definition, ok bool) {
	t, ok = AWSTemplatesDefinitions[key]
	return
}

type driverCall struct {
	d       driver.Driver
	fn      interface{}
	logger  *logger.Logger
	desc    string
	setters []setter
}

func (dc *driverCall) execute(input interface{}) (output interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			output = nil
			err = fmt.Errorf("%s", e)
		}
	}()

	for _, s := range dc.setters {
		if err = s.set(input); err != nil {
			return nil, err
		}
	}

	fnVal := reflect.ValueOf(dc.fn)
	values := []reflect.Value{reflect.ValueOf(input)}

	start := time.Now()
	results := fnVal.Call(values)

	if err, ok := results[1].Interface().(error); ok && err != nil {
		return nil, fmt.Errorf("%s: %s", dc.desc, err)
	}

	dc.logger.ExtraVerbosef("%s call took %s", dc.desc, time.Since(start))
	dc.logger.Verbosef("%s done", dc.desc)

	output = results[0].Interface()

	return
}
