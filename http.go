/*
	(c) Yariya
*/

package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"h12.io/socks"
)

type Proxy struct {
	ips                  map[string]struct{}
	targetSites          []string
	httpStatusValidation bool
	timeout              time.Duration
	maxHttpThreads       int64

	openHttpThreads int64
	mu              sync.Mutex
}

var Proxies = &Proxy{
	// in work
	targetSites: []string{"https://google.com", "https://cloudflare.com"},

	httpStatusValidation: false,
	// now cfg file
	timeout:        time.Second * 5,
	maxHttpThreads: int64(config.HttpThreads),
	ips:            make(map[string]struct{}),
}

func (p *Proxy) WorkerThread() {
	for {
		for atomic.LoadInt64(&p.openHttpThreads) < int64(config.HttpThreads) {
			p.mu.Lock()
			for proxy, _ := range p.ips {
				go p.CheckAllProxyType(proxy)
				delete(p.ips, proxy)
				break
			}
			p.mu.Unlock()

		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (p *Proxy) checkSuccess(ip string, port int, proxyType string) {
	if config.PrintIps.Enabled {
		go PrintProxy(ip, port)
	}
	atomic.AddUint64(&success, 1)
	exporter.Add(fmt.Sprintf("%s:%d,%s", ip, port, proxyType))
}

func (p *Proxy) checkFail() {
	atomic.AddUint64(&proxyErr, 1)
}

// 检测所有类型的代理
func (p *Proxy) CheckAllProxyType(proxy string) {
	// 先检测 socks5 类型的代理
	ip, port, err := p.prevCheckProxy(proxy)
	if err != nil {
		log.Println(err)
		return
	}

	// 检测 socks5 类型的代理
	success, _ := p.performCheck(ip, port, "socks5")
	if success {
		p.checkSuccess(ip, port, "socks5")
		return
	}
	// 检测 https 类型的代理
	success, _ = p.performCheck(ip, port, "https")
	if success {
		p.checkSuccess(ip, port, "https")
		return
	}

	// 检测 http 类型的代理
	success, _ = p.performCheck(ip, port, "http")
	if success {
		p.checkSuccess(ip, port, "http")
		return
	}

	p.checkFail()
}

// 检测代理前 预处理操作
func (p *Proxy) prevCheckProxy(proxy string) (string, int, error) {
	atomic.AddInt64(&p.openHttpThreads, 1)
	defer func() {
		atomic.AddInt64(&p.openHttpThreads, -1)
		atomic.AddUint64(&checked, 1)
	}()

	s := strings.Split(proxy, ":")
	if len(s) < 2 {
		return "", 0, errors.New("invalid proxy format")
	}

	ip := strings.TrimSpace(s[0])
	port, err := strconv.Atoi(strings.TrimSpace(s[1]))
	if err != nil {
		return "", 0, err
	}

	return ip, port, nil
}

// 执行代理检测函数，参数为检测类型，ip 和端口
func (p *Proxy) performCheck(ip string, port int, proxyType string) (bool, bool) {
	var tr *http.Transport

	switch proxyType {
	case "http":
		// 解析 HTTP 代理地址
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, port))
		if err != nil {
			log.Println(err)
			return false, false
		}

		// 配置 HTTP Transport
		tr = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			DialContext: (&net.Dialer{
				Timeout:   time.Second * time.Duration(config.Timeout.HttpTimeout),
				KeepAlive: time.Second,
				DualStack: true,
			}).DialContext,
		}
		break

	case "https":
		// 解析 HTTPS 代理地址
		proxyUrl, err := url.Parse(fmt.Sprintf("https://%s:%d", ip, port))
		if err != nil {
			log.Println(err)
			return false, false
		}

		// 配置 HTTPS Transport
		tr = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			DialContext: (&net.Dialer{
				Timeout:   time.Second * time.Duration(config.Timeout.HttpTimeout),
				KeepAlive: time.Second,
				DualStack: true,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过证书验证
			},
		}
		break

	case "socks4":
		// 配置 SOCKS4 Transport
		tr = &http.Transport{
			Dial: socks.Dial(fmt.Sprintf("socks4://%s:%d?timeout=%ds", ip, port, config.Timeout.Socks4Timeout)),
		}
		break

	case "socks5":
		// 配置 SOCKS5 Transport
		tr = &http.Transport{
			Dial: socks.Dial(fmt.Sprintf("socks5://%s:%d?timeout=%ds", ip, port, config.Timeout.Socks5Timeout)),
		}
		break

	default:
		// 不支持的代理类型
		return false, false
	}

	// 创建 HTTP 客户端
	client := http.Client{
		Timeout:   time.Second * time.Duration(config.Timeout.HttpTimeout),
		Transport: tr,
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", config.CheckSite, nil)
	if err != nil {
		log.Println(err)
		return false, false
	}
	req.Header.Add("User-Agent", config.Headers.UserAgent)
	req.Header.Add("accept", config.Headers.Accept)

	// 发送请求并获取响应
	res, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			// 连接超时错误
			return false, true
		}
		// 其他错误
		return false, false
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// 响应状态码非 200，检测失败
		return false, false
	} else {
		if config.PrintIps.Enabled {
			// 打印代理地址
			go PrintProxy(ip, port)
		}
		// 检测成功
		return true, false
	}
}
