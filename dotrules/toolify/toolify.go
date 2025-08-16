package toolify

// 基础包
import (
	// "log"

	"encoding/json"
	"fmt"
	"io"

	"github.com/buger/jsonparser"
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	"github.com/henrylee2cn/pholcus/logs"                   //信息输出

	// . "github.com/henrylee2cn/pholcus/app/spider/common" //选用

	// net包
	"net/http" //设置http.Header
	"net/url"

	// 编码包
	// "encoding/xml"

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
	URL_ROOT     = "https://www.toolify.ai"
	URL_CATEGORY = "https://www.toolify.ai/category?group=writing-editing"

	AGENT_PUBLIC  = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36"
	AGENT_WX      = "Mozilla/5.0 (Linux; Android 6.0; 1503-M02 Build/MRA58K) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile MQQBrowser/6.2 TBS/036558 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/WIFI Language/zh_CN"
	AGENT_WX_3G   = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/3G+"
	AGENT_WX_WIFI = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/WIFI"
	AGENT_WX_IOS  = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Mobile/14D27 MicroMessenger/6.5.6 NetType/4G Language/zh_CN"
	AGENT_WX_AND  = "Mozilla/5.0 (Linux; Android 5.1; OPPO R9tm Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.49 Mobile MQQBrowser/6.2 TBS/043220 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/4G Language/zh_CN"

	Cookie = "locale=en; toolify_isLogin=false; toolify_userinfo=%7B%7D; timezone=Asia%2FShanghai; _ga=GA1.1.1372319488.1755309426; _clck=1rios6e%7C2%7Cfyi%7C0%7C2054; utm=https%3A%2F%2Fwww.toolify.ai%2F; TOOLIFY_SESSION_ID=GlMBm9IZEGydETp29sRTB7D7lx5IN9IYeBvICufE; _ga_FQFB4N81SR=GS2.1.s1755321252$o3$g1$t1755321258$j54$l0$h0; _clsk=b6j4jp%7C1755321259445%7C5%7C1%7Ca.clarity.ms%2Fcollect"
)

var TECHNIQUE_TABLE_IDX = []int{3, 4} //3=Column of Execution; 4=Column of Persistance

func init() {
	Toolify.Register()
}

const (
	CategoryLimit = 100000
	ToolLimit     = 200000
	PageLimit     = 200
)

var (
	categoryIdx = 0
	toolIdx     = 0
)

