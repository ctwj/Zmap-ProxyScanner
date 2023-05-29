# Zmap-ProxyScanner

A Thread Safe fast way to find proxies. Find 2000-5000 working http,socks4,socks5 proxies in one scan.

# Config
  ```json
   {
    "check-site": "https://google.com",
    "proxy-type": "http",

    "http_threads": 2000,
    "headers": {
      "user-agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.115 Safari/537.36",
      "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"
    },
    "print_ips": {
      "enabled": true,
      "display-ip-info": true
    },
    "timeout": {
      "http_timeout": 5,
      "socks4_timeout": 5,
      "socks5_timeout": 5
    }
  }
  ```
## Flag Args
  ```shell
-p <port> - Port you want to scan.
-o <proxies.txt> - Writes proxy hits to file.
-input <proxies.txt> - Loads the proxy list and checks it.
-url https://api.com/proxies - Loads the proxies from an api and checks it.
  ```

```
# 通过 masscan


~/Program/go/Zmap-ProxyScanner masscan ⚡
$ sudo ~/safe/masscan/bin/masscan -p8000,3128,8081,8089,8585,9080,1080,7001,7890 --rate=15000 0.0.0.0/0  --exclude 255.255.255.255 | ./ZmapProxyScanner -o proxies.txt



# 通过文件
ZmapProxyScanner -o proxies.txt
```