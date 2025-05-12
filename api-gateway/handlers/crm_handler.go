package handlers

import (
    "api-gateway/proxy"
    "log"

    "github.com/gorilla/mux"
)

// RegisterCRMRoutes registers routes for the CRM service
func RegisterCRMRoutes(router *mux.Router, crmServiceURL string) {
    crmProxy, err := proxy.NewReverseProxy(crmServiceURL, "/api/crm")
    if err != nil {
        log.Fatalf("Failed to create CRM service proxy: %v", err)
    }
    
    // Customer endpoints
    router.PathPrefix("/customers").Handler(crmProxy)
    
    // Ticket endpoints
    router.PathPrefix("/tickets").Handler(crmProxy)
    
    // Agent endpoints
    router.PathPrefix("/agents").Handler(crmProxy)
    
    // Analytics endpoints
    router.PathPrefix("/analytics").Handler(crmProxy)
}