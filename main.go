package main

func main() {
	server := InitWebServer()
	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
