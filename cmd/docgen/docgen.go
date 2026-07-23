package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	//"regexp"
	//"sync"
)

var header = `---
title: Module %s
description: %s
layout: doc
menu:
  docs:
    parent: "%s"
---

`

type emmyPiece struct {
	DocPiece    *docPiece
	Annotations []string
	Params      []string // we only need to know param name to put in function
	FuncName    string
}

type module struct {
	Name             string     `json:"name"`
	Section          string     `json:"section,omitempty"` // "api" or "nature", picks the docs/ subdir
	ShortDescription string     `json:"shortDescription"`
	Description      string     `json:"description"`
	ParentModule     string     `json:"parent,omitempty"`
	Properties       []docPiece `json:"properties"`
	Fields           []docPiece `json:"fields"`
	Types            []docPiece `json:"types,omitempty"`
	Docs             []docPiece `json:"docs"`
	Since            string     `json:"since,omitempty"`
}

type param struct {
	Name       string
	Type       string
	Doc        []string
	Default    string
	IsOptional bool
	Keys       []param
}

type docPiece struct {
	FuncName     string           `json:"name"`
	Doc          []string         `json:"description"`
	ParentModule string           `json:"parent,omitempty"`
	Interfacing  string           `json:"interfaces,omitempty"`
	FuncSig      string           `json:"signature,omitempty"`
	GoFuncName   string           `json:"goFuncName,omitempty"`
	IsInterface  bool             `json:"isInterface"`
	IsMember     bool             `json:"isMember"`
	IsType       bool             `json:"isType"`
	Fields       []docPiece       `json:"fields,omitempty"`
	Properties   []docPiece       `json:"properties,omitempty"`
	Params       []param          `json:"params,omitempty"`
	Returns      []param          `json:"returns,omitempty"`
	Tags         map[string][]tag `json:"tags,omitempty"`
	Type         string           `json:"type,omitempty"`
	Since        string           `json:"since,omitempty"`
	SeeAlso      []string         `json:"seeAlso,omitempty"`
	Notes        []string         `json:"notes,omitempty"`
}

type tag struct {
	Id       string   `json:"id"`
	Fields   []string `json:"fields"`
	StartIdx int      `json:"startIdx"`
}

var docs = make(map[string]module)
var emmyDocs = make(map[string][]emmyPiece)
var typeTable = make(map[string][]string) // [0] = parentMod, [1] = interfaces
var prefix = map[string]string{
	"main":      "hl",
	"hilbish":   "hl",
	"fs":        "f",
	"commander": "c",
	"bait":      "b",
	"terminal":  "term",
	"snail":     "snail",
	"readline":  "rl",
	"yarn":      "yarn",
}

var allowedTags = map[string]bool{
	"param": true, "return": true, "private": true, "field": true, "property": true,
	"type": true, "interface": true, "member": true, "example": true, "default": true,
	"tparam": true, "since": true, "see": true, "note": true, "treturn": true,
	// hook-specific tags
	"hook": true, "group": true, "var": true,
}

// getTagsAndDocs parses @-tagged doc-comment lines out of docs. context is a
// human-readable location (e.g. "hilbish.foo") used only for the unknown-tag
// warning, so typos and stray legacy tags don't silently vanish.
func getTagsAndDocs(docs string, context string) (map[string][]tag, []string) {
	pts := strings.Split(docs, "\n")
	parts := []string{}
	tags := make(map[string][]tag)

	for idx, part := range pts {
		if strings.HasPrefix(part, "@") {
			tagParts := strings.Split(strings.TrimPrefix(part, "@"), " ")
			tagName := tagParts[0]

			if !allowedTags[tagName] {
				fmt.Printf("WARNING! Unknown tag '@%s' in %s\n", tagName, context)
			}

			if tags[tagName] == nil {
				var id string
				if len(tagParts) > 1 {
					id = tagParts[1]
				}
				tags[tagName] = []tag{
					{Id: id, StartIdx: idx},
				}
				if len(tagParts) >= 2 {
					tags[tagName][0].Fields = tagParts[2:]
				}
			} else {
				if tagName == "example" {
					exampleIdx := tags["example"][0].StartIdx
					exampleCode := pts[exampleIdx+1 : idx]

					tags["example"][0].Fields = exampleCode
					parts = strings.Split(strings.Replace(strings.Join(parts, "\n"), strings.TrimPrefix(strings.Join(exampleCode, "\n"), "@example\n"), "", -1), "\n")
					continue
				}

				fleds := []string{}
				if len(tagParts) >= 2 {
					fleds = tagParts[2:]
				}
				tags[tagName] = append(tags[tagName], tag{
					Id:     tagParts[1],
					Fields: fleds,
				})
			}
		} else {
			parts = append(parts, part)
		}
	}

	return tags, parts
}

func docPieceTag(tagName string, tags map[string][]tag) []docPiece {
	dps := []docPiece{}
	for _, tag := range tags[tagName] {
		dps = append(dps, docPiece{
			FuncName: tag.Id,
			Doc:      tag.Fields,
		})
	}

	return dps
}

// looksLikeType returns true when s is a type token rather than the first word
// of a description sentence. Used to decide if @field name <token> ... has a
// type as the second word or not.
func looksLikeType(s string) bool {
	if s == "" {
		return false
	}
	switch s {
	case "string", "number", "boolean", "bool", "table", "function", "nil", "any", "integer", "int":
		return true
	}
	r := rune(s[0])
	return (r >= 'A' && r <= 'Z') || s[0] == '@' || strings.Contains(s, "<") || strings.Contains(s, "|")
}

// docPieceTagTyped is like docPieceTag but for @field/@property tags that
// optionally carry a type as the second token: @field name [type] description...
func docPieceTagTyped(tagName string, tags map[string][]tag) []docPiece {
	dps := []docPiece{}
	for _, t := range tags[tagName] {
		dp := docPiece{FuncName: t.Id}
		if len(t.Fields) > 0 && looksLikeType(t.Fields[0]) {
			dp.Type = t.Fields[0]
			dp.Doc = t.Fields[1:]
		} else {
			dp.Doc = t.Fields
		}
		dps = append(dps, dp)
	}
	return dps
}

func setupDocType(mod string, typ *doc.Type) *docPiece {
	docs := strings.TrimSpace(typ.Doc)
	tags, doc := getTagsAndDocs(docs, mod+"."+typ.Name)

	if tags["type"] == nil || tags["private"] != nil {
		return nil
	}
	inInterface := tags["interface"] != nil

	var interfaces string
	typeName := strings.ToUpper(string(typ.Name[0])) + typ.Name[1:]
	typeDoc := []string{}

	if inInterface {
		interfaces = tags["interface"][0].Id
	}

	fields := docPieceTagTyped("field", tags)
	properties := docPieceTagTyped("property", tags)

	for _, d := range doc {
		if !strings.HasPrefix(d, "---") {
			typeDoc = append(typeDoc, d)
		}
	}

	var isMember bool
	if tags["member"] != nil {
		isMember = true
	}

	var since string
	if t := tags["since"]; len(t) > 0 {
		since = t[0].Id
		if since == "" && len(t[0].Fields) > 0 {
			since = strings.Join(t[0].Fields, " ")
		}
	}
	var seeAlso []string
	for _, t := range tags["see"] {
		ref := strings.TrimSpace(t.Id + " " + strings.Join(t.Fields, " "))
		if ref != "" {
			seeAlso = append(seeAlso, ref)
		}
	}
	var notes []string
	for _, t := range tags["note"] {
		note := strings.TrimSpace(t.Id + " " + strings.Join(t.Fields, " "))
		if note != "" {
			notes = append(notes, note)
		}
	}

	parentMod := mod
	dps := &docPiece{
		Doc:          typeDoc,
		FuncName:     typeName,
		Interfacing:  interfaces,
		IsInterface:  inInterface,
		IsMember:     isMember,
		IsType:       true,
		ParentModule: parentMod,
		Fields:       fields,
		Properties:   properties,
		Tags:         tags,
		Since:        since,
		SeeAlso:      seeAlso,
		Notes:        notes,
	}

	typeTable[strings.ToLower(typeName)] = []string{parentMod, interfaces}

	return dps
}

