// TODO: Clean up this file with new tricks like your new response utils

package userController

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/njdup/wakeup-call-backend/models/group"
	"github.com/njdup/wakeup-call-backend/models/user"
	"github.com/njdup/wakeup-call-backend/utils/errors"
	"github.com/njdup/wakeup-call-backend/utils/responses"
)

func AllUsers(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		return
	})
}

// GetUser returns information for a currently signed in user
func GetUserInfo(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// First ensure that a user is logged in
		session, _ := sessionStore.Get(req, "wakeup-session")
		username, authenticated := session.Values["user"]
		if !authenticated {
			errorMsg := &errorUtils.GeneralError{Message: "You must be signed in to get your user info"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}

		usernameStr := fmt.Sprintf("%s", username)
		user, err := user.FindMatchingUser(usernameStr)
		if err != nil {
			errorMsg := &errorUtils.GeneralError{Message: "Unable to retrieve user information"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusInternalServerError, res)
			return
		}
		APIResponses.SendSuccessResponse(user, res)
		return
	})
}

func CreateUser(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Prepare new user from form data
		req.ParseForm()
		newUser := &user.User{
			Username:  req.FormValue("Username"),
			Firstname: req.FormValue("Firstname"),
			Lastname:  req.FormValue("Lastname"),
		}

		// TODO: Add parsePhonenumber to user model
		phonenumber, err := parsePhonenumber(req.FormValue("Phonenumber"))
		if err != nil {
			errorMsg := &errorUtils.InvalidFieldsError{
				Message: "Given phone number is invalid",
				Fields:  []string{"Phonenumber"},
			}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}
		newUser.Phonenumber = phonenumber

		err = newUser.HashPassword(req.FormValue("Password"))
		if err != nil {
			errorMsg := &errorUtils.InvalidFieldsError{
				Message: "The given password is invalid",
				Fields:  []string{"Password"},
			}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}

		// Now attempt to save, create appropriate response
		resContent := &APIResponses.Response{}
		err = newUser.Save()
		if err != nil {
			resContent.Status = 400
			resContent.Error = err
		} else {
			resContent.Status = 200
			resContent.Data = "success"
		}

		payload, err := json.MarshalIndent(resContent, "", "  ")
		if err != nil {
			fmt.Fprintf(res, `{"Status": 500, "Error": "Unable to prepare server response"}`)
			return
		}
		fmt.Fprintf(res, string(payload))
		return
	})
}

func Login(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// First check if a session is already active
		session, _ := sessionStore.Get(req, "wakeup-session")
		if _, ok := session.Values["user"]; ok {
			http.Error(res, "User is already signed in", http.StatusBadRequest)
			return
		}

		// Otherwise, try authenticating
		req.ParseForm()
		matchedUser, err := user.FindMatchingUser(req.FormValue("Username"))
		if err != nil {
			fmt.Fprintf(res, err.Error())
			return
		}

		if matchedUser.ConfirmPassword(req.FormValue("Password")) {
			session.Values["user"] = matchedUser.Username
			session.Save(req, res)

			resContent := &APIResponses.Response{Status: 200, Data: "Successfully signed in"}
			payload, _ := json.MarshalIndent(resContent, "", "  ")
			fmt.Fprintf(res, string(payload))
		} else {
			errorMsg := &errorUtils.InvalidFieldsError{
				Message: "The given password is incorrect",
				Fields:  []string{"Password"},
			}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}
		return
	})
}

func Logout(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		session, _ := sessionStore.Get(req, "wakeup-session")
		if _, ok := session.Values["user"]; !ok {
			http.Error(res, "No user present to sign out", http.StatusBadRequest)
			return
		}
		delete(session.Values, "user")
		session.Save(req, res)

		resContent := &APIResponses.Response{Status: 200, Data: "Successfully logged out"}
		payload, _ := json.MarshalIndent(resContent, "", "  ")
		fmt.Fprintf(res, string(payload))
		return
	})
}

// CheckSession determines whether the given request's cookie is valid
func CheckSession(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		session, _ := sessionStore.Get(req, "wakeup-session")
		if _, ok := session.Values["user"]; ok {
			APIResponses.SendSuccessResponse("Valid cookie found", res)
			return
		}
		errorMsg := &errorUtils.GeneralError{Message: "Invalid cookie given"}
		APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
		return
	})
}

func GetUserGroups(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		session, _ := sessionStore.Get(req, "wakeup-session")
		if _, ok := session.Values["user"]; !ok {
			errorMsg := &errorUtils.GeneralError{Message: "You must be signed in to retrieve a user's groups"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}

		vars := mux.Vars(req)
		username := vars["username"]
		user, err := user.FindMatchingUser(username)
		if err != nil {
			errorMsg := &errorUtils.GeneralError{Message: "A user matching the given username was not found"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}

		groups, err := group.GetGroupsForUser(user)
		if err != nil {
			errorMsg := &errorUtils.GeneralError{Message: "An error occurred retrieving the user's groups"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusInternalServerError, res)
			return
		}

		APIResponses.SendSuccessResponse(groups, res)
		return
	})
}

func GetUser(sessionStore *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		queryValues := req.URL.Query()
		if len(queryValues["phoneNumber"]) == 0 {
			errorMsg := &errorUtils.GeneralError{Message: "A phone number must be included to query for a user"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusBadRequest, res)
			return
		}

		phoneNumber := queryValues["phoneNumber"][0]
		fmt.Println("Received query for user with phone number: " + phoneNumber)
		user, err := user.FindUserWithNumber(phoneNumber)
		// TODO: Improve this to detect the type of error. Will probably have to
		// expand functionality of FindUserWithNumber
		if err != nil {
			errorMsg := &errorUtils.GeneralError{Message: "Unable to find user with the given phone number"}
			APIResponses.SendErrorResponse(errorMsg, http.StatusInternalServerError, res)
			return
		}
		APIResponses.SendSuccessResponse(user, res)
		return
	})
}

// ConfigRoutes initializes all application routes specific to users
func ConfigRoutes(router *mux.Router, sessionStore *sessions.CookieStore) {
	router.Handle("/users", CreateUser(sessionStore)).Methods("POST")
	router.Handle("/users", GetUser(sessionStore)).Methods("GET")
	router.Handle("/users/login", Login(sessionStore)).Methods("POST")
	router.Handle("/users/logout", Logout(sessionStore)).Methods("POST")
	router.Handle("/users/info", GetUserInfo(sessionStore)).Methods("GET")
	router.Handle("/users/{username}/groups", GetUserGroups(sessionStore)).Methods("GET")

	router.Handle("/users/sessioncheck", CheckSession(sessionStore)).Methods("GET")
}

// TODO: Write this to format phone numbers given in form
func parsePhonenumber(inputNumber string) (string, error) {
	return inputNumber, nil
}
