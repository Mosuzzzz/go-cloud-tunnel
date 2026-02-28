package main

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	urls = make(map[string]string)
	mutex = &sync.Mutex{}
)

func main() {
	//Root path to check if it's working 
	http.HandleFunc("/", func(w http.ResponseWriter,r *http.Request){
		fmt.Fprintf(w,"Welcome to your Secure Edge API!\n Use /shorten?Key=x&url=y to create a link.")
	})
	

	//API to create a short link
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request){
		url := r.URL.Query().Get("url")
		key := r.URL.Query().Get("key")
		if url == "" || key == "" {
			http.Error(w, "Missing url or key", 400)
			return
		}

		mutex.Lock()
		urls[key] = url
		mutex.Unlock()
		fmt.Fprintf(w,"Shortened! Your link is: /r/%s", key)
	})

	//Redirect handler
	http.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request){
		key := r.URL.Path[len("/r/"):]
		mutex.Lock()
		target, ok := urls[key]
		mutex.Unlock()
		if !ok {
			http.Error(w, "Short URL not found" , 404)
			return
		}
		http.Redirect(w,r ,target, http.StatusFound)
	})
	

	fmt.Println("URL shortener APT started at :8080")
	http.ListenAndServe(":8080", nil)

}
