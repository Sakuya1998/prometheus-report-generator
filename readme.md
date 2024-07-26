# 统计每天响应时长最慢的前20url
topk(20, max_over_time({job="nginx_access_log"} | json | url !~ `(.*mp4)|(.*mov)|(.*jpg)|(.*png)|(.*js)|(.*css)` | status = `200` | unwrap request_time [1d])by(http_host,url))

# 统计每天请求次数最多的前20url
topk(20, sum(count_over_time({job="nginx_access_log"} | json | url !~ `(.*mp4)|(.*mov)|(.*jpg)|(.*png)|(.*js)|(.*css)` | status = "200" [1d])) by (http_host,url))

# 统计 状态码 分布
sum(count_over_time({job="nginx_access_log"} | json [1d])) by (status)

# 统计 各域名 QPS 
sum by (http_host) (rate({job="nginx_access_log"} | json | http_host !="" and http_host=~"(.*bestshowtv.com)|(.*playlet.com)|(.*spicychiliti.com)" [1d]))

# 统计 各域名 pv
sum by (http_host) (count_over_time({job="nginx_access_log"} | json | http_host !="" and http_host=~"(.*bestshowtv.com)|(.*playlet.com)|(.*spicychiliti.com)" [1d]))

# 请求成功率
sum (rate({job="nginx_access_log"} | json | http_host !="" and http_host=~"(.*bestshowtv.com)|(.*playlet.com)|(.*spicychiliti.com)" | status=~"2.." [1m]))by(http_host) / sum(rate({job="nginx_access_log"} | json | http_host !="" and http_host=~"(.*bestshowtv.com)|(.*playlet.com)|(.*spicychiliti.com)" [1m]))by(http_host) *100
    
