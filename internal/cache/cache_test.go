package cache

import (
	"testing"

	"l0_wb/internal/model"
	"l0_wb/internal/util"
)

// TestOrderCache проверяет базовые операции (Set и Get) работы с OrderCache.
func TestOrderCache(t *testing.T) {
	err := util.InitLogger()
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	defer util.SyncLogger()

	cache := NewOrderCache()

	// Добавляем тестовый заказ в кэш
	order := &model.Order{OrderUID: "test_uid"}
	cache.Set(order)

	// Проверяем, что заказ успешно добавлен в кэш
	got := cache.Get("test_uid")
	if got == nil {
		t.Fatalf("expected order in cache, got nil")
	}
	if got.OrderUID != "test_uid" {
		t.Errorf("expected order UID test_uid, got %s", got.OrderUID)
	}

	// Проверяем, что запрос несуществующего UID возвращает nil
	nonexistent := cache.Get("nonexistent_uid")
	if nonexistent != nil {
		t.Errorf("expected nil for nonexistent UID, got %v", nonexistent)
	}

	// Проверяем обновление данных в кэше
	updatedOrder := &model.Order{OrderUID: "test_uid", TrackNumber: "updated_track"}
	cache.Set(updatedOrder)
	updatedGot := cache.Get("test_uid")
	if updatedGot == nil {
		t.Fatalf("expected updated order in cache, got nil")
	}
	if updatedGot.TrackNumber != "updated_track" {
		t.Errorf("expected updated TrackNumber updated_track, got %s", updatedGot.TrackNumber)
	}
}
