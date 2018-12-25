package base

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/microsvs/base/pkg/rpc"
)

// 针对graphql Enum, Scalar{DateTime,...} 转化为Go支持的String, time.Time , *time.Time类型
// 主要用于rpc调用
/* example
var t time.Time= time.Now()
FixTypeFromGoToGraphql(t, graphql.DateTime).(string)//2006-01-02 15:04:05 -> "2006-01-02 15:04:05"
FixTypeFromGoToGraphql(fency.ParkZone, fency.GLPOIType) // ParkZone ---> "parkzone"
*/
func FixTypeFromGoToGraphql(v interface{}, argType graphql.Input) (result interface{}) {
	var ret interface{}
	switch val := argType.(type) {
	case *graphql.List:
		ret = buildSlice(v.([]interface{}), val.OfType)
	case *graphql.Scalar:
		ret = val.Serialize(v)
	case *graphql.Enum:
		return val.Serialize(v)
	case *graphql.NonNull:
		return FixTypeFromGoToGraphql(v, val.OfType)
	}
	if _, ok := v.(string); ok {
		return fmt.Sprintf("\"%s\"", ret)
	}
	return ret
}

func buildSlice(items []interface{}, typ graphql.Type) (values []interface{}) {
	for _, item := range items {
		if enum, eok := typ.(*graphql.Enum); eok {
			//convert int value to string value
			values = append(values, enum.Serialize(item))
		} else if v, ok := item.(string); ok {
			values = append(values, fmt.Sprintf("\"%s\"", v))
		} else {
			values = append(values, item)
		}
	}
	return values
}

// 针对graphql内部能够处理的类型, 转化为golang内部的枚举类型、[]byte、time.Time、*time.Time等类型
// FixTypeFromGraphqlToGo与上面的FixTypeFromGoToGraphql作用相反
/* example
FixTypeFromGraphqlToGo("2006-01-02 15:04:05", graphql.DateTime) // return time.Time
FixTypeFromGraphqlToGo("parkzone", graphql.Enum) // return ParkZone
*/
func FixTypeFromGraphqlToGo(data interface{}, t graphql.Output) interface{} {
	switch val := t.(type) {
	case *graphql.List:
		for idx, v := range data.([]interface{}) {
			data.([]interface{})[idx] = FixTypeFromGraphqlToGo(v, val.OfType)
		}
	case *graphql.NonNull:
		data = FixTypeFromGraphqlToGo(data, val.OfType)
	case *graphql.Enum:
		data = val.ParseValue(data)
	case *graphql.Scalar:
		data = val.ParseValue(data)
	case *graphql.Object:
		imap, _ := data.(map[string]interface{})
		fmap := val.Fields()
		for key, value := range imap {
			if _, ok := fmap[key]; !ok {
				continue
			}
			imap[key] = FixTypeFromGraphqlToGo(value, fmap[key].Type)
		}
		data = imap
	default:
	}
	return data
}

// graphql request method = {query, mutation}
func GetSchemeMethodName(p graphql.ResolveParams) (method string) {
	if len(p.Info.FieldASTs) > 0 && p.Info.FieldASTs[0].Name != nil {
		return p.Info.FieldASTs[0].Name.Value
	}
	return
}

