package server

import (
	"encoding/json"
	"fmt"
	commons "github.com/DAv10195/submit_commons"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/users"
)

type agentTaskResponseHandler func([]byte, map[string]interface{}) error

var agentTaskRespHandlers = make(map[string]agentTaskResponseHandler)

// no-op handler for on demand tasks
func handleOnDemandTask(_ []byte, _ map[string]interface{}) error {
	return nil
}

type TestResponse struct {
	Grade		int			`json:"grade"`
	Output		string		`json:"output"`
}

// handle task responses which represent a test execution
func handleTestTask(payload []byte, labels map[string]interface{}) error {
	tr := &TestResponse{}
	if err := json.Unmarshal(payload, tr); err != nil {
		return err
	}
	if tr.Grade < 0 || tr.Grade > 100 {
		return fmt.Errorf("test task handler: grade in test task response (%d) is not >= 0 ^ <= 100", tr.Grade)
	}
	od, ok := labels[onDemandTask]
	if !ok {
		return fmt.Errorf("test task handler: missing label '%s' in task labels", onDemandTask)
	}
	onDemand, ok := od.(bool)
	if !ok {
		return fmt.Errorf("test task handler: label '%s' has a non boolean value", onDemandTask)
	}
	// if this is an on demand test execution, then notify the user which triggered it that the test finished executing
	if onDemand {
		se, ok := labels[onSubmitExec]
		if ok {
			submitExec, ok := se.(bool)
			if !ok {
				return fmt.Errorf("test task handler: label '%s' has a non boolean value", onSubmitExec)
			}
			if submitExec {
				auname, ok := labels[assInstUsrName]
				if !ok {
					return fmt.Errorf("test task handler: missing label '%s' in task labels", assInstUsrName)
				}
				assUsername, ok := auname.(string)
				if !ok {
					return fmt.Errorf("test task handler: label '%s' has a non string value", assInstUsrName)
				}
				assUser, err := users.Get(assUsername)
				if err != nil {
					return err
				}
				adname, ok := labels[assDefName]
				if !ok {
					return fmt.Errorf("test task handler: missing label '%s' in task labels", assDefName)
				}
				assDefinitionName, ok := adname.(string)
				if !ok {
					return fmt.Errorf("test task handler: label '%s' has a non string value", assDefName)
				}
				tn, ok := labels[testName]
				if !ok {
					return fmt.Errorf("test task handler: missing label '%s' in task labels", onDemandTask)
				}
				execTestName, ok := tn.(string)
				if !ok {
					return fmt.Errorf("test task handler: label '%s' has a non string value", onDemandTask)
				}
				if _, err := messages.NewMessage(db.System, fmt.Sprintf("execution of '%s' test on submit of '%s' assingment: grade: %d, output: '%s'", execTestName, assDefinitionName, tr.Grade, tr.Output), assUser.MessageBox, true); err != nil {
					return err
				}
			}
		}
		return nil
	}
	// well, not on demand, then mark the assignment as graded in the db
	adname, ok := labels[assDefName]
	if !ok {
		return fmt.Errorf("test task handler: missing label '%s' in task labels", assDefName)
	}
	assDefinitionName, ok := adname.(string)
	if !ok {
		return fmt.Errorf("test task handler: label '%s' has a non string value", assDefName)
	}
	auname, ok := labels[assInstUsrName]
	if !ok {
		return fmt.Errorf("test task handler: missing label '%s' in task labels", assInstUsrName)
	}
	assUsername, ok := auname.(string)
	if !ok {
		return fmt.Errorf("test task handler: label '%s' has a non string value", assInstUsrName)
	}
	assInstKey := fmt.Sprintf("%s%s%s", assDefinitionName, db.KeySeparator, assUsername)
	assInst, err := assignments.GetInstance(assInstKey)
	if err != nil {
		return err
	}
	assUser, err := users.Get(assUsername)
	if err != nil {
		return err
	}
	tn, ok := labels[testName]
	if !ok {
		return fmt.Errorf("test task handler: missing label '%s' in task labels", onDemandTask)
	}
	execTestName, ok := tn.(string)
	if !ok {
		return fmt.Errorf("test task handler: label '%s' has a non string value", onDemandTask)
	}
	msg, err := messages.NewMessage(db.System, fmt.Sprintf("execution of '%s' test for testing '%s' assingment: grade: %d, output: '%s'", execTestName, assDefinitionName, tr.Grade, tr.Output), assUser.MessageBox, true)
	if err != nil {
		return err
	}
	assInst.Grade = tr.Grade
	assInst.State = assignments.Graded
	return db.Update(db.System, assInst, msg)
}

// handle copy detection execution response
func handleMossTask(payload []byte, labels map[string]interface{}) error {
	mo := &submitws.MossOutput{}
	if err := json.Unmarshal(payload, mo); err != nil {
		return err
	}
	threshold := int(labels[mossCopyThreshold].(float64))
	assignment := labels[assDefName].(string)
	var assignmentsToMarkAsCopy []db.IBucketElement
	for _, mop := range mo.Pairs {
		if mop.Percentage1 >= threshold || mop.Percentage2 >= threshold {
			ass1, err := assignments.GetInstance(fmt.Sprintf("%s%s%s", assignment, db.KeySeparator, mop.Name1))
			if err != nil {
				return err
			}
			ass2, err := assignments.GetInstance(fmt.Sprintf("%s%s%s", assignment, db.KeySeparator, mop.Name2))
			if err != nil {
				return err
			}
			ass1.MarkedAsCopy = true
			ass2.MarkedAsCopy = true
			assignmentsToMarkAsCopy = append(assignmentsToMarkAsCopy, ass1, ass2)
		}
	}
	return db.Update(db.System, assignmentsToMarkAsCopy...)
}

func init() {
	agentTaskRespHandlers[onDemandTask] = handleOnDemandTask
	agentTaskRespHandlers[testTask] = handleTestTask
	agentTaskRespHandlers[commons.Moss] = handleMossTask
}