func setupDoc(mod string, fun *doc.Func) *docPiece {
	if fun.Doc == "" {
		return nil
	}

	docs := strings.TrimSpace(fun.Doc)
	tags, parts := getTagsAndDocs(docs, mod+"."+fun.Name)

	if tags["private"] != nil {
		return nil
	}

	// i couldnt fit this into the condition below for some reason so here's a goto!
	if tags["member"] != nil {
		goto start
	}

	if prefix[mod] == "" {
		return nil
	}

	if (!strings.HasPrefix(fun.Name, prefix[mod]) && tags["interface"] == nil) || (strings.ToLower(fun.Name) == "loader" && tags["interface"] == nil) {
		return nil
	}

start:
	inInterface := tags["interface"] != nil
	var interfaces string
	funcName := strings.TrimPrefix(fun.Name, prefix[mod])
	doc := parts
	funcdoc := []string{}

	if inInterface {
		interfaces = tags["interface"][0].Id
		// @interface functions: Lua name comes from the first comment line (e.g. "start()")
		// because Go names like "luaStartJob" don't follow the prefix convention
		if len(parts) > 0 {
			rawName := strings.Split(parts[0], "(")[0]
			funcName = interfaces + "." + strings.TrimSpace(rawName)
			doc = parts[1:]
		}
	}
	em := emmyPiece{FuncName: funcName}

	fields := docPieceTagTyped("field", tags)
	properties := docPieceTagTyped("property", tags)
	var params []param
	defaultValues := make(map[string]string)
	if defaultsRaw := tags["default"]; defaultsRaw != nil {
		for _, d := range defaultsRaw {
			defaultValues[d.Id] = strings.Join(d.Fields, " ")
		}
	}
	// tparamKeys: map from parent param name to list of table key params
	tparamKeys := make(map[string][]param)
	if tparamsRaw := tags["tparam"]; tparamsRaw != nil {
		for _, tp := range tparamsRaw {
			// @tparam <parentName> <keyName> <type> <doc...>
			parent := tp.Id
			keyName := ""
			keyType := ""
			var keyDoc []string
			if len(tp.Fields) > 0 {
				keyName = tp.Fields[0]
			}
			if len(tp.Fields) > 1 {
				keyType = tp.Fields[1]
			}
			if len(tp.Fields) > 2 {
				keyDoc = []string{strings.Join(tp.Fields[2:], " ")}
			}
			isOptional := strings.HasSuffix(keyName, "?")
			if isOptional {
				keyName = strings.TrimSuffix(keyName, "?")
			}
			tparamKeys[parent] = append(tparamKeys[parent], param{
				Name:       keyName,
				Type:       keyType,
				Doc:        keyDoc,
				IsOptional: isOptional,
			})
		}
	}
	if paramsRaw := tags["param"]; paramsRaw != nil {
		params = make([]param, len(paramsRaw))
		for i, p := range paramsRaw {
			name := p.Id
			isOptional := strings.HasSuffix(name, "?")
			if isOptional {
				name = strings.TrimSuffix(name, "?")
			}
			params[i] = param{
				Name:       name,
				Type:       p.Fields[0],
				Doc:        p.Fields[1:],
				Default:    defaultValues[p.Id],
				IsOptional: isOptional,
				Keys:       tparamKeys[p.Id],
			}
		}
	}
	// treturnKeys: map from return name to list of table key params, same shape
	// as tparamKeys but keyed off @return's name instead of @param's.
	treturnKeys := make(map[string][]param)
	if treturnsRaw := tags["treturn"]; treturnsRaw != nil {
		for _, tp := range treturnsRaw {
			// @treturn <name> <keyName> <type> <doc...>
			parent := tp.Id
			keyName := ""
			keyType := ""
			var keyDoc []string
			if len(tp.Fields) > 0 {
				keyName = tp.Fields[0]
			}
			if len(tp.Fields) > 1 {
				keyType = tp.Fields[1]
			}
			if len(tp.Fields) > 2 {
				keyDoc = []string{strings.Join(tp.Fields[2:], " ")}
			}
			isOptional := strings.HasSuffix(keyName, "?")
			if isOptional {
				keyName = strings.TrimSuffix(keyName, "?")
			}
			treturnKeys[parent] = append(treturnKeys[parent], param{
				Name:       keyName,
				Type:       keyType,
				Doc:        keyDoc,
				IsOptional: isOptional,
			})
		}
	}
	var returns []param
	if returnsRaw := tags["return"]; returnsRaw != nil {
		returns = make([]param, len(returnsRaw))
		for i, r := range returnsRaw {
			// @return <type> <name> [description...]
			name := ""
			var doc []string
			if len(r.Fields) > 0 {
				name = r.Fields[0]
			}
			if len(r.Fields) > 1 {
				doc = r.Fields[1:]
			}
			isOptional := strings.HasSuffix(r.Id, "?")
			returns[i] = param{
				Name:       name,
				Type:       r.Id,
				Doc:        doc,
				IsOptional: isOptional,
				Keys:       treturnKeys[name],
			}
		}
	}

	for _, d := range doc {
		if strings.HasPrefix(d, "---") {
			emmyLine := strings.TrimSpace(strings.TrimPrefix(d, "---"))
			emmyLinePieces := strings.Split(emmyLine, " ")
			emmyType := emmyLinePieces[0]
			if emmyType == "@param" {
				em.Params = append(em.Params, emmyLinePieces[1])
			}
			if emmyType == "@vararg" {
				em.Params = append(em.Params, "...") // add vararg
			}
			em.Annotations = append(em.Annotations, d)
		} else {
			funcdoc = append(funcdoc, d)
		}
	}

	var isMember bool
	if tags["member"] != nil {
		isMember = true
	}
	var parentMod string
	if inInterface {
		parentMod = mod
	}

	// Build the Lua signature from param names + @return tags rather than
	// relying on a hand-written first-line comment.
	shortName := funcName
	if dot := strings.LastIndex(funcName, "."); dot >= 0 {
		shortName = funcName[dot+1:]
	}
	paramNames := make([]string, len(params))
	for i, p := range params {
		name := p.Name
		if strings.HasPrefix(p.Type, "...") {
			name = "..." + name
		}
		paramNames[i] = name
	}
	funcsig := shortName + "(" + strings.Join(paramNames, ", ") + ")"
	if len(returns) > 0 {
		retTypes := make([]string, len(returns))
		for i, r := range returns {
			retTypes[i] = r.Type
		}
		funcsig += " -> " + strings.Join(retTypes, ", ")
	}

	var since string
	if t := tags["since"]; len(t) > 0 {
		since = t[0].Id
		if since == "" && len(t[0].Fields) > 0 {
			since = strings.Join(t[0].Fields, " ")
		}
	}
	var seeAlso []string
	for _, t := range tags["see"] {
		ref := strings.TrimSpace(t.Id + " " + strings.Join(t.Fields, " "))
		if ref != "" {
			seeAlso = append(seeAlso, ref)
		}
	}
	var notes []string
	for _, t := range tags["note"] {
		note := strings.TrimSpace(t.Id + " " + strings.Join(t.Fields, " "))
		if note != "" {
			notes = append(notes, note)
		}
	}

	dps := &docPiece{
		Doc:          funcdoc,
		FuncSig:      funcsig,
		FuncName:     funcName,
		Interfacing:  interfaces,
		GoFuncName:   strings.ToLower(fun.Name),
		IsInterface:  inInterface,
		IsMember:     isMember,
		ParentModule: parentMod,
		Fields:       fields,
		Properties:   properties,
		Params:       params,
		Returns:      returns,
		Tags:         tags,
		Since:        since,
		SeeAlso:      seeAlso,
		Notes:        notes,
	}
	if strings.HasSuffix(dps.GoFuncName, strings.ToLower("loader")) {
		dps.Doc = parts
	}
	em.DocPiece = dps

	emmyDocs[mod] = append(emmyDocs[mod], em)
	return dps
}

func main() {
	// collect documentation from Go and Lua sources into defs.
	collectDefs()
	// render the defs into markdown docs.
	renderDocs()
	// render the defs into LuaLS type definition files.
	renderLuaDefs()
}

