# WebPure

[![Release](https://img.shields.io/github/v/release/cloudwego/netpoll)](https://github.com/jacoblai/webpure/releases)
[![License](https://img.shields.io/github/license/jacoblai/webpure)](https://github.com/jacoblai/webpure/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/jacoblai/webpure)](https://goreportcard.com/report/github.com/jacoblai/webpure)
[![OpenIssue](https://img.shields.io/github/issues/jacoblai/webpure)](https://github.com/jacoblai/webpure/issues)
[![ClosedIssue](https://img.shields.io/github/issues-closed/jacoblai/webpure)](https://github.com/jacoblai/webpure/issues?q=is%3Aissue+is%3Aclosed)
![Stars](https://img.shields.io/github/stars/jacoblai/webpure)
![Forks](https://img.shields.io/github/forks/jacoblai/webpure)

## 简介

WebPure 是 Golang 开发的无依赖 Web 服务器，兼容支持 Nginx配置文件格式，支持Vue,React,Angular等现代前端。

## 特性

* **已经支持**
    - `SharePort` 端口共享
    - `ShareDomain` 域名分流
    - `TLS,SSL` 支持
    - 支持 Nginx配置文件格式
    - 支持 Linux，macOS（操作系统）

* **不被支持**
    - Windows（操作系统）

## 运行

* **参数**
    - `-f` 指定配置文件夹路径，会自动识别文件夹下所有以.conf扩展名的配置文件，默认为程序启动所在路径目录。

```
$ curl https://github.com/jacoblai/webpure/blob/main/bin/webpure_amd64_linux
$ sudo chmod -x webpure_amd64_linux
// /var/www/conf.d文件夹为自定义配置文件所有路径，雷同于Nginx的/etc/nginx/conf.d
$ ./webpure_amd64_linux -f /var/www/conf.d
```

## 热重载配置文件

当配置文件修改后不需要重启WebPure,通过ps -ef|grep webpure找到pid后通过以下指令热更新所有配置文件
```
$ kill -USR2 pid
```

## 配置文件参考

* 普通Web站点配置
```
server {
       listen 80;

       server_name aaa.com;
       root /var/www/dist;
       index index.html;

       location / {
               try_files $uri $uri/ =404;
       }
}
```

* TLS,SSL站点配置
```
server {
       listen 443 ssl;

       server_name aaa.net;
       ssl_certificate  /var/www/dist/aaa.net.pem;
       ssl_certificate_key /var/www/dist/aaa.net.key;
       root /var/www/dist;
       index index.html;

       location / {
               try_files $uri $uri/ =404;
       }
}
```
