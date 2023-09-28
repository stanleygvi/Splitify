package main

import (
	"backend/spotify"
	"fmt"
)

// calculate how many slices of 100 goes into length
func calc_slices(length int) int {
	if length <= 0 {
		return 0
	}

	return (length + 99) / 100
}

func main() {
	id := "01NTza8NIKI7vBIz1jRJD6"
	auth_token := "BQAYOSbvnug0qJV_wqQ98UUPc9EUyxkfQkeQMDMDI6fRnKXHNJGt3AUdSqkBEZhyveZ_G2_kASQv9BKxbaN0VDEaJLY54L-shvpQ1KMSRbcnoigi8r0"
	length := spotify.Get_playlist_length(id, auth_token)
	slices := calc_slices(length)
	for i := 0; i < slices; i++ {
		resp, err := spotify.Get_playlist_children(0, id, auth_token)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(resp)
	}

}