// collectDefs parses both the Go source (via go/doc) and the Lua source (via
// collectLuaModules) into a flat set of module defs, one JSON file per page,
// written to defs/<name>.json. Each def carries its Section ("api"/"nature")
// so the renderer knows which docs/ subdir it belongs in.
func collectDefs() {
	fset := token.NewFileSet()
	os.RemoveAll("defs")
	os.Mkdir("defs", 0777)

	dirs := []string{"./", "./util"}
	filepath.Walk("golibs/", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		dirs = append(dirs, "./"+path)
		return nil
	})

	pkgs := make(map[string]*ast.Package)
	for _, path := range dirs {
		d, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
			return
		}
		for k, v := range d {
			pkgs[k] = v
		}
	}

	// Go interfaces (@interface tags) become standalone modules of their own,
	// e.g. hilbish.jobs, keyed by "<parent>.<interface>".
	interfaceModules := make(map[string]*module)
	for l, f := range pkgs {
		p := doc.New(f, "./", doc.AllDecls)
		pieces := []docPiece{}
		typePieces := []docPiece{}
		mod := l
		if mod == "main" || mod == "util" {
			mod = "hilbish"
		}
		for _, t := range p.Funcs {
			piece := setupDoc(mod, t)
			if piece == nil {
				continue
			}

			pieces = append(pieces, *piece)
		}
		for _, t := range p.Types {
			typePiece := setupDocType(mod, t)
			if typePiece != nil {
				typePieces = append(typePieces, *typePiece)
			}

			for _, m := range t.Methods {
				piece := setupDoc(mod, m)
				if piece == nil {
					continue
				}

				pieces = append(pieces, *piece)
			}
		}

		tags, descParts := getTagsAndDocs(strings.TrimSpace(p.Doc), mod+" (package doc)")
		shortDesc := descParts[0]
		desc := descParts[1:]
		var modSince string
		if t := tags["since"]; len(t) > 0 {
			modSince = t[0].Id
			if modSince == "" && len(t[0].Fields) > 0 {
				modSince = strings.Join(t[0].Fields, " ")
			}
		}
		filteredPieces := []docPiece{}
		filteredTypePieces := []docPiece{}
		for _, piece := range pieces {
			if !piece.IsInterface {
				filteredPieces = append(filteredPieces, piece)
				continue
			}

			modname := piece.ParentModule + "." + piece.Interfacing
			if interfaceModules[modname] == nil {
				interfaceModules[modname] = &module{
					Name:         modname,
					Section:      "api",
					ParentModule: piece.ParentModule,
				}
			}

			if strings.HasSuffix(piece.GoFuncName, strings.ToLower("loader")) {
				shortDesc := piece.Doc[0]
				desc := piece.Doc[1:]
				interfaceModules[modname].ShortDescription = shortDesc
				interfaceModules[modname].Description = strings.Replace(strings.Join(desc, "\n"), "<nl>", "\n", -1)
				interfaceModules[modname].Fields = piece.Fields
				interfaceModules[modname].Properties = piece.Properties
				continue
			}

			interfaceModules[modname].Docs = append(interfaceModules[modname].Docs, piece)
		}

		for _, piece := range typePieces {
			if !piece.IsInterface {
				filteredTypePieces = append(filteredTypePieces, piece)
				continue
			}

			modname := piece.ParentModule + "." + piece.Interfacing
			if interfaceModules[modname] == nil {
				interfaceModules[modname] = &module{
					Name:         modname,
					Section:      "api",
					ParentModule: piece.ParentModule,
				}
			}

			interfaceModules[modname].Types = append(interfaceModules[modname].Types, piece)
		}

		if newDoc, ok := docs[mod]; ok {
			oldMod := docs[mod]
			newDoc.Types = append(filteredTypePieces, oldMod.Types...)
			newDoc.Docs = append(filteredPieces, oldMod.Docs...)
			if newDoc.ShortDescription == "" && shortDesc != "" {
				newDoc.ShortDescription = shortDesc
				newDoc.Description = strings.Replace(strings.Join(desc, "\n"), "<nl>", "\n", -1)
				newDoc.Properties = docPieceTagTyped("property", tags)
				newDoc.Fields = docPieceTagTyped("field", tags)
				newDoc.Since = modSince
			}

			docs[mod] = newDoc
		} else {
			docs[mod] = module{
				Name:             mod,
				Section:          "api",
				Types:            filteredTypePieces,
				Docs:             filteredPieces,
				ShortDescription: shortDesc,
				Since:            modSince,
				Description:      strings.Replace(strings.Join(desc, "\n"), "<nl>", "\n", -1),
				Properties:       docPieceTagTyped("property", tags),
				Fields:           docPieceTagTyped("field", tags),
			}
		}
	}

	// Lua-implemented modules (nature/*.lua). A Lua module that shares a name
	// with a Go module/interface (hilbish, hilbish.runner) is the Lua-side of
	// the same thing, so its functions merge into the existing def.
	for _, lmod := range collectLuaModules() {
		if existing, ok := docs[lmod.Name]; ok {
			existing.Docs = append(existing.Docs, lmod.Docs...)
			if existing.ShortDescription == "" {
				existing.ShortDescription = lmod.ShortDescription
				existing.Description = lmod.Description
				existing.Since = lmod.Since
			}
			docs[lmod.Name] = existing
			continue
		}
		if existing, ok := interfaceModules[lmod.Name]; ok {
			existing.Docs = append(existing.Docs, lmod.Docs...)
			if existing.ShortDescription == "" {
				existing.ShortDescription = lmod.ShortDescription
				existing.Description = lmod.Description
				existing.Since = lmod.Since
			}
			continue
		}
		docs[lmod.Name] = lmod
	}

	// Flatten everything (top-level modules + interfaces) and write one def
	// per page.
	all := make(map[string]module)
	for name, mod := range docs {
		all[name] = mod
	}
	for name, mod := range interfaceModules {
		all[name] = *mod
	}

	for name, v := range all {
		// The Docs/Types slices are merged across packages whose iteration
		// order (via the pkgs map) is randomized between runs, so sort them
		// here to keep defs/*.json output stable across docgen runs.
		sort.SliceStable(v.Docs, func(i, j int) bool {
			return v.Docs[i].FuncName < v.Docs[j].FuncName
		})
		sort.SliceStable(v.Types, func(i, j int) bool {
			return v.Types[i].FuncName < v.Types[j].FuncName
		})

		u, err := json.MarshalIndent(v, "", "	")
		if err != nil {
			panic(err)
		}

		f, err := os.Create("defs/" + name + ".json")
		if err != nil {
			panic(err)
		}
		f.WriteString(string(u))
		f.Close()
	}

	// collect hook documentation from nature/hooks.lua
	collectHooks()
}

// renderDocs reads every def written by collectDefs and renders it to
// docs/<section>/<name>.md. The api section is regenerated wholesale; the
// nature section is only overwritten per-file so hand-written pages such as
// docs/nature/_index.md survive.
func renderDocs() {
	os.Mkdir("docs", 0777)
	os.RemoveAll("docs/api")
	os.MkdirAll("docs/api", 0777)
	os.MkdirAll("docs/nature", 0777)

	f, err := os.Create("docs/api/_index.md")
	if err != nil {
		panic(err)
	}
	f.WriteString(`---
title: API
layout: doc
weight: -70
menu: docs
---

Welcome to the API documentation for Hilbish. This documents Lua functions
provided by Hilbish.
`)
	f.Close()

	defs, err := os.ReadDir("defs")
	if err != nil {
		panic(err)
	}

	for _, defEntry := range defs {
		name := defEntry.Name()
		// hook defs are rendered separately below
		if strings.HasPrefix(name, "hook.") {
			continue
		}
		defContent, err := os.ReadFile(filepath.Join("defs", name))
		if err != nil {
			panic(err)
		}

		var def module
		err = json.Unmarshal(defContent, &def)
		if err != nil {
			panic(err)
		}

		generateFile(def)
	}

	// render hook defs into docs/hooks/<group>.md.
	// MkdirAll (not RemoveAll) preserves the hand-written docs/hooks/_index.md.
	os.MkdirAll("docs/hooks", 0777)
	for _, defEntry := range defs {
		name := defEntry.Name()
		if !strings.HasPrefix(name, "hook.") {
			continue
		}
		defContent, err := os.ReadFile(filepath.Join("defs", name))
		if err != nil {
			panic(err)
		}
		var g hookGroup
		if err := json.Unmarshal(defContent, &g); err != nil {
			panic(err)
		}
		generateHookFile(g)
	}
}

