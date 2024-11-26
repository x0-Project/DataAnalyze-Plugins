package main

import (
	"HaE-AnalyzeEngine/Config"
	"HaE-AnalyzeEngine/Logger"
	"HaE-AnalyzeEngine/PluginInterface"
	"HaE-AnalyzeEngine/Proto"
	"database/sql"
	"encoding/hex"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
	"sync"
	"time"
)

type MyPlugin struct{}

var (
	db   *sql.DB
	once sync.Once
)

func initializeDatabase() {

	var err error
	dsn := "YourDbName:YourPwd@tcp(127.0.0.1:3306)/YourDbName"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		Logger.ErrorLogger.Println("Failed to open database: %v", err)
	}
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	db.SetConnMaxLifetime(time.Hour)
	err = db.Ping()
	if err != nil {
		Logger.ErrorLogger.Println("Failed to ping database: %v", err)
	}

	Logger.InfoLogger.Printf("Mysql Server Connect Success！")

}

func (p *MyPlugin) Init() {

	once.Do(initializeDatabase)
	Logger.InfoLogger.Printf("Loaded Plugin: %s, version: %s, Author: %s\n", p.Name(), p.Version(), p.Author())

}

func (p *MyPlugin) ProcessBegin(data *Proto.NetworkData, data2 interface{}, data3 interface{}) error {

	if db == nil {
		log.Fatalf("Database must be initialized before calling ProcessEnd")
	} else {

		go func() {
			isChunkedInt := 0
			if data.IsChunked {
				isChunkedInt = 1
			}

			//存储抹掉TraceID请求类型前缀标志 方便后期定位索引
			TraceID := data.TraceID
			TraceID = strings.TrimPrefix(TraceID, "REQ-")
			TraceID = strings.TrimPrefix(TraceID, "RES-")

			ReqType := 1
			if strings.HasPrefix(data.TraceID, "RES") {
				ReqType = 0
			}

			query := `INSERT INTO cd_flows (TraceID, RawData, ChunkNum, IsChunked, ServiceHost, ServicePort, reqType,node) VALUES (?, ?, ?, ?, ?, ?, ?,?) `

			_, err := db.Exec(query, TraceID, hex.EncodeToString(data.RawData), data.ChunkNum, isChunkedInt, data.ServiceHost, data.ServicePort, ReqType, data2.(string))

			if err != nil {
				Logger.InfoLogger.Println(err)
			}

		}()
	}

	return nil
}

func (p *MyPlugin) ProcessEnd(data *Proto.NetworkData, data2 interface{}, data3 interface{}, data4 interface{}, data5 interface{}) error {

	if db == nil {
		log.Fatalf("Database must be initialized before calling ProcessEnd")
	} else {

		TraceID := data.TraceID
		TraceID = strings.TrimPrefix(TraceID, "REQ-")
		TraceID = strings.TrimPrefix(TraceID, "RES-")

		query := `INSERT INTO cd_regexresult (TraceID, node,matchedText, GroupName, ruleName) VALUES (?,?, ?, ?, ?)`

		_, err := db.Exec(query, TraceID, data2.(string), data3.(string), data4.(Config.Group).GroupName, data5.(Config.Rule).Name)

		if err != nil {
			Logger.InfoLogger.Println(err)
		}
	}

	return nil

}

func (p *MyPlugin) Name() string {
	return "storeData2Mysql"
}

func (p *MyPlugin) Version() string {
	return "1.1"
}

func (p *MyPlugin) Author() string {
	return "depy"
}

func RegisterPlugin() PluginInterface.Processor {
	return &MyPlugin{}
}
