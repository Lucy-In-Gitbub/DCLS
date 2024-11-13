package dcls

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	xlsx "github.com/tealeg/xlsx/v3"
)

func TestDcls(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	Bucket := NewBucketClient(rdb)

	//模拟并发请求
	var requests int = 100
	var sprequests int = 2000
	gw := sync.WaitGroup{}
	gw.Add(requests)
	sp := sync.WaitGroup{}
	sp.Add(sprequests)

	count := atomic.Int64{}
	spcount := atomic.Int64{}
	j := 29
	var cap int64 = 50
	var rate int64 = 20
	for i := 0; i < requests; i++ {
		go func(i int) {
			defer gw.Done()

			status, err := Bucket.Check(context.Background(), "test", cap, rate)

			//模拟请求
			if status {
				count.Add(1)
			}

			//打印日志
			now := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf(now+"i: %d,status: %v,err: %v\n", i, status, err)
		}(i)

		time.Sleep(10 * time.Millisecond)
	}
	gw.Wait()
	fmt.Printf("count: %d\n", count.Load())

	for i := 0; i < sprequests; i++ {
		go func(i int) {
			defer sp.Done()

			status, err := Bucket.Check(context.Background(), "test", cap, rate)

			//模拟请求
			if status {
				spcount.Add(1)
			}

			//打印日志
			now := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf(now+"i: %d,status: %v,err: %v\n", i, status, err)
		}(i)

	}
	sp.Wait()
	fmt.Printf("spcount: %d\n", spcount.Load())
	//打开文件
	file, err := xlsx.OpenFile("../data/data.xlsx")
	if err != nil {
		fmt.Println(err)
		recover()
	}

	//选择表表格
	sheet := file.Sheets[0]

	//row 0,accept 1,cap 2,rate
	cell, err := sheet.Cell(j, 0)
	if err != nil {
		fmt.Println(err)
		recover()
	}
	cell.Value = strconv.FormatInt(int64(count.Load()), 10)

	cell, err = sheet.Cell(j+1, 0)
	if err != nil {
		fmt.Println(err)
		recover()
	}
	cell.Value = strconv.FormatInt(int64(spcount.Load()), 10)

	cell, err = sheet.Cell(j, 1)
	if err != nil {
		fmt.Println(err)
		recover()
	}
	cell.Value = strconv.FormatFloat(float64(count.Load())/float64(requests), 'f', 2, 64)

	cell, err = sheet.Cell(j+1, 1)
	if err != nil {
		fmt.Println(err)
		recover()
	}
	cell.Value = strconv.FormatFloat(float64(spcount.Load())/float64(requests), 'f', 2, 64)

	err = file.Save("../data/data.xlsx")
	if err != nil {
		fmt.Println(err)
		recover()
	}
}
