package httpreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type RequestParam struct {
	EndPoint    string                 `json:"endPoint"`
	ArgsMap     map[string]interface{} `json:"args"`
	QueryParams map[string]string      `json:"queryParams"`
	Method      string                 `json:"method"`
	Token       string                 `json:"token"`
	Headers     map[string]string      `json:"headers"`
	Args        interface{}            // when we don't need key-value
}

func Do(hreq RequestParam, resp interface{}) (*http.Response, error) {
	method := "POST"
	if len(hreq.Method) > 0 {
		method = hreq.Method
	}

	var requestBody bytes.Buffer

	if hreq.ArgsMap != nil && hreq.Args != nil {
		return nil, errors.New("args and argsmap cannot exist at the same time")
	}

	if hreq.ArgsMap != nil {
		err := json.NewEncoder(&requestBody).Encode(hreq.ArgsMap)
		if err != nil {
			return nil, errors.Wrap(err, "encode body")
		}
	}

	if hreq.Args != nil {
		err := json.NewEncoder(&requestBody).Encode(hreq.Args)
		if err != nil {
			return nil, errors.Wrap(err, "encode body")
		}
	}

	params := url.Values{}
	// params.Add("query", "keyword")
	for qparamKey, qparamVal := range hreq.QueryParams {
		params.Add(qparamKey, qparamVal)
	}

	paramEncoded := params.Encode()
	if len(paramEncoded) > 0 {
		hreq.EndPoint += "?" + paramEncoded
	}

	fmt.Println(hreq.EndPoint)

	client := &http.Client{}
	req, err := http.NewRequest(method, hreq.EndPoint, &requestBody)

	if err != nil {
		return nil, err
	}

	if len(hreq.Token) > 0 {
		req.Header.Add("Authorization", "Bearer "+hreq.Token)

	}
	req.Header.Add("Content-Type", "application/json")

	for headerKey, headerValue := range hreq.Headers {
		req.Header.Add(headerKey, headerValue)
		// fmt.Println(req.Header)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 299 && body != nil {
		return nil, fmt.Errorf("status: %d | error:%s", res.StatusCode, string(body))
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("status: %d", res.StatusCode)
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return res, nil
}
