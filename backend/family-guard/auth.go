package main

import (
    "crypto/subtle"
    "net/http"
    "os"
    "strings"
)

// requireAPIAuth is opt-in: production deployments should set
// FTN_FAMILY_GUARD_API_TOKEN. Local development remains usable when unset.
func requireAPIAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        expected := os.Getenv("FTN_FAMILY_GUARD_API_TOKEN")
        if expected == "" { next.ServeHTTP(w, r); return }
        got := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
        if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
