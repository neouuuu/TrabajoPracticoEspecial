package main

import (
	"net/http"
)

const ruta_media = "./static"

func main() {
	fs := http.FileServer(http.Dir(ruta_media))
	http.Handle("/", fs)
	http.ListenAndServe(":8080", nil)
}
