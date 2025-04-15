package pkg

import "time"

type ServiceName string

type OrderState string

const (
	ServiceBroadband ServiceName = "broadband"
	ServiceEnergy    ServiceName = "energy"
	ServiceInsurance ServiceName = "insurance"

	OrderStatePlaced OrderState = "order_state_placed"
	OrderStateSent   OrderState = "order_state_sent"
)

type Order struct {
	ServiceName string    `json:"service_name"`
	Customer    Customer  `json:"customer"`
	Timestamp   time.Time `json:"timestamp"`
	State       OrderState
}

type Customer struct {
	Address        Address `json:"address"`
	CustomerNumber int64   `json:"customer_number,omitempty"`
	Email          string  `json:"email,omitempty"`
}

type Address struct {
	FirstLine string `json:"first_line,omitempty"`
	City      string `json:"city,omitempty"`
	Postcode  string `json:"postcode,omitempty"`
}
