package main

import(
	"fmt"
	"math/rand"
    "strings"
//"bytes"
	"strconv"
	"net/http"
	"io/ioutil"
	"net/url"
)



func main(){
    n:=100
	urls:=[]string{"http://localhost:9090/read",
				   "http://localhost:9090/write",
				   "http://localhost:9090/list"}

	ch:=make(chan string)

	for i:=0;i<n;i++{
        x:=rand.Intn(3)
		go MakeRequest(i,x, urls[x],ch)
	}

	for i:=0;i<n;i++{
		fmt.Println(<-ch)
	}

}

func MakeRequest(i int,x int,apiUrl string , ch chan<-string){
   if x==0 {

	   input :=`
	   {
		   "filename":"Hello%s.txt"
	   }`
	   input=fmt.Sprintf(input,strconv.Itoa(i))
   
	   requestBody:=strings.NewReader(input)

	   resp,err:=http.Post(apiUrl,"application/json;charset=UTF-8",requestBody)
	   if err!=nil{
		ch<-fmt.Sprintf("error in getting response from url %s ,error:%s",apiUrl,err.Error())
	    return	
	}
	   defer resp.Body.Close()
	   body,_:=ioutil.ReadAll(resp.Body)
	   ch<- fmt.Sprintf("value returned from %s is %s",apiUrl,string(body))
   }else if x==1{
	fmt.Println("i :",i)
	   inputFileName:= fmt.Sprintf("hello%s.txt",strconv.Itoa(i))
	   inputFileBody:=fmt.Sprintf("%s how are you", strconv.Itoa(i))
		resp,err:=http.PostForm(apiUrl, url.Values{"filename":{inputFileName}, "filebody":{inputFileBody}})
		if err!=nil{
			ch<-fmt.Sprintf("error in getting response from url %s ,error:%s",apiUrl,err.Error())
             return
		}
		defer resp.Body.Close()
		body,_:=ioutil.ReadAll(resp.Body)
		ch<- fmt.Sprintf("value returned from %s is %s",apiUrl,string(body))
   }else{
	fmt.Println("i :",i)
		resp,err:=http.Get(apiUrl)
		if err!=nil{
                 ch<-fmt.Sprintf("error in getting response from url %s ,error:%s",apiUrl,err.Error())
		         return
				}
	    defer resp.Body.Close()
	    body,_:=ioutil.ReadAll(resp.Body)
	    ch<- fmt.Sprintf("value returned from %s is %s",apiUrl,string(body))
   }
}