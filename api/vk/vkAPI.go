package vk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	ACCESS_TOKEN = "access_token"
	VERSION      = "v"
	RESPONSE     = "response"
	MESSAGE      = "message"
)

func ResendMessageVK(method string, parameters map[string]string,
	vkVersion string, vkEndpoint string, vkToken string) {
	parameters[ACCESS_TOKEN] = vkToken
	parameters[VERSION] = vkVersion

	query := buildQuery(parameters)
	reqUrl := vkEndpoint + method + "?" + query

	response, err := http.Get(reqUrl)
	if err != nil {
		//		fmt.Printf("The error acquired while sending message %s to bot: %s", parameters[MESSAGE], err)
		return
	}
	defer response.Body.Close()

	body, errorResp := io.ReadAll(response.Body)
	if errorResp != nil {
		//		fmt.Println("Error acquired while executing VK API method!")
		//		fmt.Println(errorResp)
		return
	}

	bodyStr := string(body)
	bodyStr, _ = url.QueryUnescape(bodyStr)
	var responseData = make(map[string]string)
	json.Unmarshal([]byte(bodyStr), &responseData)

	if responseData[RESPONSE] == "" {
		return
	}
}

func buildQuery(paramMap map[string]string) string {
	var query string = ""
	for key, value := range paramMap {
		query += key + "=" + value + "&"
	}

	return strings.TrimRight(query, "&")
}
