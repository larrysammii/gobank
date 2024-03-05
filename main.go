package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store Storage, fName, lName, pw string) *Account {

	acc, err := NewAccount(fName, lName, pw)

	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new account => ", acc.Number)
	return acc

}

func seedAccounts(s Storage) {
	seedAccount(s, "cock", "penis", "fuckfuck")
}

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}

	server := NewAPISever(":8080", store)
	server.Run()
}
