package webapp

import (
	_ "embed"
)

//go:embed index.html
var app []byte

//go:embed favicon.ico
var favicon []byte

//go:embed icon.png
var icon []byte

//go:embed style.css
var css []byte

//go:embed app.js
var js []byte

func GetApp() []byte {
	return app
}

func GetFavicon() []byte {
	return favicon
}

func GetIcon() []byte {
	return icon
}

func GetCSS() []byte {
	return css
}

func GetJS() []byte {
	return js
}
