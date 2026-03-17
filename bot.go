package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/goloop/env"

	"bot/api/max"
	"bot/api/vk"
)

const (
	// .env constants
	DEFAULT_PORT        = ":9090" // FOR TEST: 9080; FOR DEPLOY 9090
	DEFAULT_DESTINATION = "ALL"
	VK_DESTINATION      = "VK"
	MAX_DESTINATION     = "MAX"
	OK_RESPONSE         = "HTTP/1.1 200 OK\r\n"
	NOT_OK_RESPONCE     = "HTTP/1.1 500 ERROR\r\n"
	TIMEOUT             = 5 * time.Second

	// Zabbix constants
	MESSAGE      = "Message"
	SUBJECT      = "Subject"
	TO           = "To"
	SAP_MON_DATA = `\[.*\]`
	DESTINATION  = "Destination"

	// Vk constants
	VK_CHAT             = "peer_id"
	VK_MESSAGE          = "message"
	VK_UNIQUE_CHECK_REQ = "random_id"
	NOT_NEEDED          = "0"
	VK_MESSAGES_METHOD  = "messages.send"

	// MAX constants
	MAX_MESSAGES_METHOD = "messages"
)

var (
	isLoaded            = true
	VK_API_VER          string
	VK_API_ENDPOINT     string
	VK_API_ACCESS_TOKEN string

	MAX_API_ENDPOINT     string
	MAX_API_ACCESS_TOKEN string

	PORT        string
	SEND_TARGET string = "ALL"
)

type Config struct {
	VKVersion   string `env:"VK_API_VER" def:"NONE"`
	VKEndpoint  string `env:"VK_API_ENDPOINT" def:"NONE"`
	VKToken     string `env:"VK_API_ACCESS_TOKEN" def:"NONE"`
	MAXEndpoint string `env:"MAX_API_ENDPOINT" def:"NONE"`
	MAXToken    string `env:"MAX_API_ACCESS_TOKEN" def:"NONE"`
}

func init() {
	if err := env.Load(".env"); err != nil {
		//		fmt.Println("Could not load the .env parameters")
		isLoaded = false
		return
	}

	var cfg Config
	if err := env.Unmarshal("", &cfg); err != nil {
		//		fmt.Println("Could not parse the .env parameters")
		isLoaded = false
		return
	}

	flag.StringVar(&PORT, "port", DEFAULT_PORT, "The port on wich the bot is running")
	flag.StringVar(&SEND_TARGET, "dest", DEFAULT_DESTINATION, "Messenger that will recieve data")
	VK_API_VER = cfg.VKVersion
	VK_API_ACCESS_TOKEN = cfg.VKToken
	VK_API_ENDPOINT = cfg.VKEndpoint
	MAX_API_ENDPOINT = cfg.MAXEndpoint
	MAX_API_ACCESS_TOKEN = cfg.MAXToken
	if (VK_API_VER == "NONE" || VK_API_ACCESS_TOKEN == "NONE" || VK_API_ENDPOINT == "NONE") && (SEND_TARGET == VK_DESTINATION || SEND_TARGET == DEFAULT_DESTINATION) || (MAX_API_ACCESS_TOKEN == "NONE" || MAX_API_ENDPOINT == "NONE") && (SEND_TARGET == MAX_DESTINATION || SEND_TARGET == DEFAULT_DESTINATION) {
		isLoaded = false
	}
}

func main() {
	if !isLoaded {
		return
	}
	flag.Parse()
	if !strings.HasPrefix(PORT, ":") {
		PORT = ":" + PORT
	}

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		//		fmt.Printf("Could not start server, error %s\n", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			//			fmt.Printf("Error acquired while recieving connection %s\n", err)
			continue
		}
		go resendIntoBot(conn)
	}
}

func resendIntoBot(connection net.Conn) {
	defer connection.Close()
	connection.SetReadDeadline(time.Now().Add(TIMEOUT))
	connection.SetWriteDeadline(time.Now().Add(TIMEOUT))

	request, err := http.ReadRequest(bufio.NewReader(connection))
	if err != nil {
		if err == io.EOF {
			//			fmt.Println("Соединение закрыто")
			connection.Write([]byte(OK_RESPONSE))
		} else {
			//			fmt.Printf("Ошибка чтения запроса: %v\n", err)
			connection.Write([]byte(NOT_OK_RESPONCE))
		}
		return
	}

	body, _ := io.ReadAll(request.Body)
	bodyStr := string(body)
	bodyStr, _ = url.QueryUnescape(bodyStr)
	var responseData = make(map[string]string)
	json.Unmarshal([]byte(bodyStr), &responseData)

	match, errCheck := regexp.Match(SAP_MON_DATA, []byte(responseData[MESSAGE]))
	if errCheck == nil && match {
		re := regexp.MustCompile(SAP_MON_DATA)
		res := re.FindStringSubmatch(responseData[MESSAGE])

		responseData[MESSAGE] = strings.ReplaceAll(responseData[MESSAGE], res[0], "")
	}

	connection.Write([]byte(OK_RESPONSE))

	switch SEND_TARGET {
	case DEFAULT_DESTINATION:
		switch responseData[DESTINATION] {
		case VK_DESTINATION:
			vk.ResendMessage(VK_MESSAGES_METHOD, map[string]string{VK_CHAT: responseData[TO],
				VK_MESSAGE: url.QueryEscape(responseData[SUBJECT] + "\n" + responseData[MESSAGE]), VK_UNIQUE_CHECK_REQ: NOT_NEEDED},
				VK_API_VER, VK_API_ENDPOINT, VK_API_ACCESS_TOKEN)
		case MAX_DESTINATION:
			max.ResendMessage(MAX_MESSAGES_METHOD, responseData[SUBJECT]+"\n"+responseData[MESSAGE],
				MAX_API_ENDPOINT, MAX_API_ACCESS_TOKEN, responseData[TO])
		}
	case VK_DESTINATION:
		vk.ResendMessage(VK_MESSAGES_METHOD, map[string]string{VK_CHAT: responseData[TO],
			VK_MESSAGE: url.QueryEscape(responseData[SUBJECT] + "\n" + responseData[MESSAGE]), VK_UNIQUE_CHECK_REQ: NOT_NEEDED},
			VK_API_VER, VK_API_ENDPOINT, VK_API_ACCESS_TOKEN)
	case MAX_DESTINATION:
		max.ResendMessage(MAX_MESSAGES_METHOD, responseData[SUBJECT]+"\n"+responseData[MESSAGE],
			MAX_API_ENDPOINT, MAX_API_ACCESS_TOKEN, responseData[TO])
	}
}
