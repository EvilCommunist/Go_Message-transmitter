package max

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	CHAT        = "chat_id"
	AUTH        = "Authorization"
	TYPE        = "Content-Type"
	MSG_TYPE    = "application/json"
	MESSAGE     = "text"
	SEND_METHOD = "POST"
)

func ResendMessage(method string, message string, maxEndpoint string,
	maxToken string, maxChatID string) {
	url := fmt.Sprintf("%s%s?%s=%s", maxEndpoint, method, CHAT, maxChatID)

	body := make(map[string]string)
	body[MESSAGE] = message

	bodyBytes, errParse := json.Marshal(body)
	if errParse != nil {
		//		fmt.Printf("Error acquired while parsing data: %s", errParse)
		return
	}

	req, err := http.NewRequest(SEND_METHOD, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		//		fmt.Printf("Error acquired while initializing request: %s", err)
		return
	}
	req.Header.Set(AUTH, maxToken)
	req.Header.Set(TYPE, MSG_TYPE)

	client := &http.Client{}
	_, errSend := client.Do(req)

	if errSend != nil {
		//		fmt.Printf("Error acquired while executing request: %s", errSend)
		return
	}
	// Resender works as a daemon without logger so any body check will give nothing
}
