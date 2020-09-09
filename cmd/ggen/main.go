package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type apiGenCnfg struct {
	Entity       string
	Item         string
	Configurable bool
	IsCollection bool
	Modificators []string
	Helpers      []string
}

func main() {
	ent := flag.String("ent", "", "Entity struct name")
	item := flag.String("item", "", "Child entity struct name")
	conf := flag.Bool("conf", false, "Has Conf() method")
	coll := flag.Bool("coll", false, "Is collection entity")
	mods := flag.String("mods", "", "Modifiers comma separated list")
	helpers := flag.String("helpers", "", "Helpers comma separated list")
	flag.Parse()

	if *ent == "" {
		fmt.Printf("can't generate %+v as no entity is provided, skipping...\n", os.Args)
		return
	}

	var m []string
	if len(*mods) > 0 {
		m = strings.Split(*mods, ",")
	}

	var h []string
	if len(*helpers) > 0 {
		h = strings.Split(*helpers, ",")
	}

	_ = generate(&apiGenCnfg{
		Entity:       *ent,
		Item:         *item,
		Configurable: *conf,
		IsCollection: *coll,
		Modificators: m,
		Helpers:      h,
	})
}

func generate(c *apiGenCnfg) error {
	pkgPath, _ := filepath.Abs("./")
	pkg := filepath.Base(pkgPath)

	instance := instanceOf(c.Entity)
	genFileName := fmt.Sprintf("%s_gen.go", instance)

	command := ""
	for _, a := range os.Args {
		command += " " + a
	}
	command = strings.TrimSpace(command)

	code := fmt.Sprintf("// Code generated by `" + command + "`; DO NOT EDIT.\n\n")
	code += fmt.Sprintf("package %s\n", pkg)

	imports := map[string]bool{}
	if c.IsCollection && len(c.Helpers) > 0 {
		for _, helper := range c.Helpers {
			if helper == "ToMap" {
				imports["encoding/json"] = true
			}
		}
	}
	if !c.IsCollection && len(c.Helpers) > 0 {
		for _, helper := range c.Helpers {
			if helper == "Data" {
				imports["encoding/json"] = true
			}
			if helper == "ToMap" {
				imports["encoding/json"] = true
			}
		}
	}
	if len(imports) > 0 {
		packages := ""
		for k := range imports {
			packages += fmt.Sprintf("\"%s\"\n", k)
		}
		code += `
			import (
				` + packages + `
			)
		`
	}

	if c.Configurable {
		code += `
			// Conf receives custom request config definition, e.g. custom headers, custom OData mod
			func (` + instance + ` *` + c.Entity + `) Conf(config *RequestConfig) *` + c.Entity + ` {
				` + instance + `.config = config
				return ` + instance + `
			}
		`
	}

	if len(c.Modificators) > 0 {
		code += modificatorsGen(c)
	}

	if len(c.Helpers) > 0 {
		code += helpersGen(c)
	}

	fmt.Printf("Generated %s (%d bytes)\n", filepath.Join("./", genFileName), len([]byte(code)))

	err := ioutil.WriteFile(filepath.Join("./", genFileName), []byte(code), 0644)
	return err
}

func modificatorsGen(c *apiGenCnfg) string {
	Ent := c.Entity
	ent := instanceOf(c.Entity)
	code := ""
	for _, mod := range c.Modificators {
		switch mod {
		case "Select":
			code += `
				// Select adds $select OData modifier
				func (` + ent + ` *` + Ent + `) Select(oDataSelect string) *` + Ent + ` {
					` + ent + `.modifiers.AddSelect(oDataSelect)
					return ` + ent + `
				}
			`
		case "Expand":
			code += `
				// Expand adds $expand OData modifier
				func (` + ent + ` *` + Ent + `) Expand(oDataExpand string) *` + Ent + ` {
					` + ent + `.modifiers.AddExpand(oDataExpand)
					return ` + ent + `
				}
			`
		case "Filter":
			code += `
				// Filter adds $filter OData modifier
				func (` + ent + ` *` + Ent + `) Filter(oDataFilter string) *` + Ent + ` {
					` + ent + `.modifiers.AddFilter(oDataFilter)
					return ` + ent + `
				}
			`
		case "Top":
			code += `
				// Top adds $top OData modifier
				func (` + ent + ` *` + Ent + `) Top(oDataTop int) *` + Ent + ` {
					` + ent + `.modifiers.AddTop(oDataTop)
					return ` + ent + `
				}
			`
		case "Skip":
			code += `
				// Skip adds $skiptoken OData modifier
				func (` + ent + ` *` + Ent + `) Skip(skipToken string) *` + Ent + ` {
					` + ent + `.modifiers.AddSkip(skipToken)
					return ` + ent + `
				}
			`
		case "OrderBy":
			code += `
				// OrderBy adds $orderby OData modifier
				func (` + ent + ` *` + Ent + `) OrderBy(oDataOrderBy string, ascending bool) *` + Ent + ` {
					` + ent + `.modifiers.AddOrderBy(oDataOrderBy, ascending)
					return ` + ent + `
				}
			`
		}
	}
	return code
}

