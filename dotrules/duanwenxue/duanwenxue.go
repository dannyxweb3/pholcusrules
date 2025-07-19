// https://www.duanwenxue.com/
package duanwenxue

// 基础包
import (
	// "log"

	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出

	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用
	articlewriter "github.com/dannyxweb3/pholcusrules/articlewriter"

	// net包
	"net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	"encoding/json"

	// 字符串处理包
	//"regexp"
	"strconv"
	"strings"

	// 其他包
	"fmt"
	// "math"
	// "time"
)

const (
	HOME_URL      = "https://www.duanwenxue.com/"
	AGENT_PUBLIC  = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36"
	AGENT_WX      = "Mozilla/5.0 (Linux; Android 6.0; 1503-M02 Build/MRA58K) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile MQQBrowser/6.2 TBS/036558 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/WIFI Language/zh_CN"
	AGENT_WX_3G   = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/3G+"
	AGENT_WX_WIFI = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/WIFI"
	AGENT_WX_IOS  = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Mobile/14D27 MicroMessenger/6.5.6 NetType/4G Language/zh_CN"
	AGENT_WX_AND  = "Mozilla/5.0 (Linux; Android 5.1; OPPO R9tm Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.49 Mobile MQQBrowser/6.2 TBS/043220 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/4G Language/zh_CN"
)

func init() {
	Duanwenxue.Register()
}

var Duanwenxue = &Spider{
	Name:        "duanwenxue.com",
	Description: "duanwenxue.com 输入格式如: https://www.duanwenxue.com/duanwen/geyan/",
	// Pausetime:    300,
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			param := ctx.GetKeyin()
			if len(param) < 5 {
				logs.Log.Warning("自定义输入的参数不正确！应该是 https://www.duanwenxue.com/duanwen/geyan/")
				//return
				param = "https://www.duanwenxue.com/duanwen/geyan/"
			}

			ctx.AddQueue(&request.Request{
				Url:  param,
				Rule: "TIMELINE",
				Header: http.Header{
					"User-Agent": []string{AGENT_PUBLIC},
					"Referer":    []string{HOME_URL},
				},
				Temp: map[string]interface{}{
					"pageNo":  "1",
					"baseurl": param,
				},
			})
		},

		Trunk: map[string]*Rule{
			"TIMELINE": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					baseUrl := ctx.GetTemp("baseurl", "").(string)

					// 翻页
					pageNoStr := ctx.GetTemp("pageNo", "1").(string)
					pageNo, _ := strconv.Atoi(pageNoStr)
					pageNo++

					pageAll := query.Find(".pageselect").Find("select").Find("option").Length()
					logs.Log.Warning("page : cur=%v all=%v", pageNo, pageAll)
					if pageNo > pageAll {
						logs.Log.Warning("page over: cur=%v all=%v", pageNo, pageAll)
					} else {
						ctx.AddQueue(&request.Request{
							Url:  fmt.Sprintf("%v/list_%d.html", baseUrl, pageNo),
							Rule: "TIMELINE",
							Header: http.Header{
								"User-Agent": []string{AGENT_PUBLIC},
								"Referer":    []string{HOME_URL},
							},
							Temp: map[string]interface{}{
								"pageNo":  fmt.Sprintf("%d", pageNo),
								"baseurl": baseUrl,
							},
						})
					}

					query.Find(".list-short-article").Find("ul").Find("li").Each(func(j int, s *goquery.Selection) {
						a := s.Find("p").Find("a")
						url, _ := a.Attr("href")

						content := a.Text()
						contentRune := []rune(content)
						// 输出到mysql
						artInfo := map[string]string{
							"title":       string(contentRune[:20]),
							"author":      "xahoo",
							"surface_url": "",
							"outer_url":   url,
							"origin":      "duanwenxue",
							"remark":      "哲理,格言,语录,唯美,名言",
							"abstract":    content,
							"content":     content,
							//"pubdate": pubtime,
						}

						buf, err := json.Marshal([]map[string]string{artInfo})
						if err != nil {
							logs.Log.Warning("json marshal error:%v", err)
							return
						}
						writer := &articlewriter.ArticleWriter{}
						_, err = writer.Write(buf)
						if err != nil {
							logs.Log.Warning("write error:%v", err)
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

					// 输出到mysql
					artInfo := map[string]string{
						"title":       title,
						"author":      author,
						"surface_url": addresses,
						"outer_url":   ctx.GetUrl(),
						"origin":      "duanwenxue",
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
