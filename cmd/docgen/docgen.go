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
	"strings"
	//"regexp"
	//"sync"
)

var header = `---
title: %s %s
description: %s
layout: doc
menu:
  docs:
    parent: "API"
---

`

type emmyPiece struct {
	DocPiece    *docPiece
	Annotations []string
	Params      []string // we only need to know param name to put in function
	FuncName    string
}

type module struct {
	Name             string            `json:"name"`
	ShortDescription string            `json:"shortDescription"`
	Description      string            `json:"description"`
	ParentModule     string            `json:"parent,omitempty"`
	HasInterfaces    bool              `json:"-"`
	HasTypes         bool              `json:"-"`
	Properties       []docPiece        `json:"properties"`
	Fields           []docPiece        `json:"fields"`
	Types            []docPiece        `json:"types,omitempty"`
	Docs             []docPiece        `json:"docs"`
	Interfaces       map[string]module `json:"interfaces,omitempty"`
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
	if len(os.Args) == 1 {
		fset := token.NewFileSet()
		os.Mkdir("defs", 0777)
		/*
			os.Mkdir("emmyLuaDocs", 0777)
		*/

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

		interfaceModules := make(map[string]*module)
		for l, f := range pkgs {
			p := doc.New(f, "./", doc.AllDecls)
			pieces := []docPiece{}
			typePieces := []docPiece{}
			mod := l
			if mod == "main" || mod == "util" {
				mod = "hilbish"
			}
			var hasInterfaces bool
			for _, t := range p.Funcs {
				piece := setupDoc(mod, t)
				if piece == nil {
					continue
				}

				pieces = append(pieces, *piece)
				if piece.IsInterface {
					hasInterfaces = true
				}
			}
			for _, t := range p.Types {
				typePiece := setupDocType(mod, t)
				if typePiece != nil {
					typePieces = append(typePieces, *typePiece)
					if typePiece.IsInterface {
						hasInterfaces = true
					}
				}

				for _, m := range t.Methods {
					piece := setupDoc(mod, m)
					if piece == nil {
						continue
					}

					pieces = append(pieces, *piece)
					if piece.IsInterface {
						hasInterfaces = true
					}
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
						ParentModule: piece.ParentModule,
					}
				}

				if strings.HasSuffix(piece.GoFuncName, strings.ToLower("loader")) {
					shortDesc := piece.Doc[0]
					desc := piece.Doc[1:]
					interfaceModules[modname].ShortDescription = shortDesc
					interfaceModules[modname].Description = strings.Replace(strings.Join(desc, "\n"), "<nl>", "\\\n \\", -1)
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
						ParentModule: piece.ParentModule,
					}
				}

				interfaceModules[modname].Types = append(interfaceModules[modname].Types, piece)
			}

			fmt.Println(filteredTypePieces)
			if newDoc, ok := docs[mod]; ok {
				oldMod := docs[mod]
				newDoc.Types = append(filteredTypePieces, oldMod.Types...)
				newDoc.Docs = append(filteredPieces, oldMod.Docs...)

				docs[mod] = newDoc
			} else {
				docs[mod] = module{
					Name:             mod,
					Types:            filteredTypePieces,
					Docs:             filteredPieces,
					ShortDescription: shortDesc,
					Description:      strings.Replace(strings.Join(desc, "\n"), "<nl>", "\\\n \\", -1),
					HasInterfaces:    hasInterfaces,
					Properties:       docPieceTag("property", tags),
					Fields:           docPieceTag("field", tags),
					Interfaces:       make(map[string]module),
				}
			}
		}

		for key, mod := range interfaceModules {
			fmt.Println(key, mod.ParentModule)
			parentMod := docs[mod.ParentModule]
			parentMod.Interfaces[key] = *mod
		}

		//var wg sync.WaitGroup
		//wg.Add(len(docs) * 2)

		for mod, v := range docs {
			u, err := json.MarshalIndent(v, "", "	")
			if err != nil {
				panic(err)
			}

			fmt.Println(mod)
			f, err := os.Create("defs/" + mod + ".json")
			if err != nil {
				panic(err)
			}
			f.WriteString(string(u))
		}
		//wg.Wait()
	} else if os.Args[1] == "md" {
		fmt.Println("Generating MD files from Hilbish doc defs!")
		os.Mkdir("docs", 0777)
		os.RemoveAll("docs/api")
		os.Mkdir("docs/api", 0777)

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
			defContent, err := os.ReadFile(filepath.Join(".", "defs", defEntry.Name()))
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
}

