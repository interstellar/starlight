package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

var fset = new(token.FileSet)

func gen(name string) (code []byte, err error) {
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg, name)
	if err != nil {
		return nil, err
	}

	if len(pkgs[0].Syntax) != 1 {
		// TODO(kr): remove this limitation
		return nil, errors.New("genbolt: schema must be a single file")
	}

	ctx := &context{
		pkg: pkgs[0],
		sch: &schema{
			Imports:          make(map[string]string),
			Keys:             make(map[string]bool),
			MapOfBucketTypes: make(map[string]string),
			SeqOfBucketTypes: make(map[string]string),
			MapOfRecordTypes: make(map[string]types.Type),
			SeqOfRecordTypes: make(map[string]types.Type),
			funcs:            make(template.FuncMap),
		},
		jsonTypes: make(map[*types.Named]bool),
		binTypes:  make(map[*types.Named]bool),
	}
	err = genFile(ctx, pkgs[0].Syntax[0])
	if err != nil {
		return nil, err
	}
	ctx.sch.InputFile = name

	var b bytes.Buffer
	tmpl, err := template.New("").
		Funcs(ctx.sch.funcs).
		Funcs(template.FuncMap{
			"trimprefix": strings.TrimPrefix,
			"identical":  types.Identical,
			"basic":      basicType,
			"isslice":    isSlice,
			"ispointer":  isPointer,
			"sizeof":     types.SizesFor("gc", "amd64").Sizeof,
		}).
		Parse(schemaTemplate)
	if err != nil {
		return nil, err
	}
	err = tmpl.Execute(&b, ctx.sch)
	if err != nil {
		return nil, err
	}
	code, err = format.Source(b.Bytes())
	if err != nil {
		return b.Bytes(), err
	}
	return code, nil
}

type context struct {
	sch *schema
	pkg *packages.Package

	jsonTypes map[*types.Named]bool
	binTypes  map[*types.Named]bool
}

