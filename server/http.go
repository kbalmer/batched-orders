package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"unicode"

	"batched-orders/batcher"
	"batched-orders/pkg"
)

type Handler struct {
	batcher *batcher.Batcher
}

func NewHandler(batcher *batcher.Batcher) *Handler {
	return &Handler{
		batcher: batcher,
	}
}

func (h *Handler) ProcessOrder(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var order pkg.Order
	if err := json.Unmarshal(b, &order); err != nil {
		log.Printf("Error unmarshalling body: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := validateOrder(order); err != nil {
		log.Printf("Error validating order: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.batcher.Batch(order)

	w.WriteHeader(http.StatusOK)
}

var validServices = []pkg.ServiceName{pkg.ServiceBroadband, pkg.ServiceInsurance, pkg.ServiceEnergy}

func validateOrder(order pkg.Order) error {
	if !slices.Contains(validServices, pkg.ServiceName(order.ServiceName)) {
		return fmt.Errorf("service %s is not a valid service", order.ServiceName)
	}

	postcode := order.Customer.Address.Postcode

	if len(postcode) > 8 {
		return fmt.Errorf("invalid postcode %s provided - over maximum 8 characters", order.Customer.Address.Postcode)
	}

	for _, r := range postcode {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return fmt.Errorf("invalid postcode %s provided - must contain letters, spaces and numbers only", postcode)
		}
	}

	return nil
}