// collectLuaModules parses the Lua-implemented modules under nature/ into the
// same module/docPiece structs the Go side produces, so a single renderer
// handles both. It ports the line-based parsing that used to live in
// cmd/docgen/docgen.lua: a leading `--- @module <name>` header, a top comment
// block as the description, and per-function doc comments (`@param`,
// `@return`, and `@example`...`@example` blocks).
func collectLuaModules() []module {
	var files []string
	for _, pat := range []string{"nature/*.lua", "nature/*/*.lua"} {
		matches, _ := filepath.Glob(pat)
		files = append(files, matches...)
	}

	modPattern := regexp.MustCompile(`^--+ @module (.+)`)
	docPattern := regexp.MustCompile(`^--+ (.+)`)
	blankDocPattern := regexp.MustCompile(`^--+\s*$`) // blank comment line — paragraph break
	emmyPattern := regexp.MustCompile(`^@(\w+)`)
	classPattern := regexp.MustCompile(`^@class\s+(\S+)`)
	fieldNamePattern := regexp.MustCompile(`^@field\s+(\S+)\s+(.*)$`)

	var mods []module
	for _, fname := range files {
		content, err := os.ReadFile(fname)
		if err != nil {
			continue
		}

		lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
		if len(lines) == 0 {
			continue
		}
		m := modPattern.FindStringSubmatch(lines[0])
		if m == nil {
			continue
		}
		modName := m[1]

		// the body is everything after the @module header line
		body := lines[1:]
		funcPattern := regexp.MustCompile(`^function ` + regexp.QuoteMeta(modName) + `\.(\w+)\(([^)]*)\)`)
		methodPattern := regexp.MustCompile(`^function [A-Z]\w*:(\w+)\(([^)]*)\)`)

		modSincePattern := regexp.MustCompile(`^@since\s+(.+)`)
		var descriptions []string
		var modSince string
		var pieces []docPiece
		var types []docPiece
		doingDescription := true

		// pre-pass: collect standalone `--- @class Name` / `--- @field ...` blocks
		// (fields-only Lua types, e.g. a duck-typed table with no methods) into
		// their own docPiece with IsType: true, mirroring the Go @type/@property path.
		{
			var blockLines []string
			flushClassBlock := func() {
				if len(blockLines) == 0 {
					return
				}
				var className string
				var classDoc []string
				var properties []docPiece
				for _, bl := range blockLines {
					dm := docPattern.FindStringSubmatch(bl)
					if dm == nil {
						continue
					}
					docline := dm[1]
					if cm := classPattern.FindStringSubmatch(docline); cm != nil {
						className = cm[1]
						continue
					}
					if fm := fieldNamePattern.FindStringSubmatch(docline); fm != nil {
						name := strings.TrimSuffix(fm[1], "?")
						rest := fm[2]

						// split type from doc by scanning space-separated tokens,
						// respecting (), <> nesting and `fun(...): Ret` return arrows,
						// so `fun(input: string): RunnerResult` stays one type token
						tokens := strings.Split(rest, " ")
						typEnd := len(rest)
						depth := 0
						consumed := 0
						for i, tok := range tokens {
							if i > 0 {
								prev := tokens[i-1]
								if depth == 0 && !strings.HasSuffix(prev, ":") {
									typEnd = consumed - 1 // back off the separating space
									break
								}
							}
							depth += strings.Count(tok, "(") + strings.Count(tok, "<")
							depth -= strings.Count(tok, ")") + strings.Count(tok, ">")
							consumed += len(tok) + 1
							typEnd = consumed - 1
						}
						var typ string
						var doc []string
						if typEnd < len(rest) {
							typ = rest[:typEnd]
							if docStr := strings.TrimSpace(rest[typEnd:]); docStr != "" {
								doc = []string{docStr}
							}
						} else {
							typ = rest
						}
						if strings.HasSuffix(fm[1], "?") && !strings.HasSuffix(typ, "?") {
							typ += "?"
						}
						properties = append(properties, docPiece{
							FuncName: name,
							Type:     typ,
							Doc:      doc,
						})
						continue
					}
					if emmyPattern.MatchString(docline) {
						continue
					}
					classDoc = append(classDoc, docline)
				}
				blockLines = nil
				if className == "" {
					return
				}
				types = append(types, docPiece{
					FuncName:   className,
					Doc:        classDoc,
					Properties: properties,
					IsType:     true,
				})
			}

			for _, line := range body {
				if docPattern.MatchString(line) || blankDocPattern.MatchString(line) {
					blockLines = append(blockLines, line)
					continue
				}
				flushClassBlock()
			}
			flushClassBlock()
		}

		for idx, line := range body {
			if dm := docPattern.FindStringSubmatch(line); dm != nil {
				if doingDescription {
					if sm := modSincePattern.FindStringSubmatch(dm[1]); sm != nil {
						modSince = sm[1]
						continue
					}
					descriptions = append(descriptions, dm[1])
				}
				continue
			}
			if blankDocPattern.MatchString(line) {
				if doingDescription {
					descriptions = append(descriptions, "<nl>")
				}
				continue
			}
			doingDescription = false

			var funcName, paramStr string
			if fm := funcPattern.FindStringSubmatch(line); fm != nil {
				funcName, paramStr = fm[1], fm[2]
			} else if mm := methodPattern.FindStringSubmatch(line); mm != nil {
				funcName, paramStr = mm[1], mm[2]
			}
			if funcName == "" {
				continue
			}

			// walk backwards over the preceding doc-comment block
			var descLines, exampleLines []string
			var params []param
			var returnParams []param
			defaultValues := make(map[string]string)
			tparamKeys := make(map[string][]param)
			treturnKeys := make(map[string][]param)
			isPrivate := false
			doingExample := false
			var since string
			var seeAlso []string
			var notes []string
			for offset := 1; idx-offset >= 0; offset++ {
				if blankDocPattern.MatchString(body[idx-offset]) {
					// blank comment line — paragraph break, not end of the doc-comment block
					if doingExample {
						exampleLines = append([]string{""}, exampleLines...)
					} else {
						descLines = append([]string{""}, descLines...)
					}
					continue
				}
				dm := docPattern.FindStringSubmatch(body[idx-offset])
				if dm == nil {
					break
				}
				docline := dm[1]

				if em := emmyPattern.FindStringSubmatch(docline); em != nil {
					emmy := em[1]
					rest := strings.TrimSpace(strings.TrimPrefix(docline, "@"+emmy))
					switch emmy {
					case "param":
						fields := strings.Split(rest, " ")
						p := param{}
						if len(fields) > 0 {
							name := fields[0]
							isOptional := strings.HasSuffix(name, "?")
							if isOptional {
								name = strings.TrimSuffix(name, "?")
							}
							p.Name = name
							p.IsOptional = isOptional
						}
						if len(fields) > 1 {
							p.Type = fields[1]
						}
						if len(fields) > 2 {
							p.Doc = []string{strings.Join(fields[2:], " ")}
						}
						p.Default = defaultValues[fields[0]]
						p.Keys = tparamKeys[fields[0]]
						params = append([]param{p}, params...)
					case "tparam":
						// @tparam <parentName> <keyName> <type> <doc...>
						fields := strings.SplitN(rest, " ", 4)
						if len(fields) >= 3 {
							parent := fields[0]
							keyName := fields[1]
							keyType := fields[2]
							var keyDoc []string
							if len(fields) > 3 {
								keyDoc = []string{fields[3]}
							}
							isOptional := strings.HasSuffix(keyName, "?")
							if isOptional {
								keyName = strings.TrimSuffix(keyName, "?")
							}
							// prepend to preserve source order (parser walks backwards)
							tparamKeys[parent] = append([]param{{
								Name:       keyName,
								Type:       keyType,
								Doc:        keyDoc,
								IsOptional: isOptional,
							}}, tparamKeys[parent]...)
						}
					case "default":
						fields := strings.SplitN(rest, " ", 2)
						if len(fields) > 0 {
							key := fields[0]
							val := ""
							if len(fields) > 1 {
								val = fields[1]
							}
							defaultValues[key] = val
						}
					case "return":
						// @return <type> <name> [doc...]; split type from the
						// rest respecting <> nesting, so table<string, string>
						// is treated as one type token
						rp := param{}
						typEnd := -1
						depth := 0
						for i, c := range rest {
							if c == '<' {
								depth++
							} else if c == '>' {
								depth--
							} else if c == ' ' && depth == 0 {
								typEnd = i
								break
							}
						}
						var nameAndDoc string
						if typEnd > 0 {
							rp.Type = rest[:typEnd]
							nameAndDoc = strings.TrimSpace(rest[typEnd:])
						} else {
							rp.Type = rest
						}
						if nameAndDoc != "" {
							fields := strings.SplitN(nameAndDoc, " ", 2)
							rp.Name = fields[0]
							if len(fields) > 1 {
								rp.Doc = []string{fields[1]}
							}
						}
						rp.IsOptional = strings.HasSuffix(rp.Type, "?")
						rp.Keys = treturnKeys[rp.Name]
						returnParams = append([]param{rp}, returnParams...)
					case "treturn":
						// @treturn <name> <keyName> <type> <doc...>
						fields := strings.SplitN(rest, " ", 4)
						if len(fields) >= 3 {
							parent := fields[0]
							keyName := fields[1]
							keyType := fields[2]
							var keyDoc []string
							if len(fields) > 3 {
								keyDoc = []string{fields[3]}
							}
							isOptional := strings.HasSuffix(keyName, "?")
							if isOptional {
								keyName = strings.TrimSuffix(keyName, "?")
							}
							// prepend to preserve source order (parser walks backwards)
							treturnKeys[parent] = append([]param{{
								Name:       keyName,
								Type:       keyType,
								Doc:        keyDoc,
								IsOptional: isOptional,
							}}, treturnKeys[parent]...)
						}
					case "since":
						since = rest
					case "see":
						seeAlso = append([]string{rest}, seeAlso...)
					case "note":
						notes = append([]string{rest}, notes...)
					case "private":
						isPrivate = true
					case "example":
						doingExample = !doingExample
					}
					continue
				}

				if doingExample {
					exampleLines = append([]string{docline}, exampleLines...)
				} else {
					descLines = append([]string{docline}, descLines...)
				}
			}

			if isPrivate {
				continue
			}

			// skip functions without any documentation at all
			if len(descLines) == 0 && len(params) == 0 && len(returnParams) == 0 {
				continue
			}

			// signature is stored without the module prefix; the renderer
			// prepends "<mod>." itself
			sig := fmt.Sprintf("%s(%s)", funcName, paramStr)
			if len(returnParams) > 0 {
				retTypes := make([]string, len(returnParams))
				for i, r := range returnParams {
					retTypes[i] = r.Type
				}
				sig += " -> " + strings.Join(retTypes, ", ")
			}

			piece := docPiece{
				FuncName: funcName,
				FuncSig:  sig,
				Doc:      descLines,
				Params:   params,
				Returns:  returnParams,
				Since:    since,
				SeeAlso:  seeAlso,
				Notes:    notes,
			}
			pieceTags := map[string][]tag{}
			if len(exampleLines) > 0 {
				pieceTags["example"] = []tag{{Fields: exampleLines}}
			}
			if len(pieceTags) > 0 {
				piece.Tags = pieceTags
			}
			pieces = append(pieces, piece)
		}

		section := "nature"
		if modName == "hilbish" || strings.HasPrefix(modName, "hilbish.") {
			section = "api"
		}

		var shortDesc, longDesc string
		if len(descriptions) > 0 {
			shortDesc = descriptions[0]
			longDesc = strings.Replace(strings.Join(descriptions[1:], "\n"), "<nl>", "\n", -1)
		}

		mods = append(mods, module{
			Name:             modName,
			Section:          section,
			ShortDescription: shortDesc,
			Description:      longDesc,
			Since:            modSince,
			Docs:             pieces,
			Types:            types,
		})
	}

	return mods
}