func generateFile(v module) {
	mod := v.Name
	docPath := "docs/api/" + mod + ".md"
	if v.HasInterfaces {
		os.Mkdir("docs/api/"+mod, 0777)
		os.Remove(docPath) // remove old doc path if it exists
		docPath = "docs/api/" + mod + "/_index.md"
	}
	if v.ParentModule != "" {
		docPath = "docs/api/" + v.ParentModule + "/" + mod + ".md"
	}

	modOrIface := "Module"
	if v.ParentModule != "" {
		modOrIface = "Module"
	}
	lastHeader := ""

	f, _ := os.Create(docPath)
	f.WriteString(fmt.Sprintf(header, modOrIface, mod, v.ShortDescription))
	typeTag, _ := regexp.Compile(`\B@\w+`)
	/*modDescription := typeTag.ReplaceAllStringFunc(strings.Replace(strings.Replace(v.Description, "<", `\<`, -1), "{{\\<", "{{<", -1), func(typ string) string {
		typName := typ[1:]
		typLookup := typeTable[strings.ToLower(typName)]
		ifaces := typLookup[0] + "." + typLookup[1] + "/"
		if typLookup[1] == "" {
			ifaces = ""
		}
		linkedTyp := fmt.Sprintf("/Hilbish/docs/api/%s/%s#%s", typLookup[0], ifaces, strings.ToLower(typName))
		return fmt.Sprintf(`<a href="%s" style="text-decoration: none;">%s</a>`, linkedTyp, typName)
	})*/
	modDescription := v.Description
	f.WriteString(heading("Introduction", 2))
	f.WriteString(modDescription)
	f.WriteString("\n\n")
	if len(v.Docs) != 0 {
		funcCount := 0
		for _, dps := range v.Docs {
			if dps.IsMember {
				continue
			}
			funcCount++
		}

		f.WriteString(heading("Functions", 2))
		//lastHeader = "functions"

		diff := 0
		funcTable := [][]string{}
		for _, dps := range v.Docs {
			if dps.IsMember {
				diff++
				continue
			}

			if len(dps.Doc) == 0 {
				fmt.Printf("WARNING! Function %s on module %s has no documentation!\n", dps.FuncName, mod)
			} else {
				funcTable = append(funcTable, []string{fmt.Sprintf(`<a href="#%s">%s</a>`, dps.FuncName, dps.FuncSig), dps.Doc[0]})
			}
		}
		f.WriteString(table(funcTable))
	}

	if len(v.Fields) != 0 {
		f.WriteString(heading("Static module fields", 2))

		fieldsTable := [][]string{}
		for _, dps := range v.Fields {
			fieldsTable = append(fieldsTable, []string{dps.FuncName, strings.Join(dps.Doc, " ")})
		}
		f.WriteString(table(fieldsTable))
	}
	if len(v.Properties) != 0 {
		f.WriteString(heading("Object properties", 2))

		propertiesTable := [][]string{}
		for _, dps := range v.Properties {
			propertiesTable = append(propertiesTable, []string{dps.FuncName, strings.Join(dps.Doc, " ")})
		}
		f.WriteString(table(propertiesTable))
	}

	if len(v.Docs) != 0 {
		if lastHeader != "functions" {
			f.WriteString(heading("Functions", 2))
		}
		for _, dps := range v.Docs {
			if dps.IsMember {
				continue
			}
			f.WriteString("``` =html\n")
			f.WriteString(fmt.Sprintf("<hr class='my-4 text-neutral-400 dark:text-neutral-600'>\n<div id='%s'>", dps.FuncName))
			htmlSig := strings.Replace(mod+"."+dps.FuncSig, "<", `\<`, -1) /*typeTag.ReplaceAllStringFunc(strings.Replace(mod+"."+dps.FuncSig, "<", `\<`, -1), func(typ string) string {
				typName := typ[1:]
				typLookup := typeTable[strings.ToLower(typName)]
				ifaces := typLookup[0] + "." + typLookup[1] + "/"
				if typLookup[1] == "" {
					ifaces = ""
				}
				linkedTyp := fmt.Sprintf("/Hilbish/docs/api/%s/%s#%s", typLookup[0], ifaces, strings.ToLower(typName))
				return fmt.Sprintf(`<a href="%s" style="text-decoration: none;" id="lol">%s</a>`, linkedTyp, typName)
			})*/
			f.WriteString(fmt.Sprintf(`
<h4 class='text-xl font-medium mb-2'>
%s
<a href="#%s" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>
</div>

`, htmlSig, dps.FuncName))
			f.WriteString("```\n\n")

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
			f.WriteString("``` =html\n<hr class='my-4 text-neutral-400 dark:text-neutral-600'>\n```\n\n")
			f.WriteString(heading(dps.FuncName, 2))
			for _, doc := range dps.Doc {
				if !strings.HasPrefix(doc, "---") {
					f.WriteString(doc + "\n")
				}
			}
			if len(dps.Properties) != 0 {
				f.WriteString(heading("Object Properties", 2))

				propertiesTable := [][]string{}
				for _, dps := range v.Properties {
					propertiesTable = append(propertiesTable, []string{dps.FuncName, strings.Join(dps.Doc, " ")})
				}
				f.WriteString(table(propertiesTable))
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
				//f.WriteString(fmt.Sprintf("#### %s\n", htmlSig))
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

func table(elems [][]string) string {
	var b strings.Builder
	b.WriteString("``` =html\n")
	b.WriteString("<div class='relative overflow-x-auto sm:rounded-lg my-4'>\n")
	b.WriteString("<table class='w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400'>\n")
	b.WriteString("<tbody>\n")
	for _, line := range elems {
		b.WriteString("<tr class='bg-white border-b dark:bg-neutral-800 dark:border-neutral-700 border-neutral-200'>\n")
		for _, col := range line {
			b.WriteString("<td class='p-3 font-medium text-black dark:text-white'>")
			b.WriteString(col)
			b.WriteString("</td>\n")
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</tbody>\n")
	b.WriteString("</table>\n")
	b.WriteString("</div>\n")
	b.WriteString("```\n\n")

	return b.String()
}
