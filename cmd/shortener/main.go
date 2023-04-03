package main

import (
	"fmt"
	"net/http"
)

func mainPage(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		/*
			var xhr = new XMLHttpRequest();
			var body = 'name=ttttttttttt';
			xhr.open("POST", '/', true);
			xhr.setRequestHeader('Content-Type', 'text/plain');
			//xhr.onreadystatechange = ...;
			xhr.send(body);
		*/

		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/plain")

		/*
				Host        string    // host or host:port
			Path        string    // path (relative paths may omit leading slash)
		*/

		//fmt.Printf("ddddd %+v", req.URL)

		dataBody := req.Body
		result := make([]byte, 10)
		dataBody.Read(result)
		fmt.Printf("wwwww %+v", result)
		fmt.Println(string(result))

		res.Write([]byte("EwHXdJfB"))
		return
	} else {
		res.Write([]byte("Что-то не то!"))
		return
	}
}

func main() {
	muxApp := http.NewServeMux()
	muxApp.HandleFunc(`/`, mainPage)

	err := http.ListenAndServe(`:8080`, muxApp)
	if err != nil {
		panic(err)
	}
}