func helpersGen(c *apiGenCnfg) string {
	Ent := c.Entity
	ent := instanceOf(c.Entity)
	code := ""
	if c.IsCollection && c.Item == "" {
		return ""
	}
	if len(c.Helpers) > 0 {
		code += fmt.Sprintf("\n/* Response helpers */\n")
	}
	if c.IsCollection {
		for _, mod := range c.Helpers {
			switch mod {
			case "Data":
				code += `
					// Data response helper
					func (` + ent + `Resp *` + Ent + `Resp) Data() []` + c.Item + `Resp {
						collection, _ := normalizeODataCollection(*` + ent + `Resp)
						` + ent + ` := []` + c.Item + `Resp{}
						for _, item := range collection {
							` + ent + ` = append(` + ent + `, ` + c.Item + `Resp(item))
						}
						return ` + ent + `
					}
				`
			case "Normalized":
				code += `
					// Normalized returns normalized body
					func (` + ent + `Resp *` + Ent + `Resp) Normalized() []byte {
						normalized, _ := NormalizeODataCollection(*` + ent + `Resp)
						return normalized
					}
				`
			case "ToMap":
				code += `
					// ToMap unmarshals response to generic map
					func (` + ent + `Resp *` + Ent + `Resp) ToMap() []map[string]interface{} {
						data, _ := NormalizeODataCollection(*` + ent + `Resp)
						var res []map[string]interface{}
						_ = json.Unmarshal(data, &res)
						return res
					}
				`
			}
		}
	}
	if !c.IsCollection {
		for _, mod := range c.Helpers {
			switch mod {
			case "Data":
				code += `
					// Data response helper
					func (` + ent + `Resp *` + Ent + `Resp) Data() *` + Ent + `Info {
						data := NormalizeODataItem(*` + ent + `Resp)
						res := &` + Ent + `Info{}
						json.Unmarshal(data, &res)
						return res
					}
				`
			case "Normalized":
				code += `
					// Normalized returns normalized body
					func (` + ent + `Resp *` + Ent + `Resp) Normalized() []byte {
						return NormalizeODataItem(*` + ent + `Resp)
					}
				`
			case "ToMap":
				code += `
					// ToMap unmarshals response to generic map
					func (` + ent + `Resp *` + Ent + `Resp) ToMap() map[string]interface{} {
						data := NormalizeODataItem(*` + ent + `Resp)
						var res map[string]interface{}
						_ = json.Unmarshal(data, &res)
						return res
					}
				`
			}
		}
	}
	return code
}

// func paginationGen(entity string) string {
// 	Ent := entity
// 	ent := instanceOf(Ent)
// 	return `
// 		/* Pagination helpers */

// 		// ` + Ent + `Page - paged items
// 		type ` + Ent + `Page struct {
// 			Items       ` + Ent + `Resp
// 			HasNextPage func() bool
// 			GetNextPage func() (*` + Ent + `Page, error)
// 		}

// 		// GetPaged gets Paged Items collection
// 		func (` + ent + ` *` + Ent + `) GetPaged() (*` + Ent + `Page, error) {
// 			data, err := ` + ent + `.Get()
// 			if err != nil {
// 				return nil, err
// 			}
// 			res := &` + Ent + `Page{
// 				Items: data,
// 				HasNextPage: func() bool {
// 					return data.HasNextPage()
// 				},
// 				GetNextPage: func() (*` + Ent + `Page, error) {
// 					nextURL := data.NextPageURL()
// 					if nextURL == "" {
// 						return nil, fmt.Errorf("unable to get next page")
// 					}
// 					return New` + Ent + `(` + ent + `.client, nextURL, ` + ent + `.config).GetPaged()
// 				},
// 			}
// 			return res, nil
// 		}

// 		// NextPageURL gets next page OData collection
// 		func (` + ent + `Resp *` + Ent + `Resp) NextPageURL() string {
// 			return getODataCollectionNextPageURL(*` + ent + `Resp)
// 		}

// 		// HasNextPage returns is true if next page exists
// 		func (` + ent + `Resp *` + Ent + `Resp) HasNextPage() bool {
// 			return ` + ent + `Resp.NextPageURL() != ""
// 		}
// 	`
// }

func instanceOf(entity string) string {
	if len(entity) < 4 {
		return strings.ToLower(entity)
	}
	ent := ""
	for i, l := range entity {
		pos := string(l)
		if i == 0 {
			pos = strings.ToLower(pos)
		}
		ent += pos
	}
	return ent
}
