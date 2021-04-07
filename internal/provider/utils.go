package googleworkspace

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mitchellh/go-homedir"

	"google.golang.org/api/googleapi"
)

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

// Check Error Code
func isApiErrorWithCode(err error, errCode int) bool {
	gerr, ok := errwrap.GetType(err, &googleapi.Error{}).(*googleapi.Error)
	return ok && gerr != nil && gerr.Code == errCode
}

func handleNotFoundError(err error, d *schema.ResourceData, resource string) diag.Diagnostics {
	if isApiErrorWithCode(err, 404) {
		log.Printf("[WARN] Removing %s because it's gone", resource)
		// The resource doesn't exist anymore
		d.SetId("")

		return nil
	}

	return diag.Errorf("Error when reading or editing %s: %s", resource, err.Error())
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
// only each field name needs to be camel case rather than snake case.
func expandInterfaceObjects(parent interface{}) []interface{} {
	objList := parent.([]interface{})
	if objList == nil || len(objList) == 0 {
		return nil
	}

	newObjList := []interface{}{}

	for _, o := range objList {
		obj := o.(map[string]interface{})
		for k, v := range obj {
			if strings.Contains(k, "_") {
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

// User type has many nested interfaces, we can set was was returned from the API as is
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
