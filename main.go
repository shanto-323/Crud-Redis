package main

import "log"

func main() {
	store, err := NewStore()
	if err != nil {
		log.Fatal(err)
	}

	if err = store.init(); err != nil {
		log.Fatal(err)
	}

	api := NewApi(":8080", *store)
	api.Start()
}
