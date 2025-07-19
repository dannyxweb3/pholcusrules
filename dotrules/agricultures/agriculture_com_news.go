package agricultures

// 基础包
import (
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
	"fmt"
	// "math"
	// "time"

	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出

	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用
	articlewriter "github.com/dannyxweb3/pholcusrules/articlewriter"
	"github.com/dannyxweb3/pholcusrules/consts"
	"github.com/dannyxweb3/pholcusrules/langtranslate"
)

const (
	HOME_URL      = "https://www.agriculture.com/"
	FIRST_URL     = "https://www.agriculture.com/news"
	TECH_URL      = "https://www.agriculture.com/news/technology"
	MACHINE_URL   = "https://www.agriculture.com/news/machinery"
	LIVESTOCK_URL = "https://www.agriculture.com/news/livestock"
	VIEW_URL      = "https://www.agriculture.com/views/ajax"
)

var trans = langtranslate.SelectTranslator(langtranslate.TRANSLATOR_YOUDAO)

func init() {

	trans.SetApiConfig(map[string]interface{}{"appid": "20180125000118458", "appsecret": "htbcOMDlQ_Q3f2Eq93up"})
	trans.SetFromLang("en")
	trans.SetToLang("zh")

	Agriculture_com.Register()
}

func getPageUrl(baseUrl string, pageNo int) string {
	return fmt.Sprintf("%s?page=%d", baseUrl, pageNo)
}

