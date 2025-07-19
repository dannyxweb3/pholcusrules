package langtranslate_test

import (
	"testing"

	. "github.com/dannyxweb3/pholcusrules/langtranslate"
)

func TestYoudaoTranslate(t *testing.T) {
	trans := SelectTranslator(TRANSLATOR_YOUDAO_FREE)

	trans.SetApiConfig(map[string]interface{}{"appid": "20180125000118458", "appsecret": "htbcOMDlQ_Q3f2Eq93up"})
	trans.SetFromLang("en")
	trans.SetToLang("zh")

	str := "the big brother is watching you"
	res, err := trans.Translate(str)
	if err != nil {
		t.Errorf("trans error:%v", err)
	}
	if len(res) == 0 || res == str {
		t.Errorf("trans failed: origin:%s", res)
	}

	t.Logf("trans over, res=%v", res)
}
