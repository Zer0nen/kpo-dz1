package park_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"vyshkat/park"
)

type fixedInspector struct {
	err   error
	calls int
}

func (i *fixedInspector) Inspect(park.Transport) error {
	i.calls++
	return i.err
}

var transportFactories = []struct {
	name     string
	create   func(int, bool) (park.Transport, error)
	electric bool
}{
	{"Самокат", func(ease int, working bool) (park.Transport, error) { return park.NewScooter(ease, working) }, false},
	{"Велосипед", func(ease int, working bool) (park.Transport, error) { return park.NewBicycle(ease, working) }, false},
	{"Электросамокат", func(ease int, working bool) (park.Transport, error) {
		return park.NewElectricScooter(ease, working, 1.5)
	}, true},
	{"Электровелосипед", func(ease int, working bool) (park.Transport, error) {
		return park.NewElectricBicycle(ease, working, 1.5)
	}, true},
}

func TestConstructors(t *testing.T) {
	for _, factory := range transportFactories {
		t.Run(factory.name, func(t *testing.T) {
			for _, ease := range []int{1, 6, 10} {
				for _, working := range []bool{false, true} {
					transport, err := factory.create(ease, working)
					if err != nil {
						t.Fatal(err)
					}
					if transport.Name() != factory.name || transport.BeginnerScore() != ease || transport.IsServiceable() != working || transport.InventoryNumber() != 0 {
						t.Fatal("характеристики нового транспорта не совпадают с заданными")
					}
					consumer, electric := transport.(park.EnergyConsumer)
					if electric != factory.electric || electric && consumer.DailyEnergyKWh() != 1.5 {
						t.Fatal("неверный учёт энергопотребления транспорта")
					}
				}
			}
			for _, ease := range []int{-1, 0, 11} {
				if _, err := factory.create(ease, true); err == nil {
					t.Fatal("принята неверная простота транспорта")
				}
			}
		})
	}

	energyFactories := []func(float64) (park.EnergyConsumer, error){
		func(energy float64) (park.EnergyConsumer, error) { return park.NewElectricScooter(6, true, energy) },
		func(energy float64) (park.EnergyConsumer, error) { return park.NewElectricBicycle(6, true, energy) },
		func(energy float64) (park.EnergyConsumer, error) { return park.NewChargingStation(energy) },
		func(energy float64) (park.EnergyConsumer, error) { return park.NewDockStation(energy) },
		func(energy float64) (park.EnergyConsumer, error) { return park.NewChargingCabinet(energy) },
	}
	for _, create := range energyFactories {
		for _, energy := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
			if _, err := create(energy); err == nil {
				t.Fatal("принят неверный расход энергии")
			}
		}
		for _, energy := range []float64{0, 0.5} {
			item, err := create(energy)
			if err != nil || item.DailyEnergyKWh() != energy {
				t.Fatal("не удалось создать объект с допустимым расходом")
			}
		}
	}
}

func TestFleet(t *testing.T) {
	fleet := park.NewFleet(park.ServiceCenter{})
	if fleet.TransportCount() != 0 || fleet.TotalEnergyKWh() != 0 || len(fleet.Items()) != 0 {
		t.Fatal("новый парк должен быть пустым")
	}
	bicycle, _ := park.NewBicycle(5, true)
	scooter, _ := park.NewScooter(7, true)
	electricScooter, _ := park.NewElectricScooter(6, true, 1.5)
	electricBicycle, _ := park.NewElectricBicycle(8, true, 2)
	station, _ := park.NewChargingStation(0.5)
	dock, _ := park.NewDockStation(0.25)
	cabinet, _ := park.NewChargingCabinet(0.75)
	expected := []park.InventoryItem{bicycle, scooter, electricScooter, electricBicycle, park.NewHelmet(), station, dock, cabinet}
	for index, item := range expected {
		number, err := fleet.Add(item)
		if err != nil || number != index+1 || item.InventoryNumber() != number {
			t.Fatalf("неверная регистрация объекта %d: %v", index+1, err)
		}
	}
	if fleet.TransportCount() != 4 || len(fleet.Items()) != 8 || fleet.TotalEnergyKWh() != 5 {
		t.Fatal("неверный отчёт по парку")
	}
	for index, item := range fleet.Items() {
		if item != expected[index] || item.Name() == "" {
			t.Fatal("инвентарь не сохранил принятые объекты")
		}
	}
	items := fleet.Items()
	items[0] = nil
	if fleet.Items()[0] != bicycle {
		t.Fatal("изменение полученного списка изменило парк")
	}
	if _, err := fleet.Add(scooter); err == nil || len(fleet.Items()) != 8 {
		t.Fatal("повторная регистрация должна отклоняться")
	}
	if _, err := fleet.Add(nil); err == nil {
		t.Fatal("пустой объект должен отклоняться")
	}
	broken, _ := park.NewBicycle(8, false)
	if _, err := fleet.Add(broken); err == nil || broken.InventoryNumber() != 0 || len(fleet.Items()) != 8 {
		t.Fatal("неисправный транспорт не должен менять парк и получать номер")
	}
	if number, err := fleet.Add(park.NewHelmet()); err != nil || number != 9 {
		t.Fatal("отказ не должен занимать инвентарный номер")
	}
}

