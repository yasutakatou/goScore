# goScore (Golang Server)

**Simple URL's security score check tool. (implement by Golang and Chrome Extension.)**.

# solution

When you're search knowledge, suddenly you may see a alert by Security Scan tool.<br>
You have unfortunately displayed a site that contains a *malware script*.<br>
But you've search a lot of for learn. I don't blame you.<br>
<br>
**Can't we at least easily check if it's suspicious before we open it?**<br>
<br>
I created this tool with that in mind *for you*!<br>

# features

When you right-click and select **context menu**, it opens a **temporary tab** before URL be opened.<br>
In that tab, **request Golang API** on server and will check the following.<br>

- Can name **resolve on the Secure DNS server**?
- Is SSL certificate is **not free type**?
- Is **higher rank search results** on search engines?

Check results are displayed with a **score(star rate)**.

# installation

If you want to put it under the path, you can use the following.

```
go get github.com/yasutakatou/goScore
```

If you want to create a binary and copy it yourself, use the following.

```
git clone https://github.com/yasutakatou/goScore
cd goScore
go build goScore.go
```

[or download binary from release page](https://github.com/yasutakatou/goScore/releases).<br>
save binary file, copy to entryed execute path directory.

# uninstall

delete that binary. del or rm command. (it's simple!)

# usecase

## 0. prepare Cert file

### You prepare cert file beforehand. (use mkcert and such more)

## 1. prepare server

- run server and **space or enter press** will display **ip addres and port**.

![1](https://github.com/yasutakatou/goScore/blob/pic/1.png)

## 2. setting on Chrome extension

- input server address to chrome extension.

![2a](https://github.com/yasutakatou/goScore/blob/pic/2a.png)

- in case of success, **it will be added**.

![2b](https://github.com/yasutakatou/goScore/blob/pic/2b.png)

## 3. Check URL

- in case of checking url, **right click on link, display context menu, send to server**.

![3a](https://github.com/yasutakatou/goScore/blob/pic/3a.png)

- new tab will open, waiting for making score from server.

![3b](https://github.com/yasutakatou/goScore/blob/pic/3b.png)

- **score displayed**.

![3c](https://github.com/yasutakatou/goScore/blob/pic/3c.png)

If you click center link, **go to original URL**.<br>
**(So, You can check before we open it.)**<br>

note) **if url cached, server return old result**. So, it's fast.<br>

# run options

this options use when run.<br>

|option name|default value|detail|
|:---|:---|:---|
-debug|false|debug mode (true is enable)|
-cert|localhost.pem|ssl_certificate file path (if you don't use https, haven't to use this option)|
-key|localhost-key.pem|ssl_certificate_key file path (if you don't use https, haven't to use this option)|
-port|8080|port number|
-config|config|config file name([detail: config file](https://github.com/yasutakatou/goScore#config))|

# config

This tool have config file, detail is following.

note) default config file name is **"config"**.

## [CACHE]

cached time.
this value format is value and unit multiplication.
unit is
"d", "D": 12 * 60 * 60
 "h", "H": 60 * 60
"m", "M": 60
"s", "S": 1
example) 
12 D -> 12days
30 M -> 30minuts

[DNS]
9.9.9.9
1.1.1.1

define DNS servers to resolve.





[SEARCH]
1

define google page rank.
If your search results rank lower than this value, Return a star to the client.


[SSL]
DigiCert
GlobalSign
Google Trust Services
GeoTrust
SECOM Passport
Sectigo RSA Organization
Cybertrust

define you trusted ssl certificate authority.



[HISTORY]
https://www.google.co.jp/ 1598448280 000

this value is server cache.
format
url, cached unit time, score.
score "0" is star, "1" is no star.

# problem

- When add server to extension, following message displaying!

*YYYY/MM/DD HH:MM:SS http: TLS handshake error from XXX.XXX.XXX.XXX:XXXXX: remote error: tls: unknown certificate*

If you use **Self-signed cert file**, Chrome must trusted this cert.<br>
You have to access endpoint once by **manualy**.<br>

## browse url to https://xxx.xxx.xxx.xxx:yyyyy/token

If alert from Chrome and you continue access, message will not displayed.<br>

# LICENSE

MIT License
