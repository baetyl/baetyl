# 安装 docker

```shell
curl -sSL https://get.docker.com | sh
```

# 安装 openedge

```
curl -sL http://106.13.24.234/install.sh | sudo -E bash -
```

# 连接云端管理套件

## Step1: 下载 agent 模块证书

## Step2: 将证书文件夹放置到 /usr/local

## Step3: 重新启动 OpenEdge

```shell
sudo systemctl restart openedge
# OR
sudo /etc/init.d/openedge restart
```
