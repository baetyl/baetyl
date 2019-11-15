# 项目名称
baetyl-remote-object 提供上传文件/打包文件上传 bos，并且支持流量控制

## 快速开始

- make 
- make clean
- make image
  
**注：make操作会删除baetyl下的vendor**
**注：删除原因是由于gomqtt引用路径不一致**

## 配置项
```YAML
hub:
  address: tcp://127.0.0.1:1883
  username: test
  password: hahaha
clients:
  - name: baidu-bos
    address: bos.gz.qasandbox.bcetest.baidu.com
    ak: ***
    sk: ***
    kind: BOS
    pool:
      worker: 1000
      idletime: 30s
    bucket: bos-remote-demo
    temppath: var/db/baetyl/tmp
    # max 10g
    limit:
      data: 9g
      path: var/db/baetyl/data/stats.yml
rules:
  - clientid: remote-write-bos
    subscribe:
      # hub topic
      topic: t
    client:
      name: baidu-bos
logger:
  path: var/log/baetyl/service.log
  level: "debug"
```
### 释义
| 配置项 | 含义 |
| ----- | --- |
| hub | HUB模块配置 |
| hub.address | HUB模块通信地址 |
| hub.clientid | 连接HUB模块时使用的客户端ID |
| hub.username | 连接HUB模块时使用的用户名 |
| hub.password | 连接HUB模块时使用的密码 |
| clients | 客户端配置‘s |
| client.name | 客户端名称 |
| client.address | 访问地址 BOS/CEPH，包含http:// 或者 https://|
| client.region | aws S3, 区域 |
| client.ak | ak |
| client.sk | sk |
| client.kind | 目前只支持BOS/CEPH/S3 |
| client.pool | 协程池 |
| client.pool.worker | 协程池最大协程数 |
| client.pool.idletime | 空闲清理时间 |
| client.bucket | bos bucket |
| client.temppath | 打包类型暂存地址 |
| client.limit | 上传限制 |
| client.limit.data | 限制流量/月，单位 k,m,g,t,p；若数值小于等于0，则不限制 |
| client.limit.path| 持久化统计地址 |
| client.multipart | 大文件分块 |
| client.multipart.partsize | 块大小,默认10m，超过10m分块 |
| client.multipart.concurrency| 并发上传块数量，默认10 |
| rules | 主题转发规则 |
| rule.clientid | 规则ID |
| rules.subscribe.topic | 订阅hub的主题 |
| rules.client.name | 使用的client |
| logger | 日志配置 |
| logger.path | 日志文件路径 |
| logger.level | 日志文件路径 |

### 示例

| 配置文件 | 服务 |
| service-bos.yml | BOS |
| service-ceph.yml | CEPH |
| service-s3.yml | AWS S3 |

### ssl
对于百度云 BOS 和 AWS S3 系统内置了根 CA，
对于自建 ceph， 若进行 ssl 调用 则需要将根 CA 挂载致容器 /etc/ssl/certs/ 目录下

**上传事件**

- 上传
```JSON   
{
    "type":"UPLOAD",
    "content":{
        "remotePath":"dir/abc.zip",
        "zip":true,
        "localPath":"var/pic/2019-01-01.png",
        "meta": {
            "key1":"value1",
            "key2":"value2"
        }
    }
}
```

一个上传日志的示例，文件夹打tar包
```JSON   
{
    "type":"UPLOAD",
    "content":{
        "remotePath":"log/log.tar",
        "zip":false,
        "localPath":"var/log/baetyl"
    }
}
```

一个上传日志的示例，文件夹打zip包
```JSON   
{
    "type":"UPLOAD",
    "content":{
        "remotePath":"log/log.zip",
        "zip":true,
        "localPath":"var/log/baetyl"
    }
}
```

### 释义
| 配置项 | 含义 |
| ----- | --- |
| type | UPLOAD |
| content.remotePath | bos bucket内路径 |
| content.zip | 打包zip? true打包为zip，false下，对于文件不作改动，对于文件夹打包为tar|
| content.localPath | 本地目录，可是文件/文件夹，文件夹可打包zip/tar |
| content.meta | k-v map[string]string  |


## 测试
如何执行自动化测试

## 如何贡献
贡献patch流程、质量要求

## 讨论
百度Hi讨论群：XXXX

## 链接
[百度golang代码库组织和引用指南](http://wiki.baidu.com/pages/viewpage.action?pageId=515622823)