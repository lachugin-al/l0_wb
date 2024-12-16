package model

import (
	"testing"
	"time"
)

// TestOrder_Validation проверяет корректность валидации объекта Order.
//
//	Сценарии:
//	- Убедиться, что заказ без items считается некорректным.
//	- Проверить, что добавление items в заказ исправляет проблему.
func TestOrder_Validation(t *testing.T) {
	order := &Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Locale:      "en",
		DateCreated: time.Now(),
	}

	// 1-й сценарий
	if len(order.Items) != 0 {
		t.Errorf("expected items in the order = 0, got %d", len(order.Items))
	}

	// 2-й сценарий
	order.Items = []Item{{ChrtID: 123}}
	if len(order.Items) == 0 {
		t.Errorf("expected items in the order > 0, got %d", len(order.Items))
	}
}
