package main

func main() {
	server := NewAPISever(":8080")
	server.Run()
}
