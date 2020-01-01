package api

import (
	"encoding/json"
	"fmt"

	"github.com/koltyakov/gosip"
)

// Files represent SharePoint Files API queryable collection struct
// Always use NewFiles constructor instead of &Files{}
type Files struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers *ODataMods
}

// FilesResp - files response type with helper processor methods
type FilesResp []byte

// NewFiles - Files struct constructor function
func NewFiles(client *gosip.SPClient, endpoint string, config *RequestConfig) *Files {
	return &Files{
		client:    client,
		endpoint:  endpoint,
		config:    config,
		modifiers: NewODataMods(),
	}
}

// ToURL gets endpoint with modificators raw URL
func (files *Files) ToURL() string {
	return toURL(files.endpoint, files.modifiers)
}

// Conf receives custom request config definition, e.g. custom headers, custom OData mod
func (files *Files) Conf(config *RequestConfig) *Files {
	files.config = config
	return files
}

// Select ...
func (files *Files) Select(oDataSelect string) *Files {
	files.modifiers.AddSelect(oDataSelect)
	return files
}

// Expand ...
func (files *Files) Expand(oDataExpand string) *Files {
	files.modifiers.AddExpand(oDataExpand)
	return files
}

// Filter ...
func (files *Files) Filter(oDataFilter string) *Files {
	files.modifiers.AddFilter(oDataFilter)
	return files
}

// Top ...
func (files *Files) Top(oDataTop int) *Files {
	files.modifiers.AddTop(oDataTop)
	return files
}

// OrderBy ...
func (files *Files) OrderBy(oDataOrderBy string, ascending bool) *Files {
	files.modifiers.AddOrderBy(oDataOrderBy, ascending)
	return files
}

// Get ...
func (files *Files) Get() (FilesResp, error) {
	sp := NewHTTPClient(files.client)
	return sp.Get(files.ToURL(), getConfHeaders(files.config))
}

// GetByName ...
func (files *Files) GetByName(fileName string) *File {
	return NewFile(
		files.client,
		fmt.Sprintf("%s('%s')", files.endpoint, fileName),
		files.config,
	)
}

// Add ...
func (files *Files) Add(name string, content []byte, overwrite bool) (FileResp, error) {
	sp := NewHTTPClient(files.client)
	endpoint := fmt.Sprintf("%s/Add(overwrite=%t,url='%s')", files.endpoint, overwrite, name)
	return sp.Post(endpoint, content, getConfHeaders(files.config))
}

/* Response helpers */

// Data : to get typed data
func (filesResp *FilesResp) Data() []FileResp {
	collection, _ := parseODataCollection(*filesResp)
	files := []FileResp{}
	for _, ct := range collection {
		files = append(files, FileResp(ct))
	}
	return files
}

// Unmarshal : to unmarshal to custom object
func (filesResp *FilesResp) Unmarshal(obj interface{}) error {
	data, _ := parseODataCollectionPlain(*filesResp)
	return json.Unmarshal(data, obj)
}