var Toolify = &Spider{
	Name:        "Toolify",
	Description: "Toolify",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  URL_CATEGORY,
				Rule: "Category",
				Header: http.Header{
					"User-Agent": []string{AGENT_PUBLIC},
					"Cookie":     []string{Cookie},
				},
			})
		},

		Trunk: map[string]*Rule{
			"Category": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					query.Find(".list").Children().Each(func(i int, s *goquery.Selection) {
						mainCategorySlug, _ := s.Attr("id")
						mainCategoryName := s.Find("h3").Text()

						logs.Log.Informational("mainCategoryName: %s, mainCategorySlug: %s", mainCategoryName, mainCategorySlug)

						// mainCategorySlug 格式: group-writing-editing
						// mainCategoryName 格式: Writing & Editing
						// 去掉group前缀
						if len(mainCategorySlug) > 6 && mainCategorySlug[:6] == "group-" {
							mainCategorySlug = mainCategorySlug[6:]
						}
						// ca
						saveCategory(mainCategoryName, mainCategorySlug, "", "")

						s.Find("a").Each(func(i int, s *goquery.Selection) {
							subCategoryName := s.Find("span").Text()
							subCategoryName = strings.Trim(subCategoryName, " \t\r\n")
							subCategorySlugHref := s.AttrOr("href", "")
							subCategorySlug := ""
							// href = "/category/ai-blog-generator"
							if len(subCategorySlugHref) > 10 && subCategorySlugHref[:10] == "/category/" {
								subCategorySlug = subCategorySlugHref[10:]
							}
							// logs.Log.Informational("subCategoryName: %s, subCategorySlug: %s under %s", subCategoryName, subCategorySlug, mainCategorySlug)

							saveCategory(subCategoryName, subCategorySlug, mainCategorySlug, "")

							subCategoryLink := URL_ROOT + subCategorySlugHref

							categoryIdx++
							if categoryIdx > CategoryLimit {
								return
							}
							ctx.AddQueue(&request.Request{
								Url:    subCategoryLink,
								Rule:   "SubCategory",
								Header: http.Header{"Referer": []string{URL_CATEGORY}},
								Temp: request.Temp{
									// "mainCategorySlug": mainCategorySlug,
									"subCategoryName": subCategoryName,
									"subCategorySlug": subCategorySlug,
									"subCategoryUrl":  subCategoryLink,
									"page":            1,
								},
							})
						})
					})

				},
			},
			"SubCategory": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					categorySlug := ctx.GetTemp("subCategorySlug", "").(string)
					categoryName := ctx.GetTemp("subCategoryName", "").(string)
					categoryLink := ctx.GetTemp("subCategoryUrl", "").(string)
					page := ctx.GetTemp("page", 1).(int)

					logs.Log.Informational("subCategoryName: %s, subCategorySlug: %s page %d", categoryName, categorySlug, page)

					// 获取内容
					findCnt := 0
					query.Find(".tools").Children().Each(func(i int, s *goquery.Selection) {
						toolLogoSrc := s.Find(".logo-img").Find("img").AttrOr("src", "")
						cardContent := s.Find(".card-text-content")
						toolName := cardContent.Find("h2").Text()
						toolLink := s.Find(".logo-img-wrapper").AttrOr("href", "")
						// toolDesc := cardContent.Find("p").Text()
						toolOutLink := URL_ROOT + toolLink

						toolIdx++
						if toolIdx > ToolLimit {
							return
						}
						findCnt++
						ctx.AddQueue(&request.Request{
							Url:  toolOutLink,
							Rule: "ToolDetail",
							Header: http.Header{
								"User-Agent": []string{AGENT_PUBLIC},
								"Cookie":     []string{Cookie},
							},
							Temp: request.Temp{
								"categorySlug": categorySlug,
								"categoryName": categoryName,
								"toolName":     toolName,
								"toolLogoSrc":  toolLogoSrc,
							},
						})

						// todo 考虑翻页
					})

					if findCnt == 0 {
						// no more pages
					} else if page < PageLimit {
						page += 1
						// ctx.GetUrl().SetQuery("page", fmt.Sprintf("%d", page))
						ctx.AddQueue(&request.Request{
							Url:    categoryLink + "?page=" + fmt.Sprintf("%d", page),
							Rule:   "SubCategory",
							Header: http.Header{"Referer": []string{URL_CATEGORY}},
							Temp: request.Temp{
								// "mainCategorySlug": mainCategorySlug,
								"subCategoryName": categoryName,
								"subCategorySlug": categorySlug,
								"subCategoryUrl":  categoryLink,
								"page":            page,
							},
						})
					}

				},
			},
			"ToolDetail": {
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					categorySlug := ctx.GetTemp("categorySlug", "").(string)
					categoryName := ctx.GetTemp("categoryName", "").(string)
					toolLogoSrc := ctx.GetTemp("toolLogoSrc", "").(string)

					// preview img:
					// <meta data-n-head="ssr" data-hid="og:image" name="og:image" content="https://cdn-images.toolify.ai/168580600419974080.jpg?x-oss-process=image/resize,l_1000,m_lfit">
					// logo: https://www.junia.ai/favicon.png

					// breadcrumbs := query.Find(".breadcrumbs")
					toolName := ctx.GetTemp("toolName", "").(string)
					tollSlug := query.Find(".breadcrumbs").AttrOr("data-handle", "")
					tooldetail := query.Find(".tool-detail-info")

					// outerLink := breadcrumbs.Siblings().Next().Find("a").First().AttrOr("href", "")
					desc := tooldetail.Find(".table-row").First().Find(".table-cell").Last().Text()
					socials := []string{}
					tooldetail.Find(".table-row").Last().Find(".table-cell").Find("a").Each(func(i int, s *goquery.Selection) {
						socials = append(socials, s.AttrOr("href", ""))
					})
					tags := []string{}
					tooldetail.Find(".table").Siblings().Last().Children().Each(func(i int, s *goquery.Selection) {
						tags = append(tags, strings.Trim(s.Text(), " \t\r\n"))
					})

					outerLink := tooldetail.Find(".to-view-btn").AttrOr("href", "")
					imgPreview := tooldetail.Find(".to-view-btn").Find("img").AttrOr("src", "")

					introSect := query.Find(".tool-detail-information")
					introBasic := ""
					introUse := ""
					features := []string{}
					introSect.Children().Each(func(i int, s *goquery.Selection) {
						if i == 0 {
							introBasic = strings.Trim(s.Find("p").Text(), " \t\r\n")
						}
						if i == 2 {
							introUse = strings.Trim(s.Find("p").Text(), " \t\r\n")
						}
						if i == 4 {
							features = append(features, strings.Trim(s.Find("h3").Text(), " \t\r\n"))
						}
					})

					prices := [][]string{}
					query.Find(".tool-prices").Find(".price-item").Each(func(i int, s *goquery.Selection) {
						priceItem := []string{}
						s.Find("p").Each(func(i int, p *goquery.Selection) {
							priceItem = append(priceItem, strings.Trim(p.Text(), " \t\r\n"))
						})
						// priceItemStr, _ := json.Marshal(priceItem)
						prices = append(prices, priceItem)
					})

					price_desc, _ := json.Marshal(prices)

					logs.Log.Informational("parsed tool: %s", toolName)
					logInfo := "toolName: " + toolName + ", categorySlug: " + categorySlug + ", categoryName: " + categoryName + ", toolSlug: " + tollSlug + ", toolLogoSrc: " + toolLogoSrc
					logInfo += ", desc: " + desc + ", socials: " + strings.Join(socials, ",") + ", tags: " + strings.Join(tags, ",") + ", outerLink: " + outerLink + ", imgPreview: " + imgPreview
					logInfo += ", introBasic: " + introBasic + ", introUse: " + introUse + ", features: " + strings.Join(features, ",") + ", prices: " + string(price_desc)
					logs.Log.Informational(logInfo)

					saveLink(outerLink, toolName, tollSlug, categoryName, toolLogoSrc, imgPreview, desc, strings.Join(tags, ","), introBasic, introUse, strings.Join(features, ","), string(price_desc), strings.Join(socials, ","))

				},
			},
		},
	},
}

