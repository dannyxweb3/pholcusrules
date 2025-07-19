// attack.mitre.org unfinished
package mitre

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
	//"strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

const (
	URL_TACTICS_EXECUTION   = "https://attack.mitre.org/tactics/TA0002/"
	URL_TACTICS_PERSISTANCE = "https://attack.mitre.org/tactics/TA0003/"

	AGENT_PUBLIC  = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36"
	AGENT_WX      = "Mozilla/5.0 (Linux; Android 6.0; 1503-M02 Build/MRA58K) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile MQQBrowser/6.2 TBS/036558 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/WIFI Language/zh_CN"
	AGENT_WX_3G   = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/3G+"
	AGENT_WX_WIFI = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/WIFI"
	AGENT_WX_IOS  = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Mobile/14D27 MicroMessenger/6.5.6 NetType/4G Language/zh_CN"
	AGENT_WX_AND  = "Mozilla/5.0 (Linux; Android 5.1; OPPO R9tm Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.49 Mobile MQQBrowser/6.2 TBS/043220 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/4G Language/zh_CN"
)

var TECHNIQUE_TABLE_IDX = []int{3, 4} //3=Column of Execution; 4=Column of Persistance

func init() {
	Wx100000p.Register()
}

var Wx100000p = &Spider{
	Name:        "mitre activities",
	Description: "find mitre activities and relative Event IDs",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  URL_TACTICS_EXECUTION,
				Rule: "Matrix-for-Enterprise",
				Header: http.Header{
					"User-Agent": []string{AGENT_PUBLIC},
				},
			})
			ctx.AddQueue(&request.Request{
				Url:  URL_TACTICS_PERSISTANCE,
				Rule: "Matrix-for-Enterprise",
				Header: http.Header{
					"User-Agent": []string{AGENT_PUBLIC},
				},
			})
		},

		Trunk: map[string]*Rule{
			"Matrix-for-Enterprise": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					contentTables := query.Find(".table-techniques").Find("table")
					for _, idx := range TECHNIQUE_TABLE_IDX {
						cateTable := contentTables.Eq(idx)
						cateTable.Find(".technique-cell").Each(func(j int, s *goquery.Selection) {
							logs.Log.Informational("find a row:%s", s.Find("a").Text())
						})
					}
					// .Each(func(j int, s *goquery.Selection) {

					// 	a := s.Find("a")
					// 	if url, ok := a.Attr("href"); ok {
					// 		// log.Print(url)
					// 		p := a.Find("p")

					// 		ctx.AddQueue(&request.Request{Url: HOME_URL + url, Rule: "DETAIL", Header: http.Header{"Referer": []string{HOME_URL}, "User-Agent": []string{AGENT_PUBLIC}}, Temp: request.Temp{"abstract": p.Text()}})
					// 	}
					// })

				},
			},
			"ROW-DETAIL": {
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
