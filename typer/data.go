package typer

import (
	"github.com/youtiaoguagua/typer/util"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Word struct {
	Name  string   `json:"name"`
	Trans []string `json:"trans"`
}

type WordConfig struct {
	// 长度
	Length int
	// 添加数字
	Numbers bool
	// 添加标点
	Punctuation bool
}

var WordList []Word

func init() {
	err := fetchWordData()
	if err != nil {
		panic(err)
	}
}

func fetchWordData() error {
	err := util.FetchData("https://kaiyiwing.gitee.io/qwerty-learner/dicts/it-words.json", &WordList)
	if err != nil {
		return err
	}
	return nil
}

func GetWordData(config *WordConfig) string {
	length := config.Length
	seed := rand.NewSource(time.Now().UnixMilli())
	r := rand.New(seed)

	var tmpWordList []string

	// 添加100以内数字
	if config.Numbers {
		l := length / 8
		for i := 0; i < l; i++ {
			num := r.Intn(1000)
			tmpWordList = append(tmpWordList, strconv.Itoa(num))
		}
		length -= l
	}

	// 添加单词
	for i := 0; i < length; i++ {
		v := r.Intn(len(WordList))
		tmpWordList = append(tmpWordList, strings.ToLower(WordList[v].Name))
	}

	r.Shuffle(len(tmpWordList), func(i, j int) {
		tmpWordList[i], tmpWordList[j] = tmpWordList[j], tmpWordList[i]
	})

	res := strings.Join(tmpWordList, " ")
	return res
}

func (c *WordConfig) GetNumBtn() string {
	if c.Numbers {
		return BtnPress.Render("numbers")
	}
	return Btn.Render("number")
}

func (c *WordConfig) GetLengthBtn(l string) string {
	if strconv.Itoa(c.Length) == l {
		return BtnPress.Render(l)
	}
	return Btn.Render(l)
}
