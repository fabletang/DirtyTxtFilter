package filter

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Filter 敏感词过滤器
type Filter struct {
	trie  *Trie
	noise *regexp.Regexp
}

// New 返回一个敏感词过滤器
func New() *Filter {
	return &Filter{
		trie:  NewTrie(),
		noise: regexp.MustCompile("[`\\s~☆★!@#$%^&*()+=|{}':;,\\[\\]》·.<>/?~！@#￥%……（）——+|{}【】‘；：”“’。，、？]+"),
		//noise: regexp.MustCompile(`[\|\s&%$@*]+`),
		//noise: regexp.MustCompile(`^[\\u4e00-\\u9fa5]$|[\|\s&%$@*]+`),
		//标点符号
		//var punctuationRegEx = "[`~☆★!@#$%^&*()+=|{}':;,\\[\\]》·.<>/?~！@#￥%……（）——+|{}【】‘；：”“’。，、？]";
	}
}

// UpdateNoisePattern 更新去噪模式
func (filter *Filter) UpdateNoisePattern(pattern string) {
	filter.noise = regexp.MustCompile(pattern)
}

// LoadWordDict 加载敏感词字典
func (filter *Filter) LoadWordDict(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return filter.Load(f)
}

// LoadNetWordDict 加载网络敏感词字典
func (filter *Filter) LoadNetWordDict(url string) error {
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	rsp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	return filter.Load(rsp.Body)
}

// Load common method to add words
func (filter *Filter) Load(rd io.Reader) error {
	buf := bufio.NewReader(rd)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		filter.trie.Add(string(line))
	}

	return nil
}
var hzRegexp = regexp.MustCompile("^[\u4e00-\u9fa5]$")
//标点符号
//var punctuationRegEx = "[`~☆★!@#$%^&*()+=|{}':;,\\[\\]》·.<>/?~！@#￥%……（）——+|{}【】‘；：”“’。，、？]";

//var emojiRegexp = regexp.MustCompile("[\ud83d\udd2a]+")
//func (filter *Filter) ReplaceEmoji(src string,replace rune) (haspass bool,rs string) {
//	var builder strings.Builder
//	for _, c := range src {
//		if hzRegexp.MatchString(string(c)) {
//			//strn += string(c)
//			builder.WriteRune('*')
//			haspass=false;
//		}
//	}
//	return true, builder.String()
//}

func (filter *Filter) FilterEmojiAndNotChinese(content string) string {
	var builder strings.Builder
	//new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		//过滤4字节
		if size <= 3 {
			//new_content += string(value)
			//过滤非中文字符
			if hzRegexp.MatchString(string(value)){
				builder.WriteRune(value)
			}
		}else{
			//println(string(value))
		}
	}
	//return new_content
	return builder.String()
}
func (filter *Filter) FilterEmoji(content string) string {
	var builder strings.Builder
	//new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		//过滤4字节
		if size <= 3 {
			//new_content += string(value)
			//过滤非中文字符
			builder.WriteRune(value)
		}
	}
	//return new_content
	return builder.String()
}

// AddWord 添加敏感词
func (filter *Filter) AddWord(words ...string) {
	filter.trie.Add(words...)
}

// Filter 过滤敏感词
func (filter *Filter) Filter(text string) string {
	return filter.trie.Filter(text)
}
// Filter 过滤敏感词
func (filter *Filter) CheckAndFilter(text string)(haspass bool,rs string) {
	tmp := filter.RemoveNoise(text)
	if len(tmp)<1{
		//println("==all 字符")
	   return true,""
	}
	//过滤表情符号以及非中文字符
	tmp=filter.FilterEmojiAndNotChinese(tmp)
	if len(tmp)<1{
		//println("==all 表情符")
		return true,""
	}
	//过滤中文
	haspass,rs=filter.trie.CheckAndFilter(tmp)
	//没有找到,试图匹配英文
	if haspass {
		tmp = filter.RemoveNoise(text)
		if len(tmp)<1{
			return true,""
		}
		//过滤表情符号
		tmp=filter.FilterEmoji(tmp)
		if len(tmp)<1{
			return true,""
		}
		haspass,rs=filter.trie.CheckAndFilter(tmp)
	}
	return
}

// Filter 检测并且替换敏感词
func (filter *Filter) CheckAndReplace(text string,replace rune)(haspass bool,rs string) {
	tmp := filter.RemoveNoise(text)
	if len(tmp)<1{
		//println("==all 字符")
		return true,""
	}
	//过滤表情符号以及非中文字符
	tmp=filter.FilterEmojiAndNotChinese(tmp)
	//if len(tmp)<1{
	//	println("==all 表情符")
	//	return true,""
	//}
	//过滤中文
	haspass,rs=filter.trie.CheckAndReplace(tmp,replace)
	//没有找到,试图匹配英文
	if haspass {
		tmp = filter.RemoveNoise(text)
		if len(tmp)<1{
			return true,""
		}
		//过滤表情符号
		tmp=filter.FilterEmoji(tmp)
		if len(tmp)<1{
			return true,""
		}
		haspass,rs=filter.trie.CheckAndReplace(tmp,replace)
	}
	return
}
// Replace 和谐敏感词
func (filter *Filter) Replace(text string, repl rune) string {
	return filter.trie.Replace(text, repl)
}

// FindIn 检测敏感词
func (filter *Filter) FindIn(text string) (bool, string) {
	text = filter.RemoveNoise(text)
	return filter.trie.FindIn(text)
}

// FindAll 找到所有匹配词
func (filter *Filter) FindAll(text string) []string {
	return filter.trie.FindAll(text)
}

// Validate 检测字符串是否合法
func (filter *Filter) Validate(text string) (bool, string) {
	text = filter.RemoveNoise(text)
	//过滤表情符号
	text=filter.FilterEmojiAndNotChinese(text)
	return filter.trie.Validate(text)
}

// RemoveNoise 去除空格等噪音
func (filter *Filter) RemoveNoise(text string) string {
	return filter.noise.ReplaceAllString(text, "")
}
