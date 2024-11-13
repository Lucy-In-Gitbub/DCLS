# DCLS

该项目为一个基于Go语言的分布式限流系统

## 项目结构

data: 存放一些数据文件
src: 存放项目源码
pkg: 存放一些工具

## 运行

```fake

import (
    DCLS
)

func main() {
    rdb:=redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        ...
    })

    Bucket := NewBucketClient(rdb)
    status, err := Bucket.Check(context.Background(), "name", cap, rate)
```
