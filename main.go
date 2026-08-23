package main

import (
	"fmt"
	"net/http"
)

const ruta_media = "./static"

func main() {
	fs := http.FileServer(http.Dir(ruta_media))
	http.Handle("/", fs)
	fmt.Printf("Andando")
	http.ListenAndServe(":8080", nil)
}
