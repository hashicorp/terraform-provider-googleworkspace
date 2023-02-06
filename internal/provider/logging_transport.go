// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

func getValuesToScrub() []string {
	return []string{
		"accessToken",
	}
}

type loggingTransport struct {
	name      string
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if logging.IsDebugOrHigher() {
		reqData, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			prettyPrint, err := prettyPrintJsonLines(reqData)
			if err != nil {
				return nil, err
			}
			log.Printf("[DEBUG] "+logReqMsg, t.name, prettyPrint)
		} else {
			log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if logging.IsDebugOrHigher() {
		respData, err := httputil.DumpResponse(resp, true)
		if err == nil {
			prettyPrint, err := prettyPrintJsonLines(respData)
			if err != nil {
				return nil, err
			}
			log.Printf("[DEBUG] "+logRespMsg, t.name, prettyPrint)
		} else {
			log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
		}
	}

	return resp, nil
}

func NewTransportWithScrubbedLogs(name string, t http.RoundTripper) *loggingTransport {
	return &loggingTransport{name, t}
}

// prettyPrintJsonLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json.
// this was copied from the SDK's logging package
func prettyPrintJsonLines(b []byte) (string, error) {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer

			var jsonMap map[string]interface{}
			err := json.Unmarshal(b, &jsonMap)
			if err != nil {
				return "", err
			}

			jsonMap = obfuscateValues(jsonMap)

			b, err = json.Marshal(jsonMap)
			if err != nil {
				return "", err
			}

			json.Indent(&out, b, "", " ")
			parts[i] = out.String()
		}
	}
	return strings.Join(parts, "\n"), nil
}

func obfuscateValues(m map[string]interface{}) map[string]interface{} {
	for _, v := range getValuesToScrub() {
		if _, ok := m[v]; ok {
			m[v] = "********"
		}
	}

	return m
}

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`
