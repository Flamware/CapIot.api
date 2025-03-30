package route

import (
	"net/http"
)

// SetupLocationRoutes initializes the location-related routes
func SetupLocationRoutes(mux *http.ServeMux) {
	// Example location endpoint
	mux.HandleFunc("/location", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a location endpoint"))
	})
}
