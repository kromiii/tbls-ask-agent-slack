package main

import (
	"fmt"

	"github.com/kromiii/tbls-ask-server/tbls"
)

var (
	query  string
	answer string
)

func main() {
	query = "ユーザーあたりのポスト数を取得"
	answer = tbls.Ask(query)
	fmt.Println(answer)
}
