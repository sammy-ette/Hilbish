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
}

type param struct {
	Name string
	Type string
	Doc  []string
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
	Tags         map[string][]tag `json:"tags,omitempty"`
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

func getTagsAndDocs(docs string) (map[string][]tag, []string) {
	pts := strings.Split(docs, "\n")
	parts := []string{}
	tags := make(map[string][]tag)

	for idx, part := range pts {
		if strings.HasPrefix(part, "#") {
			tagParts := strings.Split(strings.TrimPrefix(part, "#"), " ")
			if tags[tagParts[0]] == nil {
				var id string
				if len(tagParts) > 1 {
					id = tagParts[1]
				}
				tags[tagParts[0]] = []tag{
					{Id: id, StartIdx: idx},
				}
				if len(tagParts) >= 2 {
					tags[tagParts[0]][0].Fields = tagParts[2:]
				}
			} else {
				if tagParts[0] == "example" {
					exampleIdx := tags["example"][0].StartIdx
					exampleCode := pts[exampleIdx+1 : idx]

					tags["example"][0].Fields = exampleCode
					parts = strings.Split(strings.Replace(strings.Join(parts, "\n"), strings.TrimPrefix(strings.Join(exampleCode, "\n"), "#example\n"), "", -1), "\n")
					continue
				}

				fleds := []string{}
				if len(tagParts) >= 2 {
					fleds = tagParts[2:]
				}
				tags[tagParts[0]] = append(tags[tagParts[0]], tag{
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

func setupDocType(mod string, typ *doc.Type) *docPiece {
	docs := strings.TrimSpace(typ.Doc)
	tags, doc := getTagsAndDocs(docs)

	if tags["type"] == nil {
		return nil
	}
	inInterface := tags["interface"] != nil

	var interfaces string
	typeName := strings.ToUpper(string(typ.Name[0])) + typ.Name[1:]
	typeDoc := []string{}

	if inInterface {
		interfaces = tags["interface"][0].Id
	}

	fields := docPieceTag("field", tags)
	properties := docPieceTag("property", tags)

	for _, d := range doc {
		if strings.HasPrefix(d, "---") {
			// TODO: document types in lua
			/*
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
			*/
		} else {
			typeDoc = append(typeDoc, d)
		}
	}

	var isMember bool
	if tags["member"] != nil {
		isMember = true
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
	}

	typeTable[strings.ToLower(typeName)] = []string{parentMod, interfaces}

	return dps
}

func setupDoc(mod string, fun *doc.Func) *docPiece {
	if fun.Doc == "" {
		return nil
	}

	docs := strings.TrimSpace(fun.Doc)
	tags, parts := getTagsAndDocs(docs)

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
	funcsig := parts[0]
	doc := parts[1:]
	funcName := strings.TrimPrefix(fun.Name, prefix[mod])
	funcdoc := []string{}

	if inInterface {
		interfaces = tags["interface"][0].Id
		funcName = interfaces + "." + strings.Split(funcsig, "(")[0]
	}
	em := emmyPiece{FuncName: funcName}

	fields := docPieceTag("field", tags)
	properties := docPieceTag("property", tags)
	var params []param
	if paramsRaw := tags["param"]; paramsRaw != nil {
		params = make([]param, len(paramsRaw))
		for i, p := range paramsRaw {
			params[i] = param{
				Name: p.Id,
				Type: p.Fields[0],
				Doc:  p.Fields[1:],
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
		Tags:         tags,
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

	// Go interfaces (#interface tags) become standalone modules of their own,
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

		tags, descParts := getTagsAndDocs(strings.TrimSpace(p.Doc))
		shortDesc := descParts[0]
		desc := descParts[1:]
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
				newDoc.Properties = docPieceTag("property", tags)
				newDoc.Fields = docPieceTag("field", tags)
			}

			docs[mod] = newDoc
		} else {
			docs[mod] = module{
				Name:             mod,
				Section:          "api",
				Types:            filteredTypePieces,
				Docs:             filteredPieces,
				ShortDescription: shortDesc,
				Description:      strings.Replace(strings.Join(desc, "\n"), "<nl>", "\n", -1),
				Properties:       docPieceTag("property", tags),
				Fields:           docPieceTag("field", tags),
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
			}
			docs[lmod.Name] = existing
			continue
		}
		if existing, ok := interfaceModules[lmod.Name]; ok {
			existing.Docs = append(existing.Docs, lmod.Docs...)
			if existing.ShortDescription == "" {
				existing.ShortDescription = lmod.ShortDescription
				existing.Description = lmod.Description
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
		defContent, err := os.ReadFile(filepath.Join("defs", defEntry.Name()))
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
}

// collectLuaModules parses the Lua-implemented modules under nature/ into the
// same module/docPiece structs the Go side produces, so a single renderer
// handles both. It ports the line-based parsing that used to live in
// cmd/docgen/docgen.lua: a leading `--- @module <name>` header, a top comment
// block as the description, and per-function doc comments (`@param`,
// `@return`/`@returns`, and `#example`...`#example` blocks).
func collectLuaModules() []module {
	var files []string
	for _, pat := range []string{"nature/*.lua", "nature/*/*.lua"} {
		matches, _ := filepath.Glob(pat)
		files = append(files, matches...)
	}

	modPattern := regexp.MustCompile(`^--+ @module (.+)`)
	docPattern := regexp.MustCompile(`^--+ (.+)`)
	emmyPattern := regexp.MustCompile(`^@(\w+)`)

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

		var descriptions []string
		var pieces []docPiece
		doingDescription := true

		for idx, line := range body {
			if dm := docPattern.FindStringSubmatch(line); dm != nil {
				if doingDescription {
					descriptions = append(descriptions, dm[1])
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
			var returns []string
			doingExample := false
			for offset := 1; idx-offset >= 0; offset++ {
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
							p.Name = fields[0]
						}
						if len(fields) > 1 {
							p.Type = fields[1]
						}
						if len(fields) > 2 {
							p.Doc = []string{strings.Join(fields[2:], " ")}
						}
						params = append([]param{p}, params...)
					case "return", "returns":
						returns = append([]string{rest}, returns...)
					}
					continue
				}

				if strings.Contains(docline, "#example") {
					doingExample = !doingExample
					continue
				}
				if doingExample {
					exampleLines = append([]string{docline}, exampleLines...)
				} else {
					descLines = append([]string{docline}, descLines...)
				}
			}

			// skip functions without any documentation at all
			if len(descLines) == 0 && len(params) == 0 && len(returns) == 0 {
				continue
			}

			// signature is stored without the module prefix; the renderer
			// prepends "<mod>." itself
			sig := fmt.Sprintf("%s(%s)", funcName, paramStr)
			if len(returns) > 0 {
				sig += " -> " + strings.Join(returns, ", ")
			}

			piece := docPiece{
				FuncName: funcName,
				FuncSig:  sig,
				Doc:      descLines,
				Params:   params,
			}
			if len(exampleLines) > 0 {
				piece.Tags = map[string][]tag{
					"example": {{Fields: exampleLines}},
				}
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
			longDesc = strings.Join(descriptions[1:], "\n")
		}

		mods = append(mods, module{
			Name:             modName,
			Section:          section,
			ShortDescription: shortDesc,
			Description:      longDesc,
			Docs:             pieces,
		})
	}

	return mods
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
	typeTag, _ := regexp.Compile(`\B@\w+`)
	modDescription := v.Description
	f.WriteString(heading("Introduction", 2))
	f.WriteString(modDescription)
	f.WriteString("\n\n")
	if len(v.Docs) != 0 {
		f.WriteString(heading("Functions", 2))

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
	}

	if len(v.Fields) != 0 {
		f.WriteString(heading("Static module fields", 2))

		fieldsList := [][]string{}
		for _, dps := range v.Fields {
			fieldsList = append(fieldsList, []string{fmt.Sprintf("`%s`", dps.FuncName), strings.Join(dps.Doc, " ")})
		}
		f.WriteString(bulletList(fieldsList))
	}
	if len(v.Properties) != 0 {
		f.WriteString(heading("Object properties", 2))

		propertiesList := [][]string{}
		for _, dps := range v.Properties {
			propertiesList = append(propertiesList, []string{fmt.Sprintf("`%s`", dps.FuncName), strings.Join(dps.Doc, " ")})
		}
		f.WriteString(bulletList(propertiesList))
	}

	if len(v.Docs) != 0 {
		for _, dps := range v.Docs {
			if dps.IsMember {
				continue
			}
			f.WriteString("---\n\n")
			f.WriteString(heading(dps.FuncName, 4))
			f.WriteString(fmt.Sprintf("%s.%s\n\n", mod, dps.FuncSig))

			for _, doc := range dps.Doc {
				if !strings.HasPrefix(doc, "---") && doc != "" {
					f.WriteString(doc + "  \n")
				}
			}
			f.WriteString("\n")
			f.WriteString(heading("Parameters", 4))
			if len(dps.Params) == 0 {
				f.WriteString("This function has no parameters.  \n")
			}
			for _, p := range dps.Params {
				isVariadic := false
				typ := p.Type
				if strings.HasPrefix(p.Type, "...") {
					isVariadic = true
					typ = p.Type[3:]
				}

				f.WriteString(fmt.Sprintf("`%s` _%s_", typ, p.Name))
				if isVariadic {
					f.WriteString(" (This type is variadic. You can pass an infinite amount of parameters with this type.)")
				}
				f.WriteString("  \n")
				f.WriteString(strings.Join(p.Doc, " "))
				f.WriteString("\n\n")
			}
			if codeExample := dps.Tags["example"]; codeExample != nil {
				f.WriteString(heading("Example", 4))
				f.WriteString(fmt.Sprintf("```lua\n%s\n```\n", strings.Join(codeExample[0].Fields, "\n")))
			}
			f.WriteString("\n\n")
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
			if len(dps.Properties) != 0 {
				f.WriteString(heading("Object Properties", 2))

				propertiesList := [][]string{}
				for _, p := range dps.Properties {
					propertiesList = append(propertiesList, []string{fmt.Sprintf("`%s`", p.FuncName), strings.Join(p.Doc, " ")})
				}
				f.WriteString(bulletList(propertiesList))
			}
			f.WriteString("\n")
			f.WriteString(heading("Methods", 3))
			for _, dps := range v.Docs {
				if !dps.IsMember {
					continue
				}
				htmlSig := typeTag.ReplaceAllStringFunc(strings.Replace(dps.FuncSig, "<", `\<`, -1), func(typ string) string {
					typName := regexp.MustCompile(`\w+`).FindString(typ[1:])
					typLookup := typeTable[strings.ToLower(typName)]
					fmt.Printf("%+q, \n", typLookup)
					linkedTyp := fmt.Sprintf("/Hilbish/docs/api/%s/%s/#%s", typLookup[0], typLookup[0]+"."+typLookup[1], strings.ToLower(typName))
					return fmt.Sprintf(`<a href="#%s" style="text-decoration: none;">%s</a>`, linkedTyp, typName)
				})
				f.WriteString(heading(htmlSig, 4))
				for _, doc := range dps.Doc {
					if !strings.HasPrefix(doc, "---") {
						f.WriteString(doc + "\n")
					}
				}
				f.WriteString("\n")
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
