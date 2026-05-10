package main

import (
	"api/handlers"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	// Security Service
	mux.HandleFunc("POST /api/v1/security/register", handlers.Register)
	mux.HandleFunc("POST /api/v1/security/login", handlers.Login)
	mux.Handle("GET /api/v1/security/me", handlers.SecSrv.RequireAuth(http.HandlerFunc(handlers.Me)))
	mux.Handle("GET /api/v1/security/users/{user_id}", handlers.SecSrv.RequireAuth(http.HandlerFunc(handlers.GetUserById)))
	mux.Handle("POST /api/v1/security/users/{user_id}/set_role/{role_id}",
		handlers.SecSrv.RequireAuth(handlers.SecSrv.AllowForRole(5, http.HandlerFunc(handlers.SetRoleForUser))))
	mux.Handle("GET /api/v1/security/roles/{role_id}",
		handlers.SecSrv.RequireAuth(handlers.SecSrv.AllowForRole(4, http.HandlerFunc(handlers.SetRoleForUser))))

	// XRay Service
	mux.Handle("GET /api/v1/xray/config", handlers.SecSrv.RequireAuth(handlers.SecSrv.AllowForRole(4, http.HandlerFunc(handlers.GetConfig))))
	mux.Handle("GET /api/v1/xray/clients/by_user_id/{user_id}", handlers.SecSrv.RequireAuth(http.HandlerFunc(handlers.GetClientsByUserId)))
	mux.Handle("GET /api/v1/xray/clients/{client_id}", handlers.SecSrv.RequireAuth(http.HandlerFunc(handlers.GetClientById)))
	mux.Handle("GET /api/v1/xray/clients/link/{client_id}", handlers.SecSrv.RequireAuth(http.HandlerFunc(handlers.GetXrayLinkById)))

	// Invite Service
	mux.Handle("POST /api/v1/invite/new",
		handlers.SecSrv.RequireAuth(handlers.SecSrv.AllowForRole(2, http.HandlerFunc(handlers.NewInvite))))
	mux.Handle("POST /api/v1/invite/activate", handlers.SecSrv.RequireAuth(http.HandlerFunc(handlers.ActivateInvite)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: withRecover(mux),
	}

	log.Println("listening on", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