// hookPiece holds the documentation for a single bait hook (signal).
type hookPiece struct {
	Name    string   `json:"name"`
	Group   string   `json:"group"`
	Doc     []string `json:"description"`
	Vars    []param  `json:"vars,omitempty"`
	Since   string   `json:"since,omitempty"`
	Example []string `json:"example,omitempty"`
}

type hookGroup struct {
	Group string      `json:"group"`
	Hooks []hookPiece `json:"hooks"`
}

var hookGroupHeader = `---
title: %s
description:
layout: doc
menu:
  docs:
    parent: "Signals"
---

`

// collectHooks parses nature/hooks.lua, the central catalog of every bait signal
// Hilbish emits, into hookGroup/hookPiece structs and writes one
// defs/hook.<group>.json per group. Uses the same @-tag vocabulary as the rest
// of the docgen system.
func collectHooks() {
	content, err := os.ReadFile("nature/hooks.lua")
	if err != nil {
		return // file doesn't exist yet during bootstrap
	}

	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	docPattern := regexp.MustCompile(`^--+ (.+)`)
	emmyPat := regexp.MustCompile(`^@(\w+)`)

	groups := map[string]*hookGroup{}
	var current *hookPiece
	doingExample := false
	var exampleLines []string

	flush := func() {
		if current == nil {
			return
		}
		g := current.Group
		if groups[g] == nil {
			groups[g] = &hookGroup{Group: g}
		}
		groups[g].Hooks = append(groups[g].Hooks, *current)
		current = nil
	}

	for _, line := range lines {
		dm := docPattern.FindStringSubmatch(line)
		if dm == nil {
			continue
		}
		docline := dm[1]

		em := emmyPat.FindStringSubmatch(docline)
		if em == nil {
			// plain text line: goes to example or description
			if doingExample {
				exampleLines = append(exampleLines, docline)
			} else if current != nil {
				current.Doc = append(current.Doc, docline)
			}
			continue
		}

		tag := em[1]
		rest := strings.TrimSpace(strings.TrimPrefix(docline, "@"+tag))

		// example content takes priority: @var/@since inside an example are literal text
		if doingExample {
			if tag == "example" {
				// closing marker
				if current != nil {
					current.Example = exampleLines
				}
				doingExample = false
				exampleLines = nil
			} else {
				exampleLines = append(exampleLines, docline)
			}
			continue
		}

		switch tag {
		case "hooks":
			// file-type header marker, not a block tag
		case "hook":
			flush()
			group := ""
			if idx := strings.IndexByte(rest, '.'); idx >= 0 {
				group = rest[:idx]
			}
			current = &hookPiece{Name: rest, Group: group}
		case "group":
			if current != nil {
				current.Group = rest
			}
		case "var":
			if current != nil {
				fields := strings.SplitN(rest, " ", 3)
				v := param{}
				if len(fields) > 0 {
					v.Name = fields[0]
				}
				if len(fields) > 1 {
					v.Type = fields[1]
				}
				if len(fields) > 2 {
					v.Doc = []string{fields[2]}
				}
				current.Vars = append(current.Vars, v)
			}
		case "since":
			if current != nil {
				current.Since = rest
			}
		case "example":
			doingExample = true
			exampleLines = []string{}
		default:
			if !allowedTags[tag] {
				fmt.Printf("WARNING! Unknown tag '@%s' in hooks.lua\n", tag)
			}
		}
	}
	flush()

	// write defs/hook.<group>.json
	for _, g := range groups {
		sort.SliceStable(g.Hooks, func(i, j int) bool {
			return g.Hooks[i].Name < g.Hooks[j].Name
		})
		u, _ := json.MarshalIndent(g, "", "\t")
		os.WriteFile(filepath.Join("defs", "hook."+g.Group+".json"), u, 0644)
	}
}

// generateHookFile renders a hookGroup into docs/hooks/<group>.md, matching the
// existing hand-written format so diffs are reviewable.
func generateHookFile(g hookGroup) {
	title := strings.ToUpper(string(g.Group[0])) + g.Group[1:]
	f, err := os.Create(filepath.Join("docs", "hooks", g.Group+".md"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf(hookGroupHeader, title))

	for i, h := range g.Hooks {
		f.WriteString(fmt.Sprintf("## %s\n\n", h.Name))
		for _, d := range h.Doc {
			f.WriteString(d + "\n")
		}
		f.WriteString("\n")
		if h.Since != "" {
			f.WriteString(fmt.Sprintf("Since: `%s`\n\n", h.Since))
		}
		f.WriteString(heading("Variables", 4))
		if len(h.Vars) == 0 {
			f.WriteString("This signal has no variables.\n")
		}
		for _, v := range h.Vars {
			f.WriteString(fmt.Sprintf("`%s` _%s_  \n", v.Type, v.Name))
			if len(v.Doc) > 0 {
				f.WriteString(strings.Join(v.Doc, " ") + "\n")
			}
			f.WriteString("\n")
		}
		if len(h.Example) > 0 {
			f.WriteString(heading("Example", 4))
			f.WriteString(fmt.Sprintf("```lua\n%s\n```\n", strings.Join(h.Example, "\n")))
		}
		if i < len(g.Hooks)-1 {
			f.WriteString("\n``` =html\n<hr class=\"my-4\">\n```\n\n")
		}
	}
}

// funcReturns extracts the documented return values from a docPiece. It prefers
// the structured Returns slice (populated from @return/@treturn tags for both
// Go-sourced defs and Lua defs), and falls back to parsing the "-> type, type"
// embedded in FuncSig for older or undescribed Lua returns.
func funcReturns(dps docPiece) []param {
	if len(dps.Returns) > 0 {
		return dps.Returns
	}
	retStr := luaSigReturn(dps.FuncSig)
	if retStr == "" {
		return nil
	}
	parts := strings.Split(retStr, ",")
	result := make([]param, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			result = append(result, param{Type: t})
		}
	}
	return result
}

