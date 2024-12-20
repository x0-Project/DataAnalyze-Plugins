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
	dsn := "xx:xx@tcp(xx:3306)/xx"
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

	Logger.InfoLogger.Printf("Loaded Plugin: %s, version: %s, Author: %s\n", p.Name(), p.Version(), p.Author())
	once.Do(initializeDatabase)

}

func (p *MyPlugin) ProcessBegin(data *Proto.NetworkData, params map[string]interface{}) error {

	if db == nil {
		log.Fatalf("Database must be initialized before calling ProcessEnd")
	} else {

		go func() {

			isChunkedInt := 0
			if data.IsChunked {
				isChunkedInt = 1
			}

			TraceID := data.TraceID

			ReqType := 0
			if data.ReqType == "REQ" {
				ReqType = 1
			} else {
				ReqType = 0
			}

			timestampSec := time.Now().Unix()
			query := `INSERT INTO cd_flows (TraceID, RawData, ChunkNum, IsChunked, ServiceHost, ServicePort, reqType,node,create_time) VALUES (?, ?, ?, ?, ?, ?, ?,?,?) `

			nodeName := params["nodename"].(string)

			_, err := db.Exec(query, TraceID, hex.EncodeToString(data.RawData), data.ChunkNum, isChunkedInt, data.ServiceHost, data.ServicePort, ReqType, nodeName, timestampSec)

			if err != nil {
				Logger.InfoLogger.Println(err)
			}

		}()
	}

	return nil
}

func (p *MyPlugin) ProcessEnd(data *Proto.NetworkData, params map[string]interface{}) error {

	if db == nil {
		log.Fatalf("Database must be initialized before calling ProcessEnd")
	} else {

		TraceID := data.TraceID

		timestampSec := time.Now().Unix()

		query := `INSERT INTO cd_regexresult (TraceID, node,matchedText, regexgroups_id, regexrules_id,create_time) VALUES (?,?, ?, ?, ?,?)`

		NodeName := params["NodeName"].(string)

		matchText := params["matchedTextTmp"].(string)

		IDGroup := params["ruleTmp"].(Config.Rule).IDGroup

		IDRule := params["ruleTmp"].(Config.Rule).IDRule

		_, err := db.Exec(query, TraceID, NodeName, matchText, IDGroup, IDRule, timestampSec)

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
	return "1.7"
}

func (p *MyPlugin) Author() string {
	return "depy"
}

func (p *MyPlugin) Topic() string {
	return "test"
}

func RegisterPlugin() PluginInterface.Processor {
	return &MyPlugin{}
}
