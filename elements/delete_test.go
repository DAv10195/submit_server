package elements

import (
	"fmt"
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/appeals"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/courses"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/elements/users"
	"testing"
)

func TestDeleteUser(t *testing.T) {
	defer db.InitDbForTest()()
	box1 := messages.NewMessageBox()
	msg1 := &messages.Message{}
	msg1.ID = commons.GenerateUniqueId()
	box1.Messages.Add(msg1.ID)
	box2 := messages.NewMessageBox()
	msg2 := &messages.Message{}
	msg2.ID = commons.GenerateUniqueId()
	box2.Messages.Add(msg2.ID)
	assDefId, userName := commons.GenerateUniqueId(), commons.GenerateUniqueId()
	assInstId := fmt.Sprintf("%s%s%s", assDefId, db.KeySeparator, userName)
	appeal := &appeals.Appeal{AssignmentInstance: assInstId, MessageBox: box1.ID}
	ass := &assignments.AssignmentInstance{AssignmentDef: assDefId, UserName: userName}
	user := &users.User{UserName: userName, MessageBox: box2.ID}
	if err := db.Update(db.System, box1, box2, msg1, msg2, appeal, ass, user); err != nil {
		t.Fatalf("error elements for test: %v", err)
	}
	if err := users.Delete(user, false); err != nil {
		t.Fatalf("error deleting assignment instance: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg1.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg2.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box1.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although appeal was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box2.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although user was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Appeals), []byte(appeal.AssignmentInstance)); err != nil {
		t.Fatalf("error checking if appeal exists: %v", err)
	} else if exists {
		t.Fatal("appeal exists although assignment instance was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInstId)); err != nil {
		t.Fatalf("error checking if assignment instance exists: %v", err)
	} else if exists {
		t.Fatal("assignment instance exists although user was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(user.UserName)); err != nil {
		t.Fatalf("error checking if user exists: %v", err)
	} else if exists {
		t.Fatal("user exists although it was deleted")
	}
}

func TestDeleteCourse(t *testing.T) {
	defer db.InitDbForTest()()
	box1 := messages.NewMessageBox()
	msg1 := &messages.Message{}
	msg1.ID = commons.GenerateUniqueId()
	box1.Messages.Add(msg1.ID)
	box2 := messages.NewMessageBox()
	msg2 := &messages.Message{}
	msg2.ID = commons.GenerateUniqueId()
	box2.Messages.Add(msg2.ID)
	courseNumber, courseYear, assDefName := 1234, 1234, commons.GenerateUniqueId()
	courseId := fmt.Sprintf("%d%s%d", courseNumber, db.KeySeparator, courseYear)
	assDefId, userName := fmt.Sprintf("%s%s%s", courseId, db.KeySeparator, assDefName), commons.GenerateUniqueId()
	assInstId := fmt.Sprintf("%s%s%s", assDefId, db.KeySeparator, userName)
	appeal := &appeals.Appeal{AssignmentInstance: assInstId, MessageBox: box1.ID}
	assInst := &assignments.AssignmentInstance{AssignmentDef: assDefId, UserName: userName}
	test := &tests.Test{AssignmentDef: assDefId, MessageBox: box2.ID}
	ass := &assignments.AssignmentDef{Course: courseId, Name: assDefName}
	course := &courses.Course{Number: courseNumber, Year: courseYear}
	if err := db.Update(db.System, box1, msg1, appeal, assInst, ass, course, test, box2, msg2); err != nil {
		t.Fatalf("error creating elements for test: %v", err)
	}
	if err := courses.Delete(course, false); err != nil {
		t.Fatalf("error deleting course: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg1.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box1.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although appeal was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg2.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box2.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although test was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Appeals), []byte(appeal.AssignmentInstance)); err != nil {
		t.Fatalf("error checking if appeal exists: %v", err)
	} else if exists {
		t.Fatal("appeal exists although assignment instance was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInstId)); err != nil {
		t.Fatalf("error checking if assignment instance exists: %v", err)
	} else if exists {
		t.Fatal("assignment instance exists def was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.AssignmentDefinitions), []byte(assDefId)); err != nil {
		t.Fatalf("error checking if assignment def exists: %v", err)
	} else if exists {
		t.Fatal("assignment def exists although course was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(courseId)); err != nil {
		t.Fatalf("error checking if course exists: %v", err)
	} else if exists {
		t.Fatal("course exists although it was deleted")
	}
}

func TestDeleteTest(t *testing.T) {
	defer db.InitDbForTest()()
	box := messages.NewMessageBox()
	msg := &messages.Message{}
	msg.ID = commons.GenerateUniqueId()
	box.Messages.Add(msg.ID)
	assId, testName := commons.GenerateUniqueId(), commons.GenerateUniqueId()
	test := &tests.Test{AssignmentDef: assId, Name: testName, MessageBox: box.ID}
	if err := db.Update(db.System, box, msg, test); err != nil {
		t.Fatalf("error creating elements for test: %v", err)
	}
	if err := tests.Delete(test, false); err != nil {
		t.Fatalf("error deleting test: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although test was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Tests), test.Key()); err != nil {
		t.Fatalf("error checking if tests exists: %v", err)
	} else if exists {
		t.Fatal("test exists although it was deleted")
	}
}

func TestDeleteMessageBox(t *testing.T) {
	defer db.InitDbForTest()()
	box := messages.NewMessageBox()
	msg := &messages.Message{}
	msg.ID = commons.GenerateUniqueId()
	box.Messages.Add(msg.ID)
	if err := db.Update(db.System, box, msg); err != nil {
		t.Fatalf("error creating elements for test: %v", err)
	}
	if err := messages.Delete(box); err != nil {
		t.Fatalf("error deleting mesage box: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although it was deleted")
	}
}

func TestDeleteAssignmentInstance(t *testing.T) {
	defer db.InitDbForTest()()
	box := messages.NewMessageBox()
	msg := &messages.Message{}
	msg.ID = commons.GenerateUniqueId()
	box.Messages.Add(msg.ID)
	assDefId, userName := commons.GenerateUniqueId(), commons.GenerateUniqueId()
	assInstId := fmt.Sprintf("%s%s%s", assDefId, db.KeySeparator, userName)
	appeal := &appeals.Appeal{AssignmentInstance: assInstId, MessageBox: box.ID}
	ass := &assignments.AssignmentInstance{AssignmentDef: assDefId, UserName: userName}
	if err := db.Update(db.System, box, msg, appeal, ass); err != nil {
		t.Fatalf("error creating elements for test: %v", err)
	}
	if err := assignments.DeleteInstance(ass, false); err != nil {
		t.Fatalf("error deleting assignment instance: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although appeal was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Appeals), []byte(appeal.AssignmentInstance)); err != nil {
		t.Fatalf("error checking if appeal exists: %v", err)
	} else if exists {
		t.Fatal("appeal exists although assignment instance was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInstId)); err != nil {
		t.Fatalf("error checking if assignment instance exists: %v", err)
	} else if exists {
		t.Fatal("assignment instance exists although it was deleted")
	}
}

func TestDeleteAssignmentDefinition(t *testing.T) {
	defer db.InitDbForTest()()
	box := messages.NewMessageBox()
	msg := &messages.Message{}
	msg.ID = commons.GenerateUniqueId()
	box.Messages.Add(msg.ID)
	courseId, assDefName := commons.GenerateUniqueId(), commons.GenerateUniqueId()
	assDefId, userName := fmt.Sprintf("%s%s%s", courseId, db.KeySeparator, assDefName), commons.GenerateUniqueId()
	assInstId := fmt.Sprintf("%s%s%s", assDefId, db.KeySeparator, userName)
	appeal := &appeals.Appeal{AssignmentInstance: assInstId, MessageBox: box.ID}
	assInst := &assignments.AssignmentInstance{AssignmentDef: assDefId, UserName: userName}
	ass := &assignments.AssignmentDef{Course: courseId, Name: assDefName}
	if err := db.Update(db.System, box, msg, appeal, assInst, ass); err != nil {
		t.Fatalf("error creating elements for test: %v", err)
	}
	if err := assignments.DeleteDef(ass, false); err != nil {
		t.Fatalf("error deleting assignment definition: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although appeal was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Appeals), []byte(appeal.AssignmentInstance)); err != nil {
		t.Fatalf("error checking if appeal exists: %v", err)
	} else if exists {
		t.Fatal("appeal exists although assignment instance was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInstId)); err != nil {
		t.Fatalf("error checking if assignment instance exists: %v", err)
	} else if exists {
		t.Fatal("assignment instance exists although def was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.AssignmentDefinitions), []byte(assDefId)); err != nil {
		t.Fatalf("error checking if assignment def exists: %v", err)
	} else if exists {
		t.Fatal("assignment def exists although it was deleted")
	}
}

func TestDeleteAppeal(t *testing.T) {
	defer db.InitDbForTest()()
	box := messages.NewMessageBox()
	msg := &messages.Message{}
	msg.ID = commons.GenerateUniqueId()
	box.Messages.Add(msg.ID)
	appeal := &appeals.Appeal{AssignmentInstance: commons.GenerateUniqueId(), MessageBox: box.ID}
	if err := db.Update(db.System, box, msg, appeal); err != nil {
		t.Fatalf("error creating box and msg for test: %v", err)
	}
	if err := appeals.Delete(appeal); err != nil {
		t.Fatalf("error deleting appeal: %v", err)
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Messages), []byte(msg.ID)); err != nil {
		t.Fatalf("error checking if message exists: %v", err)
	} else if exists {
		t.Fatal("message exists although box was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.MessageBoxes), []byte(box.ID)); err != nil {
		t.Fatalf("error checking if message box exists: %v", err)
	} else if exists {
		t.Fatal("message box exists although appeal was deleted")
	}
	if exists, err := db.KeyExistsInBucket([]byte(db.Appeals), []byte(appeal.AssignmentInstance)); err != nil {
		t.Fatalf("error checking if appeal exists: %v", err)
	} else if exists {
		t.Fatal("appeal exists although it was deleted")
	}
}