// writeFuncBlock writes the full documented block for a single function/method:
// separator, name heading, fenced signature, description, Parameters, Returns,
// and Example(s). Used for both top-level module functions and type Methods so
// both get identical detail.
func writeFuncBlock(f *os.File, mod string, dps docPiece) {
	f.WriteString("---\n\n")
	f.WriteString(heading(dps.FuncName, 4))

	separator := "."
	if dps.IsMember {
		separator = ":"
	}

	// signature in a fenced :::signature div for distinct styling
	f.WriteString(fmt.Sprintf(":::signature\n```lua\n%s%s%s\n```\n:::\n\n", mod, separator, dps.FuncSig))

	if dps.Since != "" {
		f.WriteString(fmt.Sprintf("Since: `%s`\n\n", dps.Since))
	}

	for _, note := range dps.Notes {
		f.WriteString(fmt.Sprintf(":::note\n%s\n:::\n\n", note))
	}

	for _, doc := range dps.Doc {
		if strings.HasPrefix(doc, "---") {
			continue
		}
		if doc == "" {
			f.WriteString("\n")
			continue
		}
		f.WriteString(doc + "  \n")
	}
	f.WriteString("\n")

	if len(dps.Params) != 0 {
		// parameters in a :::params div
		f.WriteString(heading("Parameters", 4))
		f.WriteString(":::params\n")
		for _, p := range dps.Params {
			isVariadic := strings.HasPrefix(p.Type, "...")
			typ := p.Type
			if isVariadic {
				typ = p.Type[3:]
			}
			f.WriteString(fmt.Sprintf("`%s` _%s_", typ, p.Name))
			if isVariadic {
				f.WriteString(" [Variadic]{.variadic}")
			}
			if p.IsOptional {
				f.WriteString(" [Optional]{.optional}")
			}
			f.WriteString("  \n")
			if len(p.Doc) > 0 {
				f.WriteString(strings.Join(p.Doc, " ") + "\n")
			}
			if p.Default != "" {
				fmt.Fprintf(f, "Default: `%s`\n", p.Default)
			}
			if len(p.Keys) > 0 {
				f.WriteString("\n:::tparams\n")
				for _, k := range p.Keys {
					fmt.Fprintf(f, "`%s` _%s_", k.Type, k.Name)
					if k.IsOptional {
						f.WriteString(" [Optional]{.optional}")
					}
					f.WriteString("  \n")
					if len(k.Doc) > 0 {
						f.WriteString(strings.Join(k.Doc, " ") + "\n")
					}
					if k.Default != "" {
						fmt.Fprintf(f, "Default: `%s`\n", k.Default)
					}
					f.WriteString("\n")
				}
				f.WriteString(":::\n")
			}
			f.WriteString("\n")
		}
		f.WriteString(":::\n\n")
	}

	rets := funcReturns(dps)
	if len(rets) != 0 {
		// returns in a :::returns div
		f.WriteString(heading("Returns", 4))
		f.WriteString(":::returns\n")
		for _, r := range rets {
			isOptional := strings.HasSuffix(r.Type, "?")
			f.WriteString(fmt.Sprintf("`%s`", r.Type))
			if isOptional {
				f.WriteString(" [Optional]{.optional}")
			}
			f.WriteString("  \n")
			if len(r.Doc) > 0 {
				f.WriteString(strings.Join(r.Doc, " ") + "\n")
			}
			if len(r.Keys) > 0 {
				f.WriteString("\n:::tparams\n")
				for _, k := range r.Keys {
					fmt.Fprintf(f, "`%s` _%s_", k.Type, k.Name)
					if k.IsOptional {
						f.WriteString(" [Optional]{.optional}")
					}
					f.WriteString("  \n")
					if len(k.Doc) > 0 {
						f.WriteString(strings.Join(k.Doc, " ") + "\n")
					}
					if k.Default != "" {
						fmt.Fprintf(f, "Default: `%s`\n", k.Default)
					}
					f.WriteString("\n")
				}
				f.WriteString(":::\n")
			}
			f.WriteString("\n")
		}
		f.WriteString(":::\n\n")
	}

	if len(dps.SeeAlso) != 0 {
		f.WriteString(heading("See also", 4))
		for _, ref := range dps.SeeAlso {
			if strings.Contains(ref, ".") {
				lastDot := strings.LastIndex(ref, ".")
				modPart := ref[:lastDot]
				fnPart := ref[lastDot+1:]
				f.WriteString(fmt.Sprintf("- [`%s`](../api/%s/#%s)\n", ref, modPart, fnPart))
			} else {
				f.WriteString(fmt.Sprintf("- [`%s`](#%s)\n", ref, ref))
			}
		}
		f.WriteString("\n")
	}

	if examples := dps.Tags["example"]; len(examples) > 0 {
		if len(examples) == 1 {
			f.WriteString(heading("Example", 4))
		} else {
			f.WriteString(heading("Examples", 4))
		}
		for _, ex := range examples {
			f.WriteString(fmt.Sprintf("```lua\n%s\n```\n", strings.Join(ex.Fields, "\n")))
		}
	}
	f.WriteString("\n\n")
}

func generateFile(v module) {
	mod := v.Name
	section := v.Section
	if section == "" {
		section = "api"
	}
	docParent := "API"
	if section == "nature" {
		docParent = "Nature"
	}
	docPath := "docs/" + section + "/" + mod + ".md"

	sort.SliceStable(v.Docs, func(i, j int) bool {
		return v.Docs[i].FuncName < v.Docs[j].FuncName
	})

	f, _ := os.Create(docPath)
	f.WriteString(fmt.Sprintf(header, mod, v.ShortDescription, docParent))
	if v.Since != "" {
		f.WriteString(fmt.Sprintf("_Added in v%s_\n\n", v.Since))
	}
	modDescription := v.Description
	f.WriteString(heading("Introduction", 2))
	f.WriteString(modDescription)
	f.WriteString("\n\n")
	if len(v.Docs) != 0 {
		f.WriteString(heading("Functions", 2))
		f.WriteString(":::funclist\n")
		funcList := [][]string{}
		for _, dps := range v.Docs {
			if dps.IsMember {
				continue
			}

			if len(dps.Doc) == 0 {
				fmt.Printf("WARNING! Function %s on module %s has no documentation!\n", dps.FuncName, mod)
				continue
			}

			funcList = append(funcList, []string{
				fmt.Sprintf("[`%s.%s`](#%s)", mod, dps.FuncSig, dps.FuncName),
				dps.Doc[0],
			})
		}
		f.WriteString(bulletList(funcList))
		f.WriteString(":::\n\n")
	}

	if len(v.Fields) != 0 {
		f.WriteString(heading("Static module fields", 2))
		f.WriteString(":::fieldlist\n")
		fieldsList := [][]string{}
		for _, dps := range v.Fields {
			label := fmt.Sprintf("`%s`", dps.FuncName)
			if dps.Type != "" {
				label = fmt.Sprintf("`%s` `%s`", dps.Type, dps.FuncName)
			}
			fieldsList = append(fieldsList, []string{label, strings.Join(dps.Doc, " ")})
		}
		f.WriteString(bulletList(fieldsList))
		f.WriteString(":::\n\n")
	}
	if len(v.Properties) != 0 {
		f.WriteString(heading("Object properties", 2))
		f.WriteString(":::fieldlist\n")
		propertiesList := [][]string{}
		for _, dps := range v.Properties {
			label := fmt.Sprintf("`%s`", dps.FuncName)
			if dps.Type != "" {
				label = fmt.Sprintf("`%s` `%s`", dps.Type, dps.FuncName)
			}
			propertiesList = append(propertiesList, []string{label, strings.Join(dps.Doc, " ")})
		}
		f.WriteString(bulletList(propertiesList))
		f.WriteString(":::\n\n")
	}

	if len(v.Docs) != 0 {
		for _, dps := range v.Docs {
			if dps.IsMember {
				continue
			}
			writeFuncBlock(f, mod, dps)
		}
	}

	if len(v.Types) != 0 {
		f.WriteString(heading("Types", 2))
		for _, dps := range v.Types {
			f.WriteString("---\n\n")
			f.WriteString(heading(dps.FuncName, 2))
			for _, doc := range dps.Doc {
				if !strings.HasPrefix(doc, "---") {
					f.WriteString(doc + "\n")
				}
			}
			f.WriteString("\n")
			if len(dps.Properties) != 0 {
				f.WriteString(heading("Object Properties", 2))

				propertiesList := [][]string{}
				for _, p := range dps.Properties {
					label := fmt.Sprintf("`%s`", p.FuncName)
					if p.Type != "" {
						label = fmt.Sprintf("`%s` `%s`", p.Type, p.FuncName)
					}
					propertiesList = append(propertiesList, []string{label, strings.Join(p.Doc, " ")})
				}
				f.WriteString(bulletList(propertiesList))
			}
			f.WriteString("\n")
			f.WriteString(heading("Methods", 3))
			for _, dps := range v.Docs {
				if !dps.IsMember {
					continue
				}
				writeFuncBlock(f, mod, dps)
			}
		}
	}
}

