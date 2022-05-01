# GoGetIt

一个练手项目，主要使用爬虫抓取最新的种子磁力链接并存储，同时分析种子质量，使用RPC推送到Aria2进行下载

## Todo
* [x] 抓取主流的站点并为种子去重
* [x] 使用sqlite存储抓取下来的种子磁力链接
* [x] 编写定时任务，按指定的周期生成可供下载的磁力链接页面
* [ ] 抓取其他主流站点，提供更多内容
* [ ] 白名单机制，只下载某些系列或演员
* [ ] 采用邮件推送的方式发送链接