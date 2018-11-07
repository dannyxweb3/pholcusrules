package somexiangbao

// 基础包
import (
	"encoding/json"
	"net/http" //设置http.Header
	"strings"
	// "log"

	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出
	articlewriter "github.com/uxff/pholcusrules/articlewriter"
	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用
)

//思路
// 抓取资讯 抓取发布者 自动注册 自动发布到对应板块

const (
	HOME_URL = "http://www.baigou.net"

	AGENT_PUBLIC  = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36"
	AGENT_WX      = "Mozilla/5.0 (Linux; Android 6.0; 1503-M02 Build/MRA58K) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile MQQBrowser/6.2 TBS/036558 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/WIFI Language/zh_CN"
	AGENT_WX_3G   = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/3G+"
	AGENT_WX_WIFI = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/WIFI"
	AGENT_WX_IOS  = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Mobile/14D27 MicroMessenger/6.5.6 NetType/4G Language/zh_CN"
	AGENT_WX_AND  = "Mozilla/5.0 (Linux; Android 5.1; OPPO R9tm Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.49 Mobile MQQBrowser/6.2 TBS/043220 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/4G Language/zh_CN"
)

func init() {
	Xiangbao1.Register()
}

var Xiangbao1 = &Spider{
	Name:        "Xiangbao1",
	Description: "http://www.baigou.net",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  HOME_URL,
				Rule: "TIMELINE",
				Header: http.Header{
					"User-Agent": []string{AGENT_PUBLIC},
				},
			})
		},

		Trunk: map[string]*Rule{
			"TIMELINE": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find(".site-content").Find("article").Each(func(j int, s *goquery.Selection) {
						a := s.Find("a")
						if url, ok := a.Attr("href"); ok {
							// log.Print(url)
							p := a.Find("p")

							ctx.AddQueue(&request.Request{Url: HOME_URL + url, Rule: "DETAIL", Header: http.Header{"Referer": []string{HOME_URL}, "User-Agent": []string{AGENT_PUBLIC}}, Temp: request.Temp{"abstract": p.Text()}})
						}
					})

				},
			},
			"DETAIL": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"Title",
					"Author",
					"Thumb",
					"Time",
					"Abstract",
					"Keywords",
					"Content",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					// 获取内容
					content, _ := query.Find(".entry-content").Html()

					// 过滤标签
					//re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					//contentText := re.ReplaceAllString(content, "")
					// 内容中如果图片不是

					// Title
					title := query.Find(".entry-title").Text()
					// Author

					// Addresses & Address
					addresses, ok := query.Find(".post-cover-title-image").Attr("style")
					if ok {
						attrArr := strings.Split(addresses, ";")
						for _, subAttr := range attrArr {
							if subAttr[:20] == "background-image:url" {
								addresses = subAttr[20:]
								addresses = strings.Trim(addresses, " ()'\"")
								break
							}
						}

					}

					// Time
					pubtime := query.Find(".entry-date").Text()

					// Abstract
					author := ""
					abstract := ctx.GetTemp("abstract", "").(string)

					// Keywords
					keywords := ""

					query.Find(".tag").Each(func(ti int, q *goquery.Selection) {
						keywords = keywords + "," + q.Text()
					})

					keywords = strings.Trim(keywords, ", \t\r\n")

					// 输出到mysql
					artInfo := map[string]string{
						"title":       title,
						"author":      author,
						"surface_url": addresses,
						"outer_url":   ctx.GetUrl(),
						"origin":      "wx100000p",
						"remark":      keywords,
						"abstract":    abstract,
						"content":     content,
						//"pubdate": pubtime,
					}

					buf, err := json.Marshal([]map[string]string{artInfo})
					if err != nil {
						logs.Log.Warning("json marshal error:%v", err)
					}

					writer := &articlewriter.ArticleWriter{}

					_, err = writer.Write(buf)
					if err != nil {
						logs.Log.Warning("write wx100000p to mysql error:%v", err)
					}

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: author,
						2: addresses,
						3: pubtime,
						4: abstract,
						5: keywords,
						6: content,
					})
				},
			},
		},
	},
}
