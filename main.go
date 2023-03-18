package main

import (
	"SpeechEnhancement/meta"
	"SpeechEnhancement/process"
	"SpeechEnhancement/upload"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/dist", "./dist")
	r.GET("/", Index)
	r.POST("/upload", Upload)
	r.GET("/status", Status)
	r.GET("/download", Download)
	r.Run(":8080")
}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func Upload(c *gin.Context) {
	processUUID := uuid.Must(uuid.NewV4())
	file, err := c.FormFile("file")
	fmt.Println(file.Filename)
	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}
	filename := filepath.Join(upload.DefaultUploadPath, processUUID.String(), file.Filename)
	if err = c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}
	go func() {
		metaFile := filepath.Join(upload.DefaultUploadPath, processUUID.String(), "meta.json")
		err := meta.EncodeWrite(metaFile, meta.Metadata{
			Source:     file.Filename,
			Target:     "",
			Status:     meta.StatusDefault,
			CreateTime: int(time.Now().Unix()),
		})
		if err != nil {
			log.Printf("Error: %s\n", err)
			return
		}
		p, err := process.New(processUUID.String(), file.Filename)
		if err != nil {
			return
		}
		p.SetChain().Handle()
	}()
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"uuid": processUUID.String(),
		},
	})
	return
}

func Status(c *gin.Context) {
	puuid := c.Query("uuid")
	var m meta.M
	status, err := meta.Get(m, puuid, meta.KeyStatus)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "error",
			"data": gin.H{
				"status": meta.StatusDefault,
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"status": status,
		},
	})
	return
}

func Download(c *gin.Context) {
	puuid := c.Query("uuid")
	var m meta.M
	target, err := meta.Get(m, puuid, meta.KeyTarget)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "error",
			"data": gin.H{
				"url": nil,
			},
		})
		return
	}
	sourceFile, err := meta.Get(m, puuid, meta.KeySource)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "error",
			"data": gin.H{
				"url": nil,
			},
		})
		return
	}
	fileParts := strings.Split(sourceFile.(string), ".")
	filename := strings.Join(fileParts[:len(fileParts)-1], ".") + "-after.wav"
	c.FileAttachment(target.(string), filename)
	return
}