// 请求api 保存category
func saveCategory(title, slug, parent, desc string) string {
	// curl -XPOST http://192.168.194.135:8000/api/admin/category -H'Token: xxx' -d 'name=$title&icon=$icon&iconcss=$iconcss&desc=$desc&is_used=1'
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjEsImV4cCI6MTc4NDQyNDE1MCwibmJmIjoxNzUyODg4MTUwLCJpYXQiOjE3NTI4ODgxNTB9.daD8-FufO1oH8heJv1ysemi3To3ycHZkHDeGWntSDkI"
	theurl := "http://192.168.194.135:8000/api/admin/category"
	// theurl := "http://u20d.local:8000/api/admin/category"
	urlValues := url.Values{}
	urlValues.Add("title", title)
	urlValues.Add("slug", slug)
	urlValues.Add("parent", parent)
	urlValues.Add("desc", desc)
	urlValues.Add("is_used", "1")

	req, err := http.NewRequest("POST", theurl, strings.NewReader(urlValues.Encode()))
	if err != nil {
		logs.Log.Error("to saveCategory error:%s", err)
		return ""
	}

	req.Header.Set("Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Log.Error("to saveCategory error:%s", err)
		return ""
	}

	logs.Log.Informational("category %s saved, data:%s, status:%d", title, urlValues.Encode(), resp.StatusCode)
	resstr, _ := io.ReadAll(resp.Body)
	id, _ := jsonparser.GetInt(resstr, "id")
	return fmt.Sprintf("%d", id)
}

// 请求api 保存link
func saveLink(link, title, slug, category, icon, imgPreview, desc, tags, intro_basic, intro_use, intro_features, price_desc, social string) {
	// curl -XPOST http://192.168.194.135:8000/api/admin/site/add -H'Token: xxx' -d 'title=$title&icon=$icon&iconcss=$iconcss&desc=$desc&is_used=1'
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjEsImV4cCI6MTc4NDQyNDE1MCwibmJmIjoxNzUyODg4MTUwLCJpYXQiOjE3NTI4ODgxNTB9.daD8-FufO1oH8heJv1ysemi3To3ycHZkHDeGWntSDkI"
	theurl := "http://192.168.194.135:8000/api/admin/site/add"
	// theurl := "http://u20d.local:8000/api/admin/site/add"
	urlValue := url.Values{}
	urlValue.Add("title", title)
	urlValue.Add("category", category)
	urlValue.Add("slug", slug)
	urlValue.Add("icon_remote", icon)
	urlValue.Add("description", desc)
	urlValue.Add("url", link)
	urlValue.Add("img_remote", imgPreview)
	// urlValue.Add("category", cateId)
	urlValue.Add("is_used", "1")
	urlValue.Add("tags", tags)
	urlValue.Add("intro_basic", intro_basic)
	urlValue.Add("intro_use", intro_use)
	urlValue.Add("intro_features", intro_features)
	urlValue.Add("price_desc", price_desc)
	urlValue.Add("social", social)

	req, err := http.NewRequest("POST", theurl, strings.NewReader(urlValue.Encode()))
	if err != nil {
		logs.Log.Error("to saveLink error:%s", err)
		return
	}

	req.Header.Set("Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Log.Error("to saveLink error:%s", err)
		return
	}

	logs.Log.Informational("link %s saved, status:%d", title, resp.StatusCode)
}
