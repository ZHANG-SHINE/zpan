## Linux
```bash
# 安装服务
curl -sSf https://dl.saltbo.cn/install.sh | sh -s zpan

# 调整配置
vi /etc/zpan/zpan.yml

# 启动服务
systemctl start zpan

# 查看服务状态
systemctl status zpan

# 设置开机启动
systemctl enable zpan
```

## Docker

如何安装docker请见：
https://www.docker.com/get-started

```bash
docker run -p 8222:8222 --name zpan -itd saltbo/zpan:latest
```

## CORS

!> 由于我们采用浏览器端直传，所以存在跨域问题，请进行如下跨域配置

- Origin: http://your-domain
- AllowMethods: PUT
- AllowHeaders: content-type,content-disposition,x-amz-acl

### Usage
>管理员默认账号密码会输出到Stdout，请到启动日志里获取并记录


启动了docker容器后使用以下命令获取账号密码：

```bash
xxx@xxx:~ » docker ps
CONTAINER ID   IMAGE                COMMAND         CREATED          STATUS          PORTS                    NAMES
4d319679c87e   saltbo/zpan:latest   "zpan server"   33 seconds ago   Up 32 seconds   0.0.0.0:8222->8222/tcp   zpan
xxx@xxx:~ » docker logs 4d319679c87e
Using config file: /etc/zpan/zpan.yml
2021/01/31 05:31:18 AdminEmail: admin@moreu.io
2021/01/31 05:31:18 AdminPassword: RbnxpZLl
WARN[0000] grbac abandoned the periodic loader because loadInterval is less than 0
WARN[0000] grbac abandoned the periodic loader because loadInterval is less than 0
2021/01/31 05:31:19 [rest server listen at :8222]
[GIN] 2021/01/31 - 05:32:08 | 302 |     408.756µs |      172.17.0.1 | GET      "/"
[GIN] 2021/01/31 - 05:32:08 | 200 |   50.263804ms |      172.17.0.1 | GET      "/moreu/signin?redirect=/"
[GIN] 2021/01/31 - 05:32:08 | 200 |    2.608926ms |      172.17.0.1 | GET      "/moreu/css/chunk-vendors.635bc84a.css"
[GIN] 2021/01/31 - 05:32:08 | 200 |      531.14µs |      172.17.0.1 | GET      "/moreu/css/index.7cb19618.css"
[GIN] 2021/01/31 - 05:32:08 | 200 |     6.45459ms |      172.17.0.1 | GET      "/moreu/js/chunk-vendors.d8746001.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |       33.46µs |      172.17.0.1 | GET      "/moreu/js/chunk-2d217357.16174a20.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |     168.547µs |      172.17.0.1 | GET      "/moreu/js/chunk-2d22ca67.f24b4d05.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |      97.121µs |      172.17.0.1 | GET      "/moreu/js/chunk-12dd1456.ab27bc96.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |      43.984µs |      172.17.0.1 | GET      "/moreu/js/chunk-1b01c185.cf674718.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |     150.821µs |      172.17.0.1 | GET      "/moreu/js/index.b1fbb183.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |     266.436µs |      172.17.0.1 | GET      "/moreu/js/chunk-e162f832.45c2adf9.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |     449.424µs |      172.17.0.1 | GET      "/moreu/js/chunk-75a56222.4d61ec5e.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |    1.135008ms |      172.17.0.1 | GET      "/moreu/js/chunk-f328b7a6.6f3488df.js"
[GIN] 2021/01/31 - 05:32:08 | 200 |     310.744µs |      172.17.0.1 | GET      "/moreu/fonts/element-icons.535877f5.woff"
```
中间的AdminEmail和AdminPassword即为账号密码

浏览 http://localhost:8222 即可访问页面
