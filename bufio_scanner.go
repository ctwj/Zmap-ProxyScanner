/*
	(c) Yariya
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func takeIpAndPort(s string) string {
	re := regexp.MustCompile(`Discovered open port (\d+)/tcp on (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	match := re.FindStringSubmatch(s)
	if len(match) == 3 {
		return match[2] + ":" + match[2]
	}
	return ""
}

func Scanner() {
	if *fetch != "" {
		log.Printf("Detected URL Mode.\n")
		res, err := http.Get(*fetch)
		if err != nil {
			log.Fatalln("fetch error")
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalln("fetch body error")
		}
		res.Body.Close()

		scanner := bufio.NewScanner(bytes.NewReader(body))
		for scanner.Scan() {
			ip := scanner.Text()
			queueChan <- ip
		}
	} else if *input != "" {
		fmt.Printf("Detected FILE Mode.\n")
		b, err := os.ReadFile(*input)
		if err != nil {
			log.Fatalln("open file err")
		}
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			queueChan <- line
		}
	} else {
		fmt.Printf("Detected ZMAP Mode.\n")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			// 这里读取出ip和端口， 再，组合成一个字符串，输入到 queueChan
			// "Discovered open port 8585/tcp on 77.134.77.194"
			str := takeIpAndPort(scanner.Text())

			if str != "" {
				queueChan <- str
			}

			// ip := scanner.Text()
			// queueChan <- ip
		}
	}
}