func genFile(ctx *context, file *ast.File) error {
	sch := ctx.sch
	typesInfo := ctx.pkg.TypesInfo

	sch.Package = file.Name.Name

	sch.funcs["typestring"] = func(t types.Type) string {
		return types.TypeString(t, func(p *types.Package) string {
			return sch.Imports[p.Path()]
		})
	}
	sch.funcs["isjsontype"] = func(v interface{}) bool {
		p, ok := v.(*types.Pointer)
		if !ok {
			return false
		}
		t, _ := p.Elem().(*types.Named)
		return ctx.jsonTypes[t]
	}
	sch.funcs["isbintype"] = func(v interface{}) bool {
		p, ok := v.(*types.Pointer)
		if !ok {
			return false
		}
		t, _ := p.Elem().(*types.Named)
		return ctx.binTypes[t]
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs := spec.(*ast.ValueSpec)
			iface := typesInfo.Types[vs.Type].Type
			if iface == nil {
				return fmt.Errorf("interface assertion has no interface: %v", esprint(vs))
			}

			var typeMap map[*types.Named]bool
			switch iface.String() {
			case "encoding/json.Marshaler":
				typeMap = ctx.jsonTypes
			case "encoding.BinaryMarshaler":
				typeMap = ctx.binTypes
			default:
				return fmt.Errorf("unsupported interface: %v", iface)
			}

			for i, value := range vs.Values {
				if vs.Names[i].Name != "_" {
					return fmt.Errorf("interface assertion has non-_ name: %v", esprint(vs))
				}
				convType := typesInfo.Types[value].Type
				ptr, ok := convType.(*types.Pointer)
				if !ok {
					return fmt.Errorf("interface assertion has bad expression (must be pointer to named type): %v", esprint(convType))
				}
				named, ok := ptr.Elem().(*types.Named)
				if !ok {
					return fmt.Errorf("interface assertion has bad expression (must be pointer to named type): %v", esprint(convType))
				}
				typeMap[named] = true
			}
		}
	}

	for _, imp := range ctx.pkg.Types.Imports() {
		sch.Imports[imp.Path()] = imp.Name()
	}
	for _, imp := range file.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		if imp.Name != nil {
			sch.Imports[path] = imp.Name.Name
		}
	}

	if len(ctx.jsonTypes) > 0 {
		sch.Imports["encoding/json"] = "json"
	}
	if len(ctx.binTypes) > 0 {
		sch.Imports["encoding"] = "encoding"
	}
	sch.Imports["bytes"] = "bytes"
	sch.Imports["encoding/binary"] = "binary"
	sch.Imports["github.com/coreos/bbolt"] = "bolt"

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			return fmt.Errorf("unexpected decl: %v", esprint(decl))
		}
		switch genDecl.Tok {
		default:
			return fmt.Errorf("unexpected decl: %v", esprint(decl))
		case token.VAR, token.IMPORT:
			continue
		case token.TYPE: // ok, proceed
		}
		for _, spec := range genDecl.Specs {
			spec := spec.(*ast.TypeSpec)
			if spec.Assign != 0 {
				return fmt.Errorf("unexpected decl: %v", esprint(decl))
			}
			doc := spec.Doc
			if doc == nil {
				doc = genDecl.Doc
			}
			err := genStruct(ctx, spec.Name.Name, spec.Type, doc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func genStruct(ctx *context, name string, typ ast.Expr, doc *ast.CommentGroup) error {
	sch := ctx.sch

	structType, ok := typ.(*ast.StructType)
	if !ok {
		return fmt.Errorf("need struct type")
	}

	isRoot := false
	if strings.HasPrefix(name, "Root") {
		isRoot = name == "Root" || ast.IsExported(name[4:])
	}

	sch.StructTypes = append(sch.StructTypes, &schemaStruct{
		Name:   name,
		IsRoot: isRoot,
		Doc:    doc,
	})

	for _, field := range structType.Fields.List {
		for _, fieldIdent := range field.Names {
			if !fieldIdent.IsExported() {
				return fmt.Errorf("all fields must be exported")
			}
			if isReserved(fieldIdent.Name) {
				return fmt.Errorf("field name %s is reserved (sorry)", fieldIdent)
			}
			sch.Keys[fieldIdent.Name] = true

			fieldType := ctx.pkg.TypesInfo.Defs[fieldIdent].Type()
			dbType, isRec, err := genType(ctx, fieldType)
			if err != nil {
				return err
			}

			if isRoot && isRec {
				return fmt.Errorf("unsupported root field type %v", fieldType)
			}

			l := &sch.BucketFields
			if isRec {
				l = &sch.RecordFields
			}
			*l = append(*l, &schemaField{
				Name:   fieldIdent.Name,
				Type:   dbType,
				Bucket: name,
				Doc:    field.Doc,
			})
		}
	}
	return nil
}

// genType does two things:
//   1. it adds an entry to one of the MapOfXTypes or
//      SeqOfXTypes, if necessary
//   2. it returns the database type for the field
//      (which may or may not be the same as fieldType)
func genType(ctx *context, fieldType types.Type) (dbType types.Type, isRec bool, err error) {
	switch fieldType := fieldType.(type) {
	default:
		return nil, false, fmt.Errorf("type %v unsupported", fieldType)
	case *types.Named:
		return nil, false, fmt.Errorf("type %v unsupported (try %v instead?)", fieldType, types.NewPointer(fieldType))
	case *types.Array:
		return nil, false, fmt.Errorf("type %v unsupported (try %v instead?)", fieldType, types.NewSlice(fieldType.Elem()))
	case *types.Basic:
		if !isSupportedBasic(fieldType) {
			return nil, false, fmt.Errorf("type %v unsupported", fieldType)
		}
		return fieldType, true, nil
	case *types.Pointer:
		named, ok := fieldType.Elem().(*types.Named)
		if !ok {
			return nil, false, fmt.Errorf("unknown type %v", fieldType)
		}

		isRec = isRecordType(ctx, named)
		if !isRec && !isBucketType(ctx, named) {
			return nil, false, fmt.Errorf("unknown type %v", fieldType)
		}

		return fieldType, isRec, nil
	case *types.Slice:
		// Special case for []byte, []int64, etc. We store
		// slices of fixed-size basic types as records.
		if t, ok := fieldType.Elem().(*types.Basic); ok && isFixedSize(t) {
			return fieldType, true, nil
		}

		dbType, err = genContainer(ctx, "SeqOf", fieldType.Elem(),
			ctx.sch.SeqOfRecordTypes,
			ctx.sch.SeqOfBucketTypes,
		)
		return dbType, false, err
	case *types.Map:
		// TODO(kr): allow numeric types as map keys too
		keyType, ok := fieldType.Key().(*types.Basic)
		if !ok || keyType.Kind() != types.String {
			return nil, false, fmt.Errorf("map key must be string")
		}

		dbType, err = genContainer(ctx, "MapOf", fieldType.Elem(),
			ctx.sch.MapOfRecordTypes,
			ctx.sch.MapOfBucketTypes,
		)
		return dbType, false, err
	}
}

func genContainer(ctx *context, prefix string, schemaElemType types.Type, recordTypes map[string]types.Type, bucketTypes map[string]string) (types.Type, error) {
	elemType, isRec, err := genType(ctx, schemaElemType)
	if err != nil {
		return nil, err
	}

	elemName, err := typeDescIdent(ctx.pkg.Types, elemType)
	if err != nil {
		return nil, err
	}

	typeName := prefix + elemName
	if isRec {
		recordTypes[typeName] = elemType
	} else {
		bucketTypes[typeName] = elemName
	}

	dbType := types.NewPointer(types.NewNamed(
		types.NewTypeName(0, ctx.pkg.Types, typeName, nil),
		types.Typ[types.Invalid],
		nil,
	))
	return dbType, nil
}

func isBucketType(ctx *context, t *types.Named) bool {
	_, ok := t.Underlying().(*types.Struct)
	return ok && t.Obj().Pkg().Path() == ctx.pkg.Types.Path()
}

func isRecordType(ctx *context, t *types.Named) bool {
	_, okJSON := ctx.jsonTypes[t]
	_, okBin := ctx.binTypes[t]
	return okJSON || okBin
}

func esprint(node interface{}) string {
	var b bytes.Buffer
	printer.Fprint(&b, fset, node)
	return b.String()
}

func isSupportedBasic(t *types.Basic) bool {
	return isFixedSize(t) || t.Kind() == types.String
}

func isFixedSize(t *types.Basic) bool {
	switch t.Kind() {
	case types.Bool,
		types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return true
	}
	return false
}

func isReserved(name string) bool {
	switch name {
	case "Tx", "Bucket":
		return true
	}
	return false
}

func isPointer(t types.Type) bool {
	_, ok := t.(*types.Pointer)
	return ok
}

func isSlice(t types.Type) bool {
	_, ok := t.(*types.Slice)
	return ok
}

// typeDescIdent returns an exported name suitable for describing t in p.
// If t is a pointer to a named type already in p, it returns t's name.
// If t is a pointer to a named type in another package, it combines
// the package name, capitalized, with t's unqualified name.
// If t is a basic type, it returns t's name capitalized.
// If t is a slice of a basic type, it returns "SliceOf"
// concatenated with t's name capitalized.
// Other types are an error.
func typeDescIdent(p *types.Package, t types.Type) (string, error) {
	switch t := t.(type) {
	case *types.Basic:
		return exported(t.Name()), nil
	case *types.Pointer:
		named, ok := t.Elem().(*types.Named)
		if !ok {
			break
		}
		s := exported(named.Obj().Name())
		if named.Obj().Pkg().Path() == p.Path() {
			return s, nil
		} else {
			return exported(named.Obj().Pkg().Name()) + s, nil
		}
	case *types.Slice:
		basic, ok := t.Elem().(*types.Basic)
		if !ok {
			break
		}
		return "SliceOf" + exported(basic.Name()), nil
	}
	return "", fmt.Errorf("cannot convert %v to a local name", t)
}

func exported(name string) string {
	if ast.IsExported(name) {
		return name
	}
	ru, n := utf8.DecodeRuneInString(name)
	return string(unicode.ToUpper(ru)) + name[n:]
}

var basicTypes = make(map[string]*types.Basic)

// basicType returns the named basic type.
func basicType(name string) *types.Basic {
	return basicTypes[name]
}

func init() {
	for _, t := range types.Typ {
		basicTypes[t.Name()] = t
	}
}
