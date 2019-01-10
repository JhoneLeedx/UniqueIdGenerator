package main

import (
	"./UniqueIdGenerator"
	"./redisTools"
	"fmt"
	"gopkg.in/redis.v4"
	"net/http"
	// "encoding/json"
	"log"
	"strconv"
)

type UnicodeRes struct {
	ret bool `json:"ret"`
	// data map[string]string `json:"data"`
}

var client *redis.Client

func uniqueid(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	countParams := request.Form["count"]
	if len(countParams) > 0 {
		var countString = countParams[0]
		count, err := strconv.Atoi(countString)
		if err != nil || count < 1 {
			// 返回空
			fmt.Fprintf(writer, "参数错误")
		} else {
			idSlice := getUniqueIdByCount(idGenerator1, count)
			var result = ""
			for _, value := range idSlice {
				result += strconv.FormatUint(value, 10)
				result += "\n"
			}
			// var hasSame bool = hasSameValue(idSlice)
			// fmt.Printf("是否有相同的 %t\n", hasSame)
			fmt.Fprintf(writer, result)
		}
	} else {
		// 返回空
		fmt.Fprintf(writer, "参数错误")
	}
}

func getUniqueIdByCount(generator *UniqueIdGenerator.UniqueIdGenerator, count int) []uint64 {
	idSlice, err := generator.GetIdByCount(count)
	if err != nil {
		workId, redisErr := strconv.Atoi(redisTools.GetOperation(client, "workId"))
		if redisErr != nil {
			log.Fatal("redisErr Error: ", redisErr)
		}
		workId++
		if workId < (1 << 12) {
			generator = UniqueIdGenerator.
				CreateIdGenerator().
				SetWorkId(getWorkId(client)).
				SetTimestampBitSize(48).
				SetSequenceBitSize(11).
				SetWorkIdBitSize(5).
				Init()

			idSlice, err = generator.GetIdByCount(count)
		}
		return nil
	}
	return idSlice
}

func hasSameValue(aSlice []uint64) bool {
	var testSlice []uint64
	for index, value := range aSlice {
		for _, v := range testSlice {
			if value == v {
				fmt.Printf("找到相同的值 : %d 在 index : %d", value, index)
				return true
			} else {
				testSlice = append(testSlice, value)
			}
		}
		if index == 0 {
			testSlice = append(testSlice, value)
			break
		}
	}
	return false
}

var idGenerator1 *UniqueIdGenerator.UniqueIdGenerator

func main() {
	client = redisTools.CreateClient()
	redisTools.ConnectPool(client)
	if client == nil {
		return
	}
	idGenerator1 = UniqueIdGenerator.
		CreateIdGenerator().
		SetWorkId(getWorkId(client)).
		SetTimestampBitSize(48).
		SetSequenceBitSize(11).
		SetWorkIdBitSize(5).
		Init()
	if idGenerator1 == nil {
		return
	}
	http.HandleFunc("/uniqueid", uniqueid)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
func getWorkId(client *redis.Client) uint64  {
		workId, redisErr := strconv.Atoi(redisTools.GetOperation(client, "workId"))
		if redisErr != nil {
			log.Fatal("redisErr Error: ", redisErr)
		}
		workId++
	if workId < (1 << 12) {
		redisTools.SetOperation(client, "workId", strconv.Itoa(workId))
		return  uint64(workId)
	} else {
		panic("数据太大了")
	}


}