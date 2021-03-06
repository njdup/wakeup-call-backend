package group

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/njdup/wakeup-call-backend/db"
	"github.com/njdup/wakeup-call-backend/models/user"
	"github.com/njdup/wakeup-call-backend/utils/errors"
)

var (
	CollectionName = "groups"
)

type Group struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Name        string        `bson:"groupName" json:"groupName"`
	Created     time.Time     `bson:"created" json:"-"`
	Phonenumber string        `bson:"phoneNumber" json:"phoneNumber"`

	Users []bson.ObjectId `bson:"users" json:"-"`
}

func (group *Group) Save() error {
	// Add validation checks here

	insertQuery := func(col *mgo.Collection) error {
		count, err := col.Find(bson.M{"groupName": group.Name}).Limit(1).Count()
		if err != nil {
			return err
		}
		if count != 0 {
			return &errorUtils.InvalidFieldsError{
				"A group with the given name already exists",
				[]string{"Name"},
			}
		}
		group.Created = time.Now() // Add creation time stamp
		return col.Insert(group)
	}

	return db.ExecWithCol(CollectionName, insertQuery)
}

// ProvisionPhoneNumber assigns a Twilio phone number for the group
// TODO: For now, returns trial number. Replace with creating # programatically
func (group *Group) ProvisionPhoneNumber() error {
	newNumber := "+18705251963"
	group.Phonenumber = newNumber

	// Check if the new number has already been used, in which case raise an error
	numberCheckQuery := func(col *mgo.Collection) error {
		count, err := col.Find(bson.M{"phoneNumber": newNumber}).Limit(1).Count()
		if err != nil {
			return err
		}
		if count != 0 {
			return &errorUtils.GeneralError{Message: "Error creating group phone number: number already exists"}
		}
		return nil
	}
	err := db.ExecWithCol(CollectionName, numberCheckQuery)
	return err
}

// AddUser adds the given user to the receiver group
// The database entries for both the group and user are updated with the
// appropriate references to one another
// Returns nil on success, encountered error on failure
// TODO: Group object will probably be out of date after this. Check that.
func (group *Group) AddUser(newUser *user.User) error {
	// First add user id to group's user array
	addUserQuery := func(col *mgo.Collection) error {
		groupSelector := bson.M{"groupName": group.Name}
		update := bson.M{"$push": bson.M{"users": newUser.Id}}
		return col.Update(groupSelector, update)
	}
	err := db.ExecWithCol(CollectionName, addUserQuery)
	if err != nil {
		return err
	}

	// Next, add group id to user's group array
	addGroupQuery := func(col *mgo.Collection) error {
		userSelector := bson.M{"userName": newUser.Username}
		update := bson.M{"$push": bson.M{"groups": group.Id}}
		return col.Update(userSelector, update)
	}
	return db.ExecWithCol(user.CollectionName, addGroupQuery)
}

// Users returns an array of all members of the receiver group
func (group *Group) GetUsers() ([]user.User, error) {
	groupUsers := []user.User{}
	searchQuery := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": bson.M{"$in": group.Users}}).All(&groupUsers)
	}
	err := db.ExecWithCol(user.CollectionName, searchQuery)
	if err != nil {
		return nil, err
	}
	return groupUsers, nil
}

func FindMatchingGroup(groupName string) (*Group, error) {
	result := Group{}
	searchQuery := func(col *mgo.Collection) error {
		return col.Find(bson.M{"groupName": groupName}).One(&result)
	}
	err := db.ExecWithCol(CollectionName, searchQuery)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func FindGroupWithNumber(phoneNumber string) (*Group, error) {
	result := Group{}
	searchQuery := func(col *mgo.Collection) error {
		return col.Find(bson.M{"phoneNumber": phoneNumber}).One(&result)
	}
	err := db.ExecWithCol(CollectionName, searchQuery)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func GetGroupsForUser(user *user.User) ([]Group, error) {
	userGroups := []Group{}
	searchQuery := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": bson.M{"$in": user.Groups}}).All(&userGroups)
	}
	err := db.ExecWithCol(CollectionName, searchQuery)
	if err != nil {
		return nil, err
	}
	return userGroups, nil
}