var Agriculture_com = &Spider{
	Name:        "Agriculture.com",
	Description: "www.agriculture.com/news",
	// Pausetime:    300,
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {

			//keyIn := ctx.GetKeyin()

			ctx.AddQueue(&request.Request{
				Url:  TECH_URL,
				Rule: "TIMELINE",
				Header: http.Header{
					"User-Agent": []string{consts.AGENT_PUBLIC},
					"Referer":    []string{TECH_URL},
				},
			})

			ctx.AddQueue(&request.Request{
				Url:  MACHINE_URL,
				Rule: "TIMELINE",
				Header: http.Header{
					"User-Agent": []string{consts.AGENT_PUBLIC},
					"Referer":    []string{TECH_URL},
				},
			})

			ctx.AddQueue(&request.Request{
				Url:  LIVESTOCK_URL,
				Rule: "TIMELINE",
				Header: http.Header{
					"User-Agent": []string{consts.AGENT_PUBLIC},
					"Referer":    []string{TECH_URL},
				},
			})
		},

		Trunk: map[string]*Rule{
			"TIMELINE": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					query.Find(".views-row").Each(func(ai int, as *goquery.Selection) {
						title := as.Find(".field-content").Find("a").Text()
						href, _ := as.Find(".field-content").Find("a").Attr("href")
						abstract := as.Find(".field-body").Find("p").Text()
						imgUrl, _ := as.Find(".field-image").Find("img").Attr("src")
						viewMark := as.Find(".views-field-type").Find(".field-content").Text()

						logs.Log.Warning("find a article:%v %v viewby:%v", title, href, viewMark)

						if viewMark != "Article" && viewMark != "Sequence" {
							logs.Log.Warning("this article has no rule:[%v] %v %v", viewMark, title, href)
							return
						}

						ctx.AddQueue(&request.Request{
							Url:  href,
							Rule: fmt.Sprintf("DETAIL_%s", strings.ToUpper(viewMark)),
							Header: http.Header{
								"User-Agent": []string{consts.AGENT_PUBLIC},
								"Referer":    []string{HOME_URL},
								"Cookie":     []string{ctx.GetCookie()},
							},
							Temp: map[string]interface{}{
								"title":       title,
								"outer_url":   href,
								"surface_url": imgUrl,
								"abstract":    abstract,
								"viewMark":    viewMark,
							},
						})

					})

				},
			},
			"DETAIL_ARTICLE": {
				ItemFields: []string{
					"Title",
					"Author",
					"Thumb",
					"Time",
					"Abstract",
					"OuterUrl",
					"Content",
					"Title-",
					"Abstract-",
					"Content-",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					author := query.Find(".field-byline").Text()

					contentDom := query.Find(".field-body")

					contentDom.Find(".square").Remove()
					contentDom.Find(".leaderboard").Remove()

					//content, _ := contentDom.Html()
					content := contentDom.Text()

					// 过滤标签
					//re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					//contentText := re.ReplaceAllString(content, "")
					// 内容中如果图片不是

					// Title
					title := ctx.GetTemp("title", "").(string)
					// Author

					// Time
					pubtime := query.Find(".byline-date").Text()

					// Abstract
					abstract := ctx.GetTemp("abstract", "").(string)

					// Keywords
					keywords := ""

					surfaceUrl := ctx.GetTemp("surface_url", "").(string)
					outerUrl := ctx.GetTemp("outer_url", "").(string)

					// translate
					absTransRet, err := trans.Translate(abstract)
					if err != nil {
						logs.Log.Warning("trans error:%v :%v", err, abstract)
					}
					//logs.Log.Warning("TRANS[%v]=>[%v]", abstract, absTransRet)
					//abstract = absTransRet

					titleTransRet, err := trans.Translate(title)
					if err != nil {
						logs.Log.Warning("trans error:%v :%v", err, title)
					}
					//title = titleTransRet

					// 长内容需要拆开翻译，翻译后拼装
					contentArr := strings.Split(content, "\n")
					//logs.Log.Notice("content from dom count=%d", len(contentArr))
					contentTransed := ""
					for _, contentLine := range contentArr {
						contentTransRet, err := trans.Translate(contentLine)
						if err != nil {
							contentLineAbsLen := len(contentLine)
							if contentLineAbsLen > 20 {
								contentLineAbsLen = 20
							}
							logs.Log.Warning("trans error:%v :%v", err, contentLine[:contentLineAbsLen])
						}
						contentTransed += contentTransRet + "<br/>\n"
					}
					//content = contentTransRet

					logs.Log.Warning("will write a article:%v", title)

					if true {

						// 输出到mysql
						artInfo := map[string]string{
							"title":       titleTransRet,
							"author":      author,
							"surface_url": surfaceUrl,
							"outer_url":   outerUrl,
							"origin":      "agri-zh",
							"remark":      keywords,
							"abstract":    absTransRet,
							"content":     contentTransed,
							//"pubdate": pubtime,
						}

						buf, err := json.Marshal([]map[string]string{artInfo})
						if err != nil {
							logs.Log.Warning("json marshal error:%v", err)
						}

						writer := &articlewriter.ArticleWriter{}

						_, err = writer.Write(buf)
						if err != nil {
							logs.Log.Warning("write article writer to mysql error:%v", err)
						}
					}

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: author,
						2: surfaceUrl,
						3: pubtime,
						4: abstract,
						5: outerUrl,
						6: content,
						7: titleTransRet,
						8: absTransRet,
						9: contentTransed,
					})
				},
			},
			"DETAIL_SEQUENCE": {
				ItemFields: []string{
					"Title",
					"Author",
					"Thumb",
					"Time",
					"Abstract",
					"OuterUrl",
					"Content",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()

					author := query.Find(".field-byline").Text()

					content := ""
					query.Find(".pane-content").Find(".slides").Find("li").Each(func(lii int, lis *goquery.Selection) {
						theImgHtml, _ := lis.Find(".field-image").Html()
						theCnt, _ := lis.Find(".step-content").Html()
						// step over ads
						if theCnt != "" {
							content += theImgHtml + theCnt + "\n"
							stepTitle := lis.Find(".step-title").Text()
							logs.Log.Warning("find a li:%v %s", lii, stepTitle)
						}
					})

					// Title
					title := ctx.GetTemp("title", "").(string)
					// Author

					// Time
					pubtime := query.Find(".byline-date").Text()

					// Abstract
					abstract := ctx.GetTemp("abstract", "").(string)

					// Keywords
					keywords := ""

					surfaceUrl := ctx.GetTemp("surface_url", "").(string)
					outerUrl := ctx.GetTemp("outer_url", "").(string)

					logs.Log.Warning("will write a article:%v", title)

					// 输出到mysql
					artInfo := map[string]string{
						"title":       title,
						"author":      author,
						"surface_url": surfaceUrl,
						"outer_url":   outerUrl,
						"origin":      "agri",
						"remark":      keywords,
						"abstract":    abstract,
						"content":     content,
						//"pubdate": pubtime,
					}

					if false {

						buf, err := json.Marshal([]map[string]string{artInfo})
						if err != nil {
							logs.Log.Warning("json marshal error:%v", err)
						}

						writer := &articlewriter.ArticleWriter{}

						_, err = writer.Write(buf)
						if err != nil {
							logs.Log.Warning("write article writer to mysql error:%v", err)
						}
					}

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: author,
						2: surfaceUrl,
						3: pubtime,
						4: abstract,
						5: outerUrl,
						6: content,
					})
				},
			},
		},
	},
}
