# 简介
本工具旨在实现文件的传输，将本地文件上传到服务器，以及将文件从服务器下载到本地  
工具包含两部分：本地client和服务器端  
# 使用方式定义
## 文件上传
``` shell script
ft send local_file_path
```
## 文件下载
``` shell script
ft fetch server_file_name
```
## 列出服务器文件列表
``` shell script
ft list
```
## 服务器启动
```shell script
ft serve [-p=8888]
```
# 安全定义
为了保证数据的安全性，文件上传时对数据流进行加密，文件在服务器端也加密保存，下载到本地的同时进行解密  
当然也可以不加密存储，只要在send时不指定密码即可
## 客户端认证
服务器端给客户端分配一个RSA的加密密钥，客户端每次向服务器端发送数据时，先发送一段用RSA加密的时间戳信息，客户端接收到加密信息后，进行数据解密，并验证时间戳和本地时间的偏差是否在5s以内  
> RSA密钥对可以使用<http://tools.jb51.net/password/rsa_encode/>或<https://www.bejson.com/enc/rsa/>生成
## 加密方式
文件流使用`AES`进行加密 
## 文件上传
SHA256(timestamp) + len(fileName) + file_name + file_stream  
# 配置文件格式定义
## 服务器端
```json
{
    "rsaDecKey": "rsa解密密钥",
    "warehouse": "本地文件仓库目录"
}
```
## 客户端
```json
{
    "rsaEncKey": "rsa解密密钥",
    "serverAddress": "服务器接口地址",
    "fileEncryptPwd": "文件加密密码"
}
```