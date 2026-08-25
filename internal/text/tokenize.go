package text

import (
	"regexp"
	"strings"
	"unicode"

	"dream117/pkg/textutil"
)

var latinWord = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_-]{1,}`)

var stopWords = map[string]struct{}{
	"然后": {}, "但是": {}, "因为": {}, "所以": {}, "自己": {}, "一个": {}, "感觉": {}, "起来": {}, "看到": {}, "里面": {}, "时候": {}, "东西": {}, "非常": {}, "好像": {}, "突然": {}, "还是": {}, "没有": {}, "可以": {}, "已经": {}, "我们": {}, "他们": {}, "这个": {}, "那个": {}, "有点": {},
}

func Tokens(
	title,
	content string,
) []string {
	value := textutil.Clean(title + " " + content)
	result := make([]string, 0)
	result = append(result,
		latinWord.FindAllString(strings.ToLower(value), -1)...)
	runes := []rune(value)
	for index := 0; index < len(runes); index++ {
		if !unicode.Is(unicode.Han, runes[index]) {
			continue
		}
		if index+1 < len(runes) && unicode.Is(unicode.Han, runes[index+1]) {
			addToken(&result, string(runes[index:index+2]))
		}
		if index+2 < len(runes) && unicode.Is(unicode.Han, runes[index+1]) && unicode.Is(unicode.Han, runes[index+2]) {
			addToken(&result, string(runes[index:index+3]))
		}
	}
	return textutil.Unique(result)
}

func addToken(
	target *[]string,
	token string,
) {
	if len([]rune(token)) > 1 {
		if _, found := stopWords[token]; !found {
			*target = append(*target, token)
		}
	}
}
