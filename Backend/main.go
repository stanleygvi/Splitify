package main

import (
	"backend/spotify"
	"fmt"
)

func main() {
	id := "01NTza8NIKI7vBIz1jRJD6"
	auth_token := "BQDN_qn0pw7XyHUyvSgR9Y1cdrFkMc_tZKme_V5KHEmqALeQBJidIQMxIhiWJDLV08XZFNl_NMmRj_qqEON6FPC5uD01bkJIiD8EF6XqULg8xUE-e5w"
	resp, err := spotify.Get_playlist(id, auth_token)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(resp)
	// sorted_lists := openai.Send()
	// fmt.Println(sorted_lists)

}
