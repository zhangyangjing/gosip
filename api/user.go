package api

import (
	"encoding/json"
	"fmt"

	"github.com/koltyakov/gosip"
)

// User represents SharePoint Site User API queryable object struct
// Always use NewUser constructor instead of &User{}
type User struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers *ODataMods
}

// UserInfo - site user API response payload structure
type UserInfo struct {
	Email         string `json:"Email"`
	ID            int    `json:"Id"`
	IsHiddenInUI  bool   `json:"IsHiddenInUI"`
	IsSiteAdmin   bool   `json:"IsSiteAdmin"`
	LoginName     string `json:"LoginName"`
	PrincipalType int    `json:"PrincipalType"`
	Title         string `json:"Title"`
}

// UserResp - site user response type with helper processor methods
type UserResp []byte

// NewUser - User struct constructor function
func NewUser(client *gosip.SPClient, endpoint string, config *RequestConfig) *User {
	return &User{
		client:    client,
		endpoint:  endpoint,
		config:    config,
		modifiers: NewODataMods(),
	}
}

// ToURL gets endpoint with modificators raw URL
func (user *User) ToURL() string {
	return toURL(user.endpoint, user.modifiers)
}

// Conf receives custom request config definition, e.g. custom headers, custom OData mod
func (user *User) Conf(config *RequestConfig) *User {
	user.config = config
	return user
}

// Select ...
func (user *User) Select(oDataSelect string) *User {
	user.modifiers.AddSelect(oDataSelect)
	return user
}

// Expand ...
func (user *User) Expand(oDataExpand string) *User {
	user.modifiers.AddExpand(oDataExpand)
	return user
}

// Get ...
func (user *User) Get() (UserResp, error) {
	sp := NewHTTPClient(user.client)
	return sp.Get(user.ToURL(), getConfHeaders(user.config))
}

// Groups ...
func (user *User) Groups() *Groups {
	return NewGroups(
		user.client,
		fmt.Sprintf("%s/Groups", user.endpoint),
		user.config,
	)
}

/* Response helpers */

// Data : to get typed data
func (userResp *UserResp) Data() *UserInfo {
	data := parseODataItem(*userResp)
	res := &UserInfo{}
	json.Unmarshal(data, &res)
	return res
}

// Unmarshal : to unmarshal to custom object
func (userResp *UserResp) Unmarshal(obj interface{}) error {
	data := parseODataItem(*userResp)
	return json.Unmarshal(data, obj)
}
