package files

import (
	"fmt"
	"github.com/FedorovVladimir/go-log/logs"
	"io/ioutil"
	"os"
	"strconv"
)

func ReadChatId() int64 {
	file, _ := os.Open("chat_id.txt")
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logs.LogError(err)
		}
	}(file)
	b, _ := ioutil.ReadAll(file)
	i, _ := strconv.Atoi(string(b))
	return int64(i)
}

func SaveChatId(id int64) {
	f, _ := os.Create("chat_id.txt")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logs.LogError(err)
		}
	}(f)
	_, _ = f.WriteString(fmt.Sprintf("%v", id))
}
