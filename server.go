package aidns

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server http 服务
func (handler *AiDNS) Server() (err error) {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		if handler.HttpToken != "" {
			r.Use(func(ctx *gin.Context) {
				if ctx.GetHeader("Authorization") != "Bearer "+handler.HttpToken {
					ctx.JSON(http.StatusOK, gin.H{
						"code": 1,
						"msg":  "authorization error",
					})
					ctx.Abort()
				}
			})
		}
		// 列表 添加 更新 删除 records
		r.GET("records", func(ctx *gin.Context) {
			resp, err := handler.findRecordsForZone(ctx)
			if err != nil {
				ctx.JSON(http.StatusOK, gin.H{
					"code": 1,
					"msg":  err.Error(),
				})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{
				"code":    0,
				"records": resp,
			})
		})
		r.POST("records", func(ctx *gin.Context) {
			err := handler.updateRecordsForZone(ctx)
			if err != nil {
				ctx.JSON(http.StatusOK, gin.H{
					"code": 1,
					"msg":  err.Error(),
				})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{
				"code": 0,
			})
		})
		r.DELETE("records", func(ctx *gin.Context) {
			err := handler.deleteRecordsForZone(ctx)
			if err != nil {
				ctx.JSON(http.StatusOK, gin.H{
					"code": 1,
					"msg":  err.Error(),
				})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{
				"code": 0,
			})
		})
		log.Info("Http Server run in port:", handler.HttpAddr)
		err = r.Run(handler.HttpAddr)
		log.Error("Http Server Error", err)
	}()
	return nil
}

// findRecordsForZone 查询 记录
func (handler *AiDNS) findRecordsForZone(ctx *gin.Context) (any, error) {
	zone := ctx.Query("zone")
	sqlQuery := fmt.Sprintf("SELECT id, name, zone, ttl, record_type, content FROM %s WHERE zone = ?",
		handler.tableName)
	result, err := handler.db.Query(sqlQuery, zone)
	if err != nil {
		return nil, err
	}
	var recordName string
	var recordZone string
	var recordType string
	var id uint32
	var ttl uint32
	var content string

	records := make([]*RecordApi, 0)
	for result.Next() {
		err = result.Scan(&id, &recordName, &recordZone, &ttl, &recordType, &content)
		if err != nil {
			return nil, err
		}
		records = append(records, &RecordApi{
			ID:         id,
			Zone:       recordZone,
			Name:       recordName,
			RecordType: recordType,
			Ttl:        ttl,
			Content:    content,
		})
	}
	return records, nil
}

// updateRecordsForZone 添加/更新 记录
func (handler *AiDNS) updateRecordsForZone(ctx *gin.Context) error {
	var params RecordApi
	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		return err
	}
	if params.ID > 0 {
		sqlQuery := fmt.Sprintf("update %s set zone =?, name =?,record_type =?, content =?, ttl =? where id =?",
			handler.tableName)
		_, err = handler.db.Exec(sqlQuery, params.Zone, params.Name,
			params.RecordType, params.Content, params.Ttl, params.ID)
		if err != nil {
			return err
		}
	} else {
		sqlQuery := fmt.Sprintf("insert into %s (zone,name,record_type,content,ttl) values (?,?,?,?,?)",
			handler.tableName)
		_, err = handler.db.Exec(sqlQuery, params.Zone, params.Name,
			params.RecordType, params.Content, params.Ttl)
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteRecordsForZone 删除 记录
func (handler *AiDNS) deleteRecordsForZone(ctx *gin.Context) error {
	var params RecordDelete
	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		return err
	}
	sqlQuery := fmt.Sprintf("delete from %s where id = ? and zone = ?",
		handler.tableName)
	result, err := handler.db.Exec(sqlQuery, params.ID, params.Zone)
	fmt.Println("deleteRecordsForZone", result, err)
	if err != nil {
		return err
	}
	return nil
}
