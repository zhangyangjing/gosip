package api

import (
	"encoding/json"
	"fmt"

	"github.com/koltyakov/gosip"
)

// ContentTypes represent SharePoint Content Types API queryable collection struct
// Always use NewContentTypes constructor instead of &ContentTypes{}
type ContentTypes struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers *ODataMods
}

// ContentTypesResp - content types response type with helper processor methods
type ContentTypesResp []byte

// NewContentTypes - ContentTypes struct constructor function
func NewContentTypes(client *gosip.SPClient, endpoint string, config *RequestConfig) *ContentTypes {
	return &ContentTypes{
		client:    client,
		endpoint:  endpoint,
		config:    config,
		modifiers: NewODataMods(),
	}
}

// ToURL gets endpoint with modificators raw URL
func (contentTypes *ContentTypes) ToURL() string {
	return toURL(contentTypes.endpoint, contentTypes.modifiers)
}

// Conf receives custom request config definition, e.g. custom headers, custom OData mod
func (contentTypes *ContentTypes) Conf(config *RequestConfig) *ContentTypes {
	contentTypes.config = config
	return contentTypes
}

// Select ...
func (contentTypes *ContentTypes) Select(oDataSelect string) *ContentTypes {
	contentTypes.modifiers.AddSelect(oDataSelect)
	return contentTypes
}

// Expand ...
func (contentTypes *ContentTypes) Expand(oDataExpand string) *ContentTypes {
	contentTypes.modifiers.AddExpand(oDataExpand)
	return contentTypes
}

// Filter ...
func (contentTypes *ContentTypes) Filter(oDataFilter string) *ContentTypes {
	contentTypes.modifiers.AddFilter(oDataFilter)
	return contentTypes
}

// Top ...
func (contentTypes *ContentTypes) Top(oDataTop int) *ContentTypes {
	contentTypes.modifiers.AddTop(oDataTop)
	return contentTypes
}

// OrderBy ...
func (contentTypes *ContentTypes) OrderBy(oDataOrderBy string, ascending bool) *ContentTypes {
	contentTypes.modifiers.AddOrderBy(oDataOrderBy, ascending)
	return contentTypes
}

// Get ...
func (contentTypes *ContentTypes) Get() (ContentTypesResp, error) {
	sp := NewHTTPClient(contentTypes.client)
	return sp.Get(contentTypes.ToURL(), getConfHeaders(contentTypes.config))
}

// GetByID ...
func (contentTypes *ContentTypes) GetByID(contentTypeID string) *ContentType {
	return NewContentType(
		contentTypes.client,
		fmt.Sprintf("%s('%s')", contentTypes.endpoint, contentTypeID),
		contentTypes.config,
	)
}

// ToDo:
// Add

/* Response helpers */

// Data : to get typed data
func (ctsResp *ContentTypesResp) Data() []ContentTypeResp {
	collection, _ := parseODataCollection(*ctsResp)
	cts := []ContentTypeResp{}
	for _, ct := range collection {
		cts = append(cts, ContentTypeResp(ct))
	}
	return cts
}

// Unmarshal : to unmarshal to custom object
func (ctsResp *ContentTypesResp) Unmarshal(obj interface{}) error {
	// collection := parseODataCollection(*ctsResp)
	// data, _ := json.Marshal(collection)
	data, _ := parseODataCollectionPlain(*ctsResp)
	return json.Unmarshal(data, obj)
}
