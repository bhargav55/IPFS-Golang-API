package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
    "path/filepath"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"encoding/json"
	shell "github.com/ipfs/go-ipfs-api"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)


var server, user, password, database = "localhost", "SA", "<password>", "ipfsFileStore"
var port = 1433

//FileInfo model
type FileInfo struct {
	
	FileID   int    `gorm:"primary_key;column:FileID"`
	FileName string `gorm:"column:FileName"`
	Hash     string `gorm:"column:Hash"`
}

type ReadRequest struct{
	FileName string `json:"filename"`
}

type Env struct{
	db *gorm.DB
	sh *shell.Shell
}


func main() {

	var err error
	connectionString := fmt.Sprintf("server=%s; user id=%s; password=%s; port=%d; database=%s",
		server, user, password, port, database)
	db, err := gorm.Open("mssql", connectionString)
	if err != nil {
		
		panic(err)
	}
	fmt.Println("db: ",db)
	// ipfs running on
	sh := shell.NewShell("localhost:5001")

	env:=&Env{db:db, sh:sh}
	db.AutoMigrate(&FileInfo{})

	router := gin.Default()

	router.POST("/read", env.readFile)
	router.POST("/write", env.addFile)
	router.GET("/list", env.retrieveAllFiles)

	router.Run(":9090")
}

func (e *Env) readFile(ctx *gin.Context) {
	var fileInfo FileInfo
	var readRequest ReadRequest
	bytedata,err := ioutil.ReadAll(ctx.Request.Body)
	if err!=nil{
		fmt.Println("error in reading data from body, err:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "error in reading data from request body"})
		return
	}
	
	err=json.Unmarshal(bytedata,&readRequest)
	if err!=nil{
		fmt.Println("error in marshalling, Error:",err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "error in unmarshalling"})
		return
	}
	
	// retrieving filehash from db
	e.db.Table("file_infos").Where("FileName=?",readRequest.FileName).First(&fileInfo)
    fmt.Println("fileInfo :",fileInfo)
	if fileInfo.FileID == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "file not found in db"})
		return
	}
    // retrieving data from ipfs with the hash
	closer, err := e.sh.Cat(fileInfo.Hash)
	defer closer.Close()
	if err != nil {
		fmt.Println("Errror in getting data from file, err:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "error in getting data from ipfs"})
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(closer)
	
	fmt.Println("data from cat, " + buf.String())
	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": buf.String()})

}

func (e *Env) addFile(ctx *gin.Context) {
	fileName := ctx.PostForm("filename")
	fileBody := ctx.PostForm("filebody")
	fmt.Println("filename, ",fileName)
	// check if the path is absolute. If yes then it has folder included in the path
	// show error for files with /folder/file
	if filepath.IsAbs(fileName){
	   fmt.Println("invalid path as folder is also included")
	   ctx.JSON(http.StatusBadRequest, gin.H{"status":http.StatusBadRequest, "message":"invalid path as folder is also included"})
	   return
	}
	// adding file data to ipfs
	cid, err := e.sh.Add(strings.NewReader(fileBody))
	if err != nil {
		fmt.Println("Error in adding file to IPFS, Error:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status":http.StatusInternalServerError, "message":"error adding file to ipfs"})
		return
	}
	fmt.Printf("added %s", cid)

	fileInfo := FileInfo{FileName: fileName, Hash: cid}
	fmt.Println("db: ",e.db)

	e.db.Table("file_infos").Where("FileName=?",fileName).First(&fileInfo)
	if fileInfo.FileID==0{
      // adding record for the file in db
	  e.db.Save(&fileInfo)
	  ctx.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "file record added"})
	  return
	}

		fileInfo = FileInfo{}
		// update record
		e.db.Table("file_infos").Where("FileName=?",fileName).First(&fileInfo).Update("Hash",cid)
		ctx.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "file record updated"})
		
}

func (e *Env) retrieveAllFiles(ctx *gin.Context) {

	var allFiles []FileInfo
	var allFileData []string
	e.db.Find(&allFiles)

	if len(allFiles) <= 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "no file retrieved"})
		return
	}

	for _, file := range allFiles {
		closer, err := e.sh.Cat(file.Hash)
		defer closer.Close()
		if err != nil {
			fmt.Println("Errror in getting data from file, err:", err.Error())
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(closer)
		
		//fmt.Println("data from cat, " + buf.String())
		allFileData = append(allFileData, buf.String())
	}
	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": allFileData})
}
