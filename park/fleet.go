package park

import (
	"fmt"
	"math"
)

type Inspector interface {
	Inspect(transport Transport) error
}

type ServiceCenter struct{}

func (s ServiceCenter) Inspect(transport Transport) error {
	if !transport.IsServiceable() {
		return fmt.Errorf("транспорт неисправен")
	}
	return nil
}

type Fleet struct {
	items      []InventoryItem
	inspector  Inspector
	nextNumber int
}

func NewFleet(inspector Inspector) *Fleet {
	return &Fleet{inspector: inspector, nextNumber: 1}
}

func (f *Fleet) Add(item InventoryItem) (int, error) {
	if item == nil {
		return 0, fmt.Errorf("объект не указан")
	}
	if item.InventoryNumber() != 0 {
		return 0, fmt.Errorf("объект уже зарегистрирован")
	}
	if consumer, ok := item.(EnergyConsumer); ok && math.IsInf(f.TotalEnergyKWh()+consumer.DailyEnergyKWh(), 0) {
		return 0, fmt.Errorf("суммарный расход энергии слишком большой")
	}

	if transport, ok := item.(Transport); ok {
		if f.inspector == nil {
			return 0, fmt.Errorf("сервис техосмотра не подключён")
		}
		if err := f.inspector.Inspect(transport); err != nil {
			return 0, err
		}
	}

	number := f.nextNumber
	item.assignInventoryNumber(number)
	f.items = append(f.items, item)
	f.nextNumber++
	return number, nil
}

func (f *Fleet) Items() []InventoryItem {
	items := make([]InventoryItem, len(f.items))
	copy(items, f.items)
	return items
}

func (f *Fleet) TransportCount() int {
	count := 0
	for _, item := range f.items {
		if _, ok := item.(Transport); ok {
			count++
		}
	}
	return count
}

func (f *Fleet) TotalEnergyKWh() float64 {
	total := 0.0
	for _, item := range f.items {
		if consumer, ok := item.(EnergyConsumer); ok {
			total += consumer.DailyEnergyKWh()
		}
	}
	return total
}

func (f *Fleet) BeginnerTransport() []Transport {
	var result []Transport
	for _, item := range f.items {
		if transport, ok := item.(Transport); ok && transport.BeginnerScore() >= 6 {
			result = append(result, transport)
		}
	}
	return result
}
