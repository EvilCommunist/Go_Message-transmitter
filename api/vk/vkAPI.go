package vk

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	ACCESS_TOKEN = "access_token"
	VERSION      = "v"
	RESPONSE     = "response"
	MESSAGE      = "message"
)

func ResendMessage(method string, parameters map[string]string,
	vkVersion string, vkEndpoint string, vkToken string) {
	parameters[ACCESS_TOKEN] = vkToken
	parameters[VERSION] = vkVersion

	query := buildQuery(parameters)
	reqUrl := fmt.Sprintf("%s%s?%s", vkEndpoint, method, query)

	_, err := http.Get(reqUrl)
	if err != nil {
		//		fmt.Printf("The error acquired while sending message %s to bot: %s", parameters[MESSAGE], err)
		return
	}
	// Resender works as a daemon without logger so any body check will give nothing
}

func buildQuery(paramMap map[string]string) string {
	var query string = ""
	for key, value := range paramMap {
		query += fmt.Sprintf("%s=%s&", key, value)
	}

	return strings.TrimRight(query, "&")
}
