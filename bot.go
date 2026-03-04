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

	"bot/api/vk"
)

const (
	// .env constants
	DEFAULT_PORT    = ":9090" // FOR TEST: 9080; FOR DEPLOY 9090
	OK_RESPONSE     = "HTTP/1.1 200 OK\r\n"
	NOT_OK_RESPONCE = "HTTP/1.1 500 ERROR\r\n"
	TIMEOUT         = 5 * time.Second

	// Zabbix constants
	MESSAGE      = "Message"
	SUBJECT      = "Subject"
	TO           = "To"
	SAP_MON_DATA = `\[.*\]`

	// Vk constants
	VK_CHAT             = "peer_id"
	VK_MESSAGE          = "message"
	VK_UNIQUE_CHECK_REQ = "random_id"
)

var (
	isLoaded            = true
	VK_API_VER          string
	VK_API_ENDPOINT     string
	VK_API_ACCESS_TOKEN string

	PORT string
)

type Config struct {
	Version  string `env:"VK_API_VER" def:"NONE"`
	Endpoint string `env:"VK_API_ENDPOINT" def:"NONE"`
	Token    string `env:"VK_API_ACCESS_TOKEN" def:"NONE"`
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
	VK_API_VER = cfg.Version
	VK_API_ACCESS_TOKEN = cfg.Token
	VK_API_ENDPOINT = cfg.Endpoint
	if VK_API_VER == "NONE" || VK_API_ACCESS_TOKEN == "NONE" || VK_API_ENDPOINT == "NONE" {
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
			connection.Write([]byte(OK_RESPONSE))
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

	vk.ResendMessageVK("messages.send", map[string]string{VK_CHAT: responseData[TO], // test peer: 2000000001
		VK_MESSAGE: url.QueryEscape(responseData[SUBJECT] + "\n" + responseData[MESSAGE]), VK_UNIQUE_CHECK_REQ: "0"},
		VK_API_VER, VK_API_ENDPOINT, VK_API_ACCESS_TOKEN)
}
