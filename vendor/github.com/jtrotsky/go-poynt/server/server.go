package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jtrotsky/go-poynt/poyntcloud/auth"
	"github.com/jtrotsky/go-poynt/poyntcloud/config"
)

// Run starts our webserver.
func Run() {
	// TODO: Auth should only happen when prompted.

	// Get config first as auth relies on it.
	config, err := config.GetConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
	}
	auth, err := auth.GetAuth(config)
	if err != nil {
		fmt.Println("Error getting auth:", err)
	}
	manager := NewManager(auth, config)

	http.HandleFunc("/", manager.Gateway)          // Has transaction status info.
	http.HandleFunc("/callback", manager.Callback) // To receive payment responses.
	http.HandleFunc("/pay", manager.Pay)           // To send payments.

	http.Handle(
		"/server/assets/",
		http.StripPrefix("/server/assets/", http.FileServer(http.Dir("server/assets/"))),
	)

	// Create webserver on localhost port 8000.
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