func TestBeginnerBoundary(t *testing.T) {
	fleet := park.NewFleet(park.ServiceCenter{})
	for _, ease := range []int{5, 6} {
		for _, factory := range transportFactories {
			item, err := factory.create(ease, true)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := fleet.Add(item); err != nil {
				t.Fatal(err)
			}
		}
	}
	fleet.Add(park.NewHelmet())
	result := fleet.BeginnerTransport()
	if len(result) != 4 {
		t.Fatal("для новичков должны подходить четыре объекта с простотой 6")
	}
	for _, transport := range result {
		if transport.BeginnerScore() != 6 {
			t.Fatal("транспорт с простотой 5 не должен подходить новичкам")
		}
	}
}

func TestInspectorDecision(t *testing.T) {
	refusal := errors.New("отказ тестового инспектора")
	for _, factory := range transportFactories {
		for _, working := range []bool{false, true} {
			for _, accept := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/working=%t/accept=%t", factory.name, working, accept), func(t *testing.T) {
					inspector := &fixedInspector{}
					if !accept {
						inspector.err = refusal
					}
					fleet := park.NewFleet(inspector)
					helmet := park.NewHelmet()
					fleet.Add(helmet)
					transport, _ := factory.create(6, working)
					number, err := fleet.Add(transport)
					nextNumber := 2
					if accept {
						if err != nil || number != 2 || transport.InventoryNumber() != 2 || len(fleet.Items()) != 2 || fleet.TransportCount() != 1 {
							t.Fatal("решение принять транспорт не выполнено")
						}
						nextNumber = 3
					} else if !errors.Is(err, refusal) || number != 0 || transport.InventoryNumber() != 0 || len(fleet.Items()) != 1 || fleet.TransportCount() != 0 {
						t.Fatal("отказ должен сохранять парк и причину ошибки")
					}
					if inspector.calls != 1 || fleet.Items()[0] != helmet || helmet.InventoryNumber() != 1 {
						t.Fatal("техосмотр должен вызываться только для нового транспорта")
					}
					if number, err := fleet.Add(park.NewHelmet()); err != nil || number != nextNumber {
						t.Fatal("нарушен порядок инвентарных номеров")
					}
				})
			}
		}
	}
	fleet := park.NewFleet(nil)
	bicycle, _ := park.NewBicycle(6, true)
	if _, err := fleet.Add(bicycle); err == nil || bicycle.InventoryNumber() != 0 || len(fleet.Items()) != 0 {
		t.Fatal("транспорт нельзя принять без техосмотра")
	}
}

func TestEnergyOverflowDoesNotChangeFleet(t *testing.T) {
	fleet := park.NewFleet(park.ServiceCenter{})
	first, _ := park.NewChargingStation(math.MaxFloat64)
	second, _ := park.NewChargingStation(math.MaxFloat64)
	if _, err := fleet.Add(first); err != nil {
		t.Fatal(err)
	}
	if number, err := fleet.Add(second); err == nil || number != 0 || second.InventoryNumber() != 0 {
		t.Fatal("объект с переполнением суммы не должен регистрироваться")
	}
	if len(fleet.Items()) != 1 || fleet.TotalEnergyKWh() != math.MaxFloat64 {
		t.Fatal("ошибка суммы не должна менять парк")
	}
	if number, err := fleet.Add(park.NewHelmet()); err != nil || number != 2 {
		t.Fatal("отказ не должен занимать номер")
	}
}