func heading(name string, level int) string {
	return fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), name)
}

func bulletList(elems [][]string) string {
	var b strings.Builder
	for _, line := range elems {
		b.WriteString(fmt.Sprintf("- %s: %s\n", line[0], line[1]))
	}
	b.WriteString("\n")

	return b.String()
}

// renderLuaDefs reads every def written by collectDefs and renders LuaLS
// @meta definition files into types/. These teach the language server about
// Go-implemented globals and modules so that hilbish-specific symbols are
// recognised in editor without generating undefined-global warnings.
func renderLuaDefs() {
	os.RemoveAll("types")
	os.MkdirAll("types", 0777)

	// Static file for bare globals injected by Go but not in any def.
	os.WriteFile("types/globals.lua", []byte(
		"---@meta\n\n"+
			"---@type string[]\nargs = {}\n\n"+
			"---@type table<string, string>\nenv = {}\n\n"+
			// golua oslib extension (not in any workspace Lua file).
			"---@param name string\n---@param value string\n"+
			"function os.setenv(name, value) end\n",
	), 0644)

	defs, err := os.ReadDir("defs")
	if err != nil {
		panic(err)
	}

	var hilbishMods []module
	for _, entry := range defs {
		// hook defs are not Lua modules; skip them here
		if strings.HasPrefix(entry.Name(), "hook.") {
			continue
		}
		content, err := os.ReadFile(filepath.Join("defs", entry.Name()))
		if err != nil {
			continue
		}
		var mod module
		json.Unmarshal(content, &mod)

		if mod.Name == "hilbish" || strings.HasPrefix(mod.Name, "hilbish.") {
			hilbishMods = append(hilbishMods, mod)
			continue
		}
		if mod.Section != "api" {
			// nature/ modules already have Lua source visible to LuaLS.
			continue
		}
		writeLuaModDef(mod)
	}

	sort.Slice(hilbishMods, func(i, j int) bool {
		a, b := hilbishMods[i].Name, hilbishMods[j].Name
		if a == "hilbish" {
			return true
		}
		if b == "hilbish" {
			return false
		}
		return a < b
	})
	writeLuaHilbishDef(hilbishMods)
}

// luaSigName extracts the function name from a signature like "foo(a, b) -> x".
func luaSigName(sig string) string {
	idx := strings.Index(sig, "(")
	if idx < 0 {
		return strings.TrimSpace(sig)
	}
	return strings.TrimSpace(sig[:idx])
}

// Matches "word (type)" descriptions like "entries (table)" or "prefix (string)".
// The space before "(" distinguishes these from function-call syntax "fun(args)".
var reDescType = regexp.MustCompile(`\w+\s+\(([^)]+)\)`)

// Matches table[X] (should be table<X> in LuaLS).
var reTableBracket = regexp.MustCompile(`table\[([^\]]*)\]`)

// Matches the "function" keyword optionally followed by a param list like
// "function(...)" so the whole thing can be replaced with "fun(...: any)".
var reBareFunction = regexp.MustCompile(`\bfunction\b(\s*\([^)]*\))?`)

// sanitizeLuaType normalises a raw type string from defs JSON into valid LuaLS
// syntax. Handles three cases from legacy doc comments:
//
//	"@Job"              → "Job"          (leading @ is Hilbish's internal ref prefix)
//	"table[Job]"        → "table<Job>"   (square → angle brackets)
//	"entries (table)"   → "table"        (name-then-parenthesised-type description)
func sanitizeLuaType(t string) string {
	t = strings.TrimSpace(t)
	// Strip leading @
	t = strings.TrimPrefix(t, "@")
	// Handle "name (type)" description style — keep only the type in parens.
	// e.g. "entries (table), prefix (string)" → "table, string"
	t = reDescType.ReplaceAllStringFunc(t, func(m string) string {
		sub := reDescType.FindStringSubmatch(m)
		if len(sub) > 1 {
			return strings.TrimSpace(sub[1])
		}
		return m
	})
	// table[X] → table<X>
	t = reTableBracket.ReplaceAllString(t, "table<$1>")
	// "function" or "function(...)" → "fun(...: any)"
	t = reBareFunction.ReplaceAllString(t, "fun(...: any)")
	return t
}

// luaSigParamNames extracts bare parameter names from a signature string like
// "write(str, flags)" → ["str", "flags"]. Used as a fallback when doc.Params
// is empty (common for Go-implemented member functions).
func luaSigParamNames(sig string) []string {
	start := strings.Index(sig, "(")
	end := strings.Index(sig, ")")
	if start < 0 || end < 0 || end <= start+1 {
		return nil
	}
	inner := strings.TrimSpace(sig[start+1 : end])
	if inner == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	names := make([]string, len(parts))
	for i, p := range parts {
		names[i] = strings.TrimSpace(p)
	}
	return names
}

// luaSigReturn extracts a return type embedded in the signature after "->".
// Lua-source defs store return info this way instead of in Tags.
func luaSigReturn(sig string) string {
	idx := strings.Index(sig, "->")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(sig[idx+2:])
}

// luaReturnStr collects the documented return types (name dropped, since
// LuaLS's "fun(...): ret" syntax has no room for one) from doc.Returns first,
// falling back to the embedded signature return.
func luaReturnStr(doc docPiece) string {
	if len(doc.Returns) > 0 {
		types := make([]string, 0, len(doc.Returns))
		for _, r := range doc.Returns {
			if r.Type != "" {
				types = append(types, r.Type)
			}
		}
		if len(types) > 0 {
			return strings.Join(types, ", ")
		}
	}
	return luaSigReturn(doc.FuncSig)
}

// luaFunField builds a "fun(params): ret" type string for use in ---@field.
func luaFunField(doc docPiece) string {
	parts := make([]string, 0, len(doc.Params))
	for _, p := range doc.Params {
		typ := sanitizeLuaType(p.Type)
		if typ == "" {
			typ = "any"
		}
		if strings.HasPrefix(typ, "...") {
			varType := strings.TrimPrefix(typ, "...")
			if varType == "" {
				parts = append(parts, "...")
			} else {
				parts = append(parts, "...: "+varType)
			}
		} else {
			name := p.Name
			if p.IsOptional {
				name = name + "?"
			}
			parts = append(parts, name+": "+typ)
		}
	}
	// If no typed params, fall back to names extracted from the signature.
	if len(parts) == 0 {
		for _, name := range luaSigParamNames(doc.FuncSig) {
			parts = append(parts, name+": any")
		}
	}

	var ret string
	if r := luaReturnStr(doc); r != "" {
		ret = ": " + sanitizeLuaType(r)
	}

	return fmt.Sprintf("fun(%s)%s", strings.Join(parts, ", "), ret)
}

