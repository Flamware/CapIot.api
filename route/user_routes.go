package route

import "net/http"

func SetupUserRoutes(mux *http.ServeMux) {
	// Example location endpoint
	mux.HandleFunc("/bind-device/{deviceID}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a location endpoint"))
	})
}
