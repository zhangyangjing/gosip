package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/koltyakov/gosip"
)

// Item represents SharePoint Lists & Document Libraries Items API queryable object struct
// Always use NewItem constructor instead of &Item{}
type Item struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers *ODataMods
}

// GenericItemInfo - list's generic item response payload structure
type GenericItemInfo struct {
	ID            int       `json:"Id"`
	Title         string    `json:"Title"`
	ContentTypeID string    `json:"ContentTypeId"`
	Attachments   bool      `json:"Attachments"`
	AuthorID      int       `json:"AuthorId"`
	EditorID      int       `json:"EditorId"`
	Created       time.Time `json:"Created"`
	Modified      time.Time `json:"Modified"`
}

// ItemResp - item response type with helper processor methods
type ItemResp []byte

// NewItem - Item struct constructor function
func NewItem(client *gosip.SPClient, endpoint string, config *RequestConfig) *Item {
	return &Item{
		client:    client,
		endpoint:  endpoint,
		config:    config,
		modifiers: NewODataMods(),
	}
}

// ToURL gets endpoint with modificators raw URL
func (item *Item) ToURL() string {
	return toURL(item.endpoint, item.modifiers)
}

// Conf receives custom request config definition, e.g. custom headers, custom OData mod
func (item *Item) Conf(config *RequestConfig) *Item {
	item.config = config
	return item
}

// Select ...
func (item *Item) Select(oDataSelect string) *Item {
	item.modifiers.AddSelect(oDataSelect)
	return item
}

// Expand ...
func (item *Item) Expand(oDataExpand string) *Item {
	item.modifiers.AddExpand(oDataExpand)
	return item
}

// Get ...
func (item *Item) Get() (ItemResp, error) {
	sp := NewHTTPClient(item.client)
	return sp.Get(item.ToURL(), getConfHeaders(item.config))
}

// Delete ...
func (item *Item) Delete() ([]byte, error) {
	sp := NewHTTPClient(item.client)
	return sp.Delete(item.endpoint, getConfHeaders(item.config))
}

// Update ...
func (item *Item) Update(body []byte) ([]byte, error) {
	body = patchMetadataTypeCB(body, func() string {
		endpoint := getPriorEndpoint(item.endpoint, "/Items")
		list := NewList(item.client, endpoint, nil)
		oDataType, _ := list.GetEntityType()
		return oDataType
	})
	sp := NewHTTPClient(item.client)
	return sp.Update(item.endpoint, body, getConfHeaders(item.config))
}

// Recycle ...
func (item *Item) Recycle() ([]byte, error) {
	endpoint := fmt.Sprintf("%s/Recycle", item.endpoint)
	sp := NewHTTPClient(item.client)
	return sp.Post(endpoint, nil, getConfHeaders(item.config))
}

// Roles ...
func (item *Item) Roles() *Roles {
	return NewRoles(item.client, item.endpoint, item.config)
}

// Attachments ...
func (item *Item) Attachments() *Attachments {
	return NewAttachments(
		item.client,
		fmt.Sprintf("%s/AttachmentFiles", item.endpoint),
		item.config,
	)
}

// ParentList ...
func (item *Item) ParentList() *List {
	return NewList(
		item.client,
		fmt.Sprintf("%s/ParentList", item.endpoint),
		item.config,
	)
}

// Records ...
func (item *Item) Records() *Records {
	return NewRecords(item)
}

// ContextInfo ...
func (item *Item) ContextInfo() (*ContextInfo, error) {
	return NewContext(item.client, item.ToURL(), item.config).Get()
}

/* Response helpers */

// Data : to get typed data
func (itemResp *ItemResp) Data() *GenericItemInfo {
	data := parseODataItem(*itemResp)
	data = fixDatesInResponse(data, []string{"Created", "Modified"})
	res := &GenericItemInfo{}
	json.Unmarshal(data, &res)
	return res
}

// Unmarshal : to unmarshal to custom object
func (itemResp *ItemResp) Unmarshal(obj interface{}) error {
	data := parseODataItem(*itemResp)
	data = normalizeMultiLookups(data)
	return json.Unmarshal(data, obj)
}
