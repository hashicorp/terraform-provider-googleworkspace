package googleworkspace

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/mitchellh/go-homedir"

	"google.golang.org/api/googleapi"
)

const CreateTimeout = 10 * time.Minute
const UpdateTimeout = 10 * time.Minute

// If the argument is a path, pathOrContents loads it and returns the contents,
// otherwise the argument is assumed to be the desired contents and is simply
// returned.
//
// The boolean second return value can be called `wasPath` - it indicates if a
// path was detected and a file loaded.
func pathOrContents(poc string) (string, bool, error) {
	if len(poc) == 0 {
		return poc, false, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, true, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), true, err
		}
		return string(contents), true, nil
	}

	return poc, false, nil
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}

func addCannotInterpolateInProviderBlockError(resp *tfsdk.ConfigureProviderResponse, attr string) {
	resp.Diagnostics.AddAttributeError(
		tftypes.NewAttributePath().WithAttributeName(attr),
		"Can't interpolate into provider block",
		"Interpolating that value into the provider block doesn't give the provider enough information to run. Try hard-coding the value, instead.",
	)
}

func addAttributeMustBeSetError(resp *tfsdk.ConfigureProviderResponse, attr string) {
	resp.Diagnostics.AddAttributeError(
		tftypes.NewAttributePath().WithAttributeName(attr),
		"Invalid provider config",
		fmt.Sprintf("%s must be set.", attr),
	)
}

// Check Error Code
func isApiErrorWithCode(err error, errCode int) bool {
	gerr, ok := errwrap.GetType(err, &googleapi.Error{}).(*googleapi.Error)
	return ok && gerr != nil && gerr.Code == errCode
}

func handleNotFoundError(err error, id string, diags *diag.Diagnostics) string {
	if isApiErrorWithCode(err, 404) {
		// The resource doesn't exist
		return ""
	}

	diags.AddError(fmt.Sprintf("Error when reading or editing %s", id), err.Error())
	return id
}

// This is a Printf sibling (Nprintf; Named Printf), which handles strings like
// Nprintf("Hello %{target}!", map[string]interface{}{"target":"world"}) == "Hello world!".
// This is particularly useful for generated tests, where we don't want to use Printf,
// since that would require us to generate a very particular ordering of arguments.
func Nprintf(format string, params map[string]interface{}) string {
	for key, val := range params {
		format = strings.Replace(format, "%{"+key+"}", fmt.Sprintf("%v", val), -1)
	}
	return format
}

// This will translate a snake cased string to a camel case string
// Note: the first letter of the camel case string will be lower case
func SnakeToCamel(s string) string {
	titled := strings.Title(strings.ReplaceAll(s, "_", " "))
	cameled := strings.Join(strings.Split(titled, " "), "")

	// Lower the first letter
	result := []rune(cameled)
	result[0] = unicode.ToLower(result[0])
	return string(result)
}

// This will translate a snake cased string to a camel case string
// Note: the first letter of the camel case string will be lower case
func CameltoSnake(s string) string {
	var res = make([]rune, 0, len(s))
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			res = append(res, '_', unicode.ToLower(r))
		} else {
			res = append(res, unicode.ToLower(r))
		}
	}
	return string(res)
}

// For resources that have many nested interfaces, we can pass them to the API as is,
// only each field name needs to be camel case rather than snake case. Additionally,
// fields that are not set should not be sent to the API.
func expandInterfaceObjects(parent interface{}) []interface{} {
	objList := parent.([]interface{})
	if len(objList) == 0 {
		return nil
	}

	newObjList := []interface{}{}

	for _, o := range objList {
		obj := o.(map[string]interface{})
		for k, v := range obj {
			if strings.Contains(k, "_") || v == "" {
				delete(obj, k)

				// In the case that the field is not set, don't send it to the API
				if v == "" {
					continue
				}

				obj[SnakeToCamel(k)] = v
			}
		}
		newObjList = append(newObjList, obj)
	}

	return newObjList
}

// For resources that have many nested interfaces, we can set was was returned from the API as is
// only the field names need to be snake case rather than the camel case that is returned
func flattenInterfaceObjects(objList interface{}) interface{} {
	if objList == nil || len(objList.([]interface{})) == 0 {
		return nil
	}

	newObjList := []map[string]interface{}{}

	for _, o := range objList.([]interface{}) {
		obj := o.(map[string]interface{})
		for k, v := range obj {
			delete(obj, k)
			obj[CameltoSnake(k)] = v
		}

		newObjList = append(newObjList, obj)
	}

	return newObjList
}

// Converts a list of interfaces to a list of strings
func listOfInterfacestoStrings(v interface{}) []string {
	result := []string{}

	if v == nil {
		return result
	}

	for _, s := range v.([]interface{}) {
		result = append(result, s.(string))
	}

	return result
}

func stringInSlice(arr []string, str string) bool {
	for _, i := range arr {
		if i == str {
			return true
		}
	}

	return false
}

// sort a slice of interfaces regardless the type, return the equivalent slice of strings
func sortListOfInterfaces(v []interface{}) []string {
	newVal := make([]string, len(v))
	for idx, attr := range v {
		kind := reflect.ValueOf(v).Kind()
		if kind == reflect.Float64 {
			attr = strconv.FormatFloat(attr.(float64), 'f', -1, 64)
		}
		newVal[idx] = fmt.Sprintf("%+v", attr)
	}
	sort.Strings(newVal)
	return newVal
}

// stringSliceToTypeList will change a slice of strings to the plugin-framework types.List
func stringSliceToTypeList(strs []string) types.List {
	result := types.List{
		ElemType: types.StringType,
	}

	for i, s := range strs {
		result.Elems[i] = types.String{Value: s}
	}

	return result
}

// typeListToSliceStrings will change a list of attr.Values to a list of strings
func typeListToSliceStrings(vals []attr.Value) []string {
	var result []string
	for _, v := range vals {
		result = append(result, v.(types.String).Value)
	}

	return result
}

// getDiagErrors prints errors in diags
func getDiagErrors(diags []diag.Diagnostic) error {
	var errs []string
	for _, d := range diags {
		errs = append(errs, fmt.Sprintf("%s : %s | %s", d.Severity(), d.Summary(), d.Detail()))
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}