func getSchemeArgs(
	p graphql.ResolveParams,
	exArgs map[string]interface{},
	excommon map[string]graphql.Input,
) (params string, output graphql.Output) {
	if len(p.Info.FieldASTs) <= 0 || p.Info.FieldASTs[0].Name == nil {
		return
	}
	var defArgs *graphql.FieldDefinition
	// find input field definition
	switch p.Info.Operation.GetOperation() {
	case ast.OperationTypeQuery:
		defArgs = p.Info.Schema.QueryType().Fields()[p.Info.FieldASTs[0].Name.Value]
	case ast.OperationTypeMutation:
		defArgs = p.Info.Schema.MutationType().Fields()[p.Info.FieldASTs[0].Name.Value]
	}
	output = defArgs.Type
	// build args: a: a-value, b: b-value, ...
	for i, arg := range defArgs.Args {
		if _, ok := p.Args[arg.PrivateName]; ok {
			if i == 0 {
				params = fmt.Sprintf("%s:%v", arg.PrivateName,
					FixTypeFromGoToGraphql(p.Args[arg.PrivateName], arg.Type))
			} else {
				params = fmt.Sprintf("%s, %s:%v", params, arg.PrivateName,
					FixTypeFromGoToGraphql(p.Args[arg.PrivateName], defArgs.Args[i].Type))
			}
		}
	}
	// expand
	for arg, value := range exArgs {
		if len(params) > 0 {
			params = fmt.Sprintf("%s, %s:%v", params, arg, FixTypeFromGoToGraphql(value, excommon[arg]))
		} else {
			params = fmt.Sprintf("%s:%v", arg, FixTypeFromGoToGraphql(value, excommon[arg]))
		}
	}
	if len(params) > 0 {
		params = fmt.Sprintf("(%s)", params)
	}
	return
}

// 获取返回参数列表
func getSchemaSelections(p graphql.ResolveParams) string {
	var ss *ast.SelectionSet
	if ss = p.Info.FieldASTs[0].GetSelectionSet(); ss == nil {
		return ""
	}
	return reverseSelection(ss.Selections)
}

func reverseSelection(sset []ast.Selection) string {
	var result []string
	for _, item := range sset {
		switch f := item.(type) {
		case *ast.Field:
			if f.Name != nil {
				result = append(result, f.Name.Value)
			}
		}
		temp := item.GetSelectionSet()
		if temp != nil && len(temp.Selections) > 0 {
			if str := reverseSelection(temp.Selections); len(str) > 0 {
				result = append(result, str)
			}
		}
	}
	if len(result) > 0 {
		return fmt.Sprintf("{%s}", strings.Join(result, " "))
	}
	return ""
}

func RedirectRequestEx(
	p graphql.ResolveParams,
	exArgs map[string]interface{},
	excommon map[string]graphql.Input,
	targetService rpc.FGService,
	targetObj interface{}) (interface{}, error) {
	var (
		params     string // 参数列表
		selections string // 返回列表
		method     string //方法名
		output     graphql.Output
	)
	// 看看自己对schema的AST了解有多深
	// get method name
	method = GetSchemeMethodName(p)
	// build args
	params, output = getSchemeArgs(p, exArgs, excommon)
	// build selections
	selections = getSchemaSelections(p)

	// composite schema
	schema := fmt.Sprintf("%s{%s%s%s}",
		p.Info.Operation.GetOperation(),
		method,
		params,
		selections,
	)
	// rpc call service
	data, err := rpc.CallService(p.Context, Service2Url(targetService), schema)
	if err != nil {
		return nil, err
	}
	// parse result
	return FixTypeFromGraphqlToGo(data[method], output), nil
}

//RedirectRequest redirect request from one resolve function to another service
func RedirectRequest(p graphql.ResolveParams, target rpc.FGService) (interface{}, error) {
	return RedirectRequestEx(p, nil, nil, target, nil)
}

// GLObjectFields获取graphql.Object的所有字段, 作为graphql Query/Mutation的返回值
func GLObjectFields(obj *graphql.Object) string {
	keys := make([]string, 0, len(obj.Fields()))
	for key, v := range obj.Fields() {
		switch typ := v.Type.(type) {
		case *graphql.Object:
			key += GLObjectFields(typ)
		}
		keys = append(keys, key)
	}
	return "{" + strings.Join(keys, "\n") + "}"
}

// 隐藏某些不愿意暴露的字段
func HideGLFields(obj *graphql.Object, v ...string) *graphql.Object {
	for _, item := range v {
		delete(obj.Fields(), item)
	}
	return obj
}
