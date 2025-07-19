package tpd

// 基础包
import (
	// "log"

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
	// "net/url"

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
	HOME_URL      = "https://theporndude.com"
	AGENT_PUBLIC  = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36"
	AGENT_WX      = "Mozilla/5.0 (Linux; Android 6.0; 1503-M02 Build/MRA58K) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile MQQBrowser/6.2 TBS/036558 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/WIFI Language/zh_CN"
	AGENT_WX_3G   = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/3G+"
	AGENT_WX_WIFI = "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/6.0 NetType/WIFI"
	AGENT_WX_IOS  = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Mobile/14D27 MicroMessenger/6.5.6 NetType/4G Language/zh_CN"
	AGENT_WX_AND  = "Mozilla/5.0 (Linux; Android 5.1; OPPO R9tm Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.49 Mobile MQQBrowser/6.2 TBS/043220 Safari/537.36 MicroMessenger/6.5.7.1041 NetType/4G Language/zh_CN"
)

func init() {
	MyWebSpider.Register()
}

var MyWebSpider = &Spider{
	Name:        "Thepdude",
	Description: "Thepdude",
	// Pausetime:    300,
	// Keyin:        KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  HOME_URL,
				Rule: "HOME",
				Header: http.Header{
					"User-Agent": []string{AGENT_PUBLIC},
				},
			})
		},

		Trunk: map[string]*Rule{
			"HOME": {
				ParseFunc: func(ctx *Context) {
					// title = .category-container/.category-header/h2/a
					// desc = .category-container/.category-header/.desc
					// css address: https://cdn.staticstack.net/includes/css/main.min.css?t=1752159517
					// icon = .category-container/.category-header/.icon-category/.icon-category-xxx @background-image:url(xxx)
					// icon potentially be: https://cdn.staticstack.net/includes/images/categories/{category}.svg
					// link = .category-container/.category-bottom/a/@href
					cnt := 0
					query := ctx.GetDom()
					query.Find(".category-container").Each(func(j int, s *goquery.Selection) {
						// each category
						header := s.Find(".category-header")
						title := header.Find("h2 a").Text()
						link := header.Find("h2 a").AttrOr("href", "")
						// link := header.Find(".category-bottom a").AttrOr("href", "")
						desc := header.Find(".desc").Text()
						iconcss, _ := header.Find(".icon-category").Attr("class")
						iconcssarr := strings.Split(iconcss, " ")
						iconbgcss := ""
						for _, css := range iconcssarr {
							if strings.Contains(css, "icon-category-") {
								iconbgcss = css
								break
							}
						}
						iconbgcss = strings.TrimPrefix(iconbgcss, "icon-category-")
						// <span class="icon-category icon-category-xxx lazyloaded"></span>

						// put category to api.
						icon := "https://cdn.staticstack.net/includes/images/categories/" + iconbgcss + ".svg"
						cateId := saveCategory(title, icon, iconbgcss, desc)
						logs.Log.Informational("get CATEGORY:%s link:%s, desc:%s, icon:%s id:%s/%d", title, link, desc, iconbgcss, cateId, cnt)
						cnt++

						if link != "" {
							// ctx.AddQueue(&request.Request{Url: link, Rule: "CATEGORY_DETAIL", Header: http.Header{"Referer": []string{HOME_URL}, "User-Agent": []string{AGENT_PUBLIC}}, Temp: request.Temp{"cate_id": cateId}})
						}
					})

				},
			},
			"CATEGORY_DETAIL": {
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
					cateId := ctx.GetTemp("cate_id", "").(string)
					query := ctx.GetDom()
					query.Find(".url_link_container").Each(func(i int, s *goquery.Selection) {
						// link := s.AttrOr(".data-external-link", "")
						link := s.Find(".url_link_title").Find("a").AttrOr("data-site-link", "")
						iconcss := s.Find(".icon-site").AttrOr("class", "")
						// <a class=" icon-site ctm-icon ctm-icon12780" href="#">A {{.Title}}</a>
						// <span class=" icon-site ctm-icon ctm-icon628" href="#">S {{.Title}}</span>

						title := s.Find(".url_link_title").Text()
						// icon := ""
						imgPreview := s.Find(".link img").AttrOr("data-src", "")
						desc := s.Find(".url_short_desc").Text()
						// 如果link是 https://pdude.link 开头的，则请求获取301跳转的url

						logs.Log.Informational("get link:%s, title:%s, icon:%s, imgPreview:%s desc:%s", link, title, iconcss, imgPreview, desc)

						// 输出给api
						saveLink(link, title, "", iconcss, imgPreview, desc, cateId)

					})
				},
			},
		},
	},
}

// 请求api 保存category
func saveCategory(title, icon, iconcss, desc string) string {
	// curl -XPOST http://192.168.194.135:8000/api/admin/category -H'Token: xxx' -d 'name=$title&icon=$icon&iconcss=$iconcss&desc=$desc&is_used=1'
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjEsImV4cCI6MTc4NDQyNDE1MCwibmJmIjoxNzUyODg4MTUwLCJpYXQiOjE3NTI4ODgxNTB9.daD8-FufO1oH8heJv1ysemi3To3ycHZkHDeGWntSDkI"
	url := "http://192.168.194.135:8000/api/admin/category"
	req, err := http.NewRequest("POST", url, strings.NewReader("name="+title+"&icon="+icon+"&iconcss="+iconcss+"&desc="+desc+"&is_used=1"))
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

	logs.Log.Informational("category %s saved, status:%d", title, resp.StatusCode)
	resstr, _ := io.ReadAll(resp.Body)
	id, _ := jsonparser.GetInt(resstr, "id")
	return fmt.Sprintf("%d", id)
}

// 请求api 保存link
func saveLink(link, title, icon, iconcss, imgPreview, desc, cateId string) {
	// curl -XPOST http://192.168.194.135:8000/api/admin/site/add -H'Token: xxx' -d 'title=$title&icon=$icon&iconcss=$iconcss&desc=$desc&is_used=1'
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjEsImV4cCI6MTc4NDQyNDE1MCwibmJmIjoxNzUyODg4MTUwLCJpYXQiOjE3NTI4ODgxNTB9.daD8-FufO1oH8heJv1ysemi3To3ycHZkHDeGWntSDkI"
	url := "http://192.168.194.135:8000/api/admin/site/add"
	req, err := http.NewRequest("POST", url, strings.NewReader("title="+title+"&icon="+icon+"&icon_css="+iconcss+"&description="+desc+"&url="+link+"&img_preview="+imgPreview+"&category_id="+cateId+"&is_used=1"))
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
