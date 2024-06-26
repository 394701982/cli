# fofa_cli

## 功能介绍

该命令行工具用于通过 FOFA API 获取搜索结果，并对结果中的网站进行浏览器截图，输出网站列表和截图结果列表。

## 功能特性
### 基础功能
1. 支持命令行参数输入搜索关键词
2. 获取 FOFA 搜索结果，包括网站 URL、状态码、网站标题
3. 对网站进行截图并保存截图文件
4. 输出网站信息及截图结果
### 额外功能
1. 支持配置代理转发
2. 支持配置并发截图数量
3. 支持配置从 FOFA 查询的结果数量
4. 支持从配置文件读取配置
## 使用方法

### 1. 安装依赖
首先，安装所需的 Go 库
1. go get -u github.com/chromedp/chromedp
2. go get -u gopkg.in/yaml.v2
### 2. 命令行输入
./cli -k FOFA_API_KEY -q 搜索关键词
## 参数说明
- -k：FOFA API Key，fofa的 API Key 
- -q：搜索关键词
## 输出格式
- 工具将输出以下信息，每行一条：
- 网站 URL,状态码,网站标题,对应的截图文件


## 示例

go run main.go -k 912112f04e3c2f038a421fa7f6b56b74 -q 百度 
### 示例输出

1. https://www.ipfscloud.top, 200, 星联云-打造下一代互联网基础, screenshots/aHR0cHM6Ly93d3cuaXBmc2Nsb3VkLnRvcA==.png
2. https://ipfscloud.top, 200, 星联云-打造下一代互联网基础, screenshots/aHR0cHM6Ly9pcGZzY2xvdWQudG9w.png

## 配置文件

### 使用配置文件

go run main.go -k FOFA_API_KEY -q 百度 -config config.yaml

## 配置文件参数说明
1. results_limit: 50    //配置从fofa查询的结果数量
2. proxy: ""            //配置代理转发
3. concurrency: 2       //配置并发截图数量

## 作者
sunhanfei



