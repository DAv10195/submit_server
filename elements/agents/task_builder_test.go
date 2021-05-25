package agents

import (
	"errors"
	"fmt"
	submiterr "github.com/DAv10195/submit_commons/errors"
	"github.com/DAv10195/submit_server/db"
	"testing"
)

func createValidTask() error {
	builder := NewTaskBuilder(db.System, false)
	builder.WithCommand("echo David").WithExecTimeout(5).WithResponseHandler("handler")
	if _, err := builder.Build(); err != nil {
		return fmt.Errorf("error creating task with valid arguments: %v", err)
	}
	return nil
}

func emptyCommand() error {
	builder := NewTaskBuilder(db.System, false)
	builder.WithExecTimeout(5).WithResponseHandler("handler")
	if _, err := builder.Build(); err == nil {
		return errors.New("error not returned while building task with empty command")
	} else if _, ok := err.(*submiterr.ErrInsufficientData); !ok {
		return errors.New("returned error is not of type ErrInsufficientData")
	}
	return nil
}

func emptyRespHandler() error {
	builder := NewTaskBuilder(db.System, false)
	builder.WithCommand("echo David").WithExecTimeout(5)
	if _, err := builder.Build(); err == nil {
		return errors.New("error not returned while building task with empty response handler")
	} else if _, ok := err.(*submiterr.ErrInsufficientData); !ok {
		return errors.New("returned error is not of type ErrInsufficientData")
	}
	return nil
}

func negativeExecTimeout() error {
	builder := NewTaskBuilder(db.System, false)
	// default value for exec timeout in builder is already negative == -1
	builder.WithCommand("echo David").WithResponseHandler("handler")
	if _, err := builder.Build(); err == nil {
		return errors.New("error not returned while building task with negative (empty) exec timeout")
	} else if _, ok := err.(*submiterr.ErrInsufficientData); !ok {
		return errors.New("returned error is not of type ErrInsufficientData")
	}
	return nil
}

func TestBuilder(t *testing.T) {
	testFuncs := []func() error{negativeExecTimeout, emptyRespHandler, emptyCommand, createValidTask}
	for _, testFunc := range testFuncs {
		if err := testFunc(); err != nil {
			t.Fatal(err)
		}
	}
}
