package models

// Orde
type OrderWebhookEvent struct {
	WebhookEvent        // Embedded base model
	OrderID      string `json:"order_id"`
}

// Extending the Process method specific to order events
func (owe *OrderWebhookEvent) Process() error {
	// Call the base process method if needed
	if err := owe.WebhookEvent.Process(); err != nil {
		return err
	}

	// Add specific logic for processing order webhook
	// Example: Update order status, validate order data, etc.
	return nil
}
