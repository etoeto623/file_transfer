# 简介
本工具旨在实现文件的传输，将本地文件上传到服务器，以及将文件从服务器下载到本地  
工具包含两部分：本地client和服务器端  
# 使用方式定义
## 文件上传
``` shell script
ft send [-c=config_file] [-p=password] [-n=file_name] [-s=server_address] local_file_path
```
## 文件下载
``` shell script
ft fetch [-c=config_file] [-p=password] [-s=server_address] server_file_name
```
## 服务器启动
```shell script
ft serve [-c=config_file] [-p=8888]
```
# 安全定义
为了保证数据的安全性，文件上传时对数据流进行加密，文件在服务器端也加密保存，下载到本地的同时进行解密  
当然也可以不加密存储，只要在send时不指定密码即可
## 客户端认证
服务器端给客户端分配一个RSA的加密密钥，客户端每次向服务器端发送数据时，先发送一段用RSA加密的时间戳信息，客户端接收到加密信息后，进行数据解密，并验证时间戳和本地时间的偏差是否在10s以内
## 加密方式
文件流使用`AES`进行加密，以256K为一个块进行分块加密传输，密码信息保留在文件的头部，使用`HMAC-SHA256`对密码进行哈希运算
## 安全校验
客户端从服务端拉取文件时，需要发送密码信息和文件名，服务器端收到文件名和密码后，对比本地该文件的文件头中的密码信息，一致的话就表示通过验证
# 数据传输格式
使用一个byte来定义数据块的类型，如下：

|  bits  |  meaning | bits follows | memo |
| :---- | :---- | :---- | ---- |
| 00000001 | 客户端认证 | 256 | 认证使用RSA加密，该256bit定义紧随其后的密文的长度(bytes) |     
| 00000010 | 文件密码 |512 | 密码使用SHA256加密，为字符串长度64，因此占512bits |  
| 00000011 | 文件名 | 8 | 定义的是文件名的长度 |  
| 00000100 | 文件流开始 | | |  

## 文件上传
SHA256(timestamp) + passwordFlag(0/1) + \[password\] + file_name + file_stream  
# 配置文件格式定义
## 服务器端
```json
{
    "rsaKey": "rsa解密密钥",
    "warehouse": "本地文件仓库目录"
}
```
## 客户端
```json
{
    "rsaKey": "rsa解密密钥",
    "serverAddress": "服务器接口地址"
}
```