// luaMethodDecl writes param/return annotations and a function declaration
// for a member method on className.
// luaClassAnchor returns the statement that gives a `---@class` block a
// concrete variable to attach `function X:method() end` declarations to.
// `local a.b = {}` is invalid Lua syntax, so dotted class names (e.g.
// "hilbish.message") must be declared as a plain field assignment instead of
// a local.
func luaClassAnchor(className string) string {
	if strings.Contains(className, ".") {
		return fmt.Sprintf("%s = {}\n\n", className)
	}
	return fmt.Sprintf("local %s = {}\n\n", className)
}

func luaMethodDecl(className string, doc docPiece) string {
	name := luaSigName(doc.FuncSig)
	if name == "" {
		return ""
	}
	var b strings.Builder

	// Prefer typed params from doc.Params; fall back to sig names as untyped.
	if len(doc.Params) > 0 {
		for _, p := range doc.Params {
			typ := sanitizeLuaType(p.Type)
			if typ == "" {
				typ = "any"
			}
			paramName := p.Name
			if p.IsOptional {
				paramName = paramName + "?"
			}
			b.WriteString(fmt.Sprintf("---@param %s %s\n", paramName, typ))
		}
	}
	if len(doc.Returns) > 0 {
		for _, r := range doc.Returns {
			typ := sanitizeLuaType(r.Type)
			if typ == "" {
				typ = "any"
			}
			name := r.Name
			if name == "" {
				name = "_"
			}
			b.WriteString(fmt.Sprintf("---@return %s %s\n", typ, name))
		}
	} else if r := luaReturnStr(doc); r != "" {
		b.WriteString(fmt.Sprintf("---@return %s\n", sanitizeLuaType(r)))
	}

	var paramList []string
	if len(doc.Params) > 0 {
		paramList = make([]string, len(doc.Params))
		for i, p := range doc.Params {
			paramList[i] = p.Name
		}
	} else {
		paramList = luaSigParamNames(doc.FuncSig)
	}
	b.WriteString(fmt.Sprintf("function %s:%s(%s) end\n\n", className, name, strings.Join(paramList, ", ")))
	return b.String()
}

// writeLuaModDef generates types/<modname>.lua for a single top-level api module.
func writeLuaModDef(mod module) {
	var b strings.Builder
	b.WriteString("---@meta\n\n")

	var memberDocs, modDocs []docPiece
	for _, d := range mod.Docs {
		if d.IsMember {
			memberDocs = append(memberDocs, d)
		} else {
			modDocs = append(modDocs, d)
		}
	}

	// Instance types (e.g. Readline, Snail, Thread).
	for _, typ := range mod.Types {
		b.WriteString(fmt.Sprintf("---@class %s\n", typ.FuncName))
		for _, p := range typ.Properties {
			b.WriteString(fmt.Sprintf("---@field %s any\n", p.FuncName))
		}
		b.WriteString(luaClassAnchor(typ.FuncName))
		for _, d := range memberDocs {
			b.WriteString(luaMethodDecl(typ.FuncName, d))
		}
	}

	// Module class — functions declared as @field fun(...) to avoid forward
	// references to the module variable.
	b.WriteString(fmt.Sprintf("---@class %s\n", mod.Name))
	for _, f := range mod.Fields {
		b.WriteString(fmt.Sprintf("---@field %s any\n", f.FuncName))
	}
	for _, d := range modDocs {
		name := luaSigName(d.FuncSig)
		if name == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("---@field %s %s\n", name, luaFunField(d)))
	}
	b.WriteString(luaClassAnchor(mod.Name))
	b.WriteString(fmt.Sprintf("return %s\n", mod.Name))

	os.WriteFile(filepath.Join("types", mod.Name+".lua"), []byte(b.String()), 0644)
}

// writeLuaHilbishDef generates types/hilbish.lua from all hilbish and
// hilbish.* defs merged into a single file.
func writeLuaHilbishDef(mods []module) {
	var b strings.Builder
	b.WriteString("---@meta\n\n")

	var mainMod module
	var subMods []module
	for _, m := range mods {
		if m.Name == "hilbish" {
			mainMod = m
		} else {
			subMods = append(subMods, m)
		}
	}

	// Sink type from the main hilbish module.
	var sinkMembers []docPiece
	for _, d := range mainMod.Docs {
		if d.IsMember {
			sinkMembers = append(sinkMembers, d)
		}
	}
	for _, typ := range mainMod.Types {
		b.WriteString(fmt.Sprintf("---@class %s\n", typ.FuncName))
		for _, p := range typ.Properties {
			b.WriteString(fmt.Sprintf("---@field %s any\n", p.FuncName))
		}
		b.WriteString(luaClassAnchor(typ.FuncName))
		for _, d := range sinkMembers {
			b.WriteString(luaMethodDecl(typ.FuncName, d))
		}
	}

	// Sub-module instance types and their classes.
	for _, sub := range subMods {
		var memberDocs []docPiece
		for _, d := range sub.Docs {
			if d.IsMember {
				memberDocs = append(memberDocs, d)
			}
		}
		for _, typ := range sub.Types {
			b.WriteString(fmt.Sprintf("---@class %s\n", typ.FuncName))
			for _, p := range typ.Properties {
				b.WriteString(fmt.Sprintf("---@field %s any\n", p.FuncName))
			}
			b.WriteString(luaClassAnchor(typ.FuncName))
			for _, d := range memberDocs {
				b.WriteString(luaMethodDecl(typ.FuncName, d))
			}
		}

		// Sub-module class (e.g. hilbish.jobs, hilbish.aliases).
		b.WriteString(fmt.Sprintf("---@class %s\n", sub.Name))
		for _, f := range sub.Fields {
			b.WriteString(fmt.Sprintf("---@field %s any\n", f.FuncName))
		}
		for _, d := range sub.Docs {
			if d.IsMember {
				continue
			}
			name := luaSigName(d.FuncSig)
			if name == "" {
				continue
			}
			b.WriteString(fmt.Sprintf("---@field %s %s\n", name, luaFunField(d)))
		}
		// Extra fields added by nature/ Lua files for specific sub-modules.
		// switch sub.Name {
		// case "hilbish.aliases":
		// 	b.WriteString("---@field all table<string, string>\n")
		// case "hilbish.runner":
		// 	b.WriteString("---@field sh fun(input: string)\n")
		// case "hilbish.processors":
		// 	b.WriteString("---@field add fun(processor: table)\n")
		// 	b.WriteString("---@field list table\n")
		// 	b.WriteString("---@field sorted table\n")
		// }
		b.WriteString("\n")
	}

	// Main Hilbish class.
	b.WriteString("---@class Hilbish\n")
	for _, f := range mainMod.Fields {
		b.WriteString(fmt.Sprintf("---@field %s any\n", f.FuncName))
	}
	// Extra fields defined by nature/ Lua files, not present in Go defs.
	b.WriteString("---@field home string\n")
	b.WriteString("---@field editor Readline\n")
	b.WriteString("---@field snail Snail\n")
	b.WriteString("---@field history table\n")
	b.WriteString("---@field opts table\n")
	b.WriteString("---@field vim table\n")
	b.WriteString("---@field sink { new: fun(): Sink }\n")
	b.WriteString("---@field motd string\n")
	b.WriteString("---@field hinter fun(line: string, pos: number): string\n")
	b.WriteString("---@field highlighter fun(line: string): string\n")
	b.WriteString("---@field inputMode fun(mode: string)\n")
	b.WriteString("---@field appendPath fun(path: string|table)\n")
	b.WriteString("---@field prependPath fun(path: string|table)\n")
	// Sub-module references.
	for _, sub := range subMods {
		fieldName := strings.TrimPrefix(sub.Name, "hilbish.")
		b.WriteString(fmt.Sprintf("---@field %s %s\n", fieldName, sub.Name))
	}
	// Main module functions.
	for _, d := range mainMod.Docs {
		if d.IsMember {
			continue
		}
		name := luaSigName(d.FuncSig)
		if name == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("---@field %s %s\n", name, luaFunField(d)))
	}
	b.WriteString("\n")

	b.WriteString("---@type Hilbish\n---@diagnostic disable-next-line: missing-fields\nhilbish = {}\n\nreturn hilbish\n")

	os.WriteFile(filepath.Join("types", "hilbish.lua"), []byte(b.String()), 0644)
}
