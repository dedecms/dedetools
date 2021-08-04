package encode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/charsetutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// Document Struct
// 内容结构 ...
type Encoding struct {
	text    string
	bytes   []byte
	Charset string
}

func GetEncoding(src []byte) (*Encoding, bool) {
	e := &Encoding{
		text:  string(src),
		bytes: src,
	}
	return e.converter()
}

func (e *Encoding) Text() string {
	return e.text
}
func (e *Encoding) Bytes() []byte {
	return e.bytes
}

// Converter Function
// 运行对当前进程进行编码转换成UTF-8 ...
func (e *Encoding) converter() (*Encoding, bool) {

	// 自动获取资源编码 ...
	charset, ok := e.getCharset()

	e.Charset = charset

	// 未获取到资源编码 ...
	if !ok {
		return e, false
	}

	// UTF-8无需转换 ...
	if charset == "UTF-8" {
		return e, true
	}

	if encode := getEncoding(charset); encode != nil {
		if reader, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(e.bytes), encode.NewDecoder())); err == nil {
			e.bytes = reader
			e.text = string(reader)
			return e, true
		} else {
			fmt.Println(err)
		}
	}

	// 转码失败
	return e, false
}

func getEncoding(charset string) encoding.Encoding {
	if e, err := ianaindex.MIB.Encoding(charset); err == nil && e != nil {
		return e
	}
	return nil
}

/*---------------------------------------------------------------*/

// Charset Function
// 返回当前进程的字符集 ...
func (e *Encoding) getCharset() (string, bool) {

	// 自动获取编码 ...
	encoding, err := charsetutil.GuessBytes(e.bytes)

	// 如果自动获取成功或encoding不为空
	// 则输出编码格式 ...
	if err == nil && encoding.Charset() != "ISO-8859-1" {
		return strings.ToUpper(encoding.Charset()), true
	}

	if isGBK(e.bytes) {
		return "GBK", true
	}

	if encoding != nil && encoding.Charset() != "WINDOWS-1252" {
		return strings.ToUpper(encoding.Charset()), true
	}

	// 如果内容中出现汉字
	// 则输出GB18030 ...
	if isHan(e.text) {
		return "GBK", true
	}

	if encoding.Charset() == "ISO-8859-1" {
		return strings.ToUpper(encoding.Charset()), true
	}

	// 不符合上述条件
	// 则返回空 ...
	return "", false
}

/*---------------------------------------------------------------*/

// IsHan Function
// 判断是否存在中文 ...
func isHan(str string) bool {
	hanLen := len(regexp.MustCompile(`[\P{Han}]`).ReplaceAllString(str, ""))
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || hanLen > 0 {
			return true
		}
	}
	return false
}

// 判断是否为GBK
func isGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0xff {
			i++
			continue
		} else {
			//大于127的使用双字节编码
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

/*---------------------------------------------------------------*/

/* End of file encoding.go */
/* Location: ./encoding.go */
