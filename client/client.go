package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	data := map[string]string{
		"email":    "Abuba@exemple.com",
		"password": "qwer1234",
	}

	json_data, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	resp, err := http.Post(
		"http://localhost:8080/auth/students",
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	token := resp.Header.Get("Token")
	fmt.Println(token)

}
