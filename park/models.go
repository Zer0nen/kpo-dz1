package park

import (
	"fmt"
	"math"
)

type InventoryItem interface {
	Name() string
	InventoryNumber() int
	assignInventoryNumber(int)
}

type EnergyConsumer interface {
	DailyEnergyKWh() float64
}

type Transport interface {
	InventoryItem
	BeginnerScore() int
	IsServiceable() bool
}

type asset struct {
	name   string
	number int
}

func (a asset) Name() string {
	return a.name
}

func (a asset) InventoryNumber() int {
	return a.number
}

func (a *asset) assignInventoryNumber(number int) {
	a.number = number
}

type vehicle struct {
	asset
	ease    int
	working bool
}

func newVehicle(name string, ease int, working bool) (vehicle, error) {
	if ease < 1 || ease > 10 {
		return vehicle{}, fmt.Errorf("простота для новичка должна быть от 1 до 10")
	}
	return vehicle{asset: asset{name: name}, ease: ease, working: working}, nil
}

func (v vehicle) BeginnerScore() int {
	return v.ease
}

func (v vehicle) IsServiceable() bool {
	return v.working
}

type energyUsage struct {
	energyKWh float64
}

func (e energyUsage) DailyEnergyKWh() float64 {
	return e.energyKWh
}

type Scooter struct {
	vehicle
}

func NewScooter(ease int, working bool) (*Scooter, error) {
	base, err := newVehicle("Самокат", ease, working)
	if err != nil {
		return nil, err
	}
	return &Scooter{vehicle: base}, nil
}

type Bicycle struct {
	vehicle
}

func NewBicycle(ease int, working bool) (*Bicycle, error) {
	base, err := newVehicle("Велосипед", ease, working)
	if err != nil {
		return nil, err
	}
	return &Bicycle{vehicle: base}, nil
}

type ElectricScooter struct {
	vehicle
	energyUsage
}

func NewElectricScooter(ease int, working bool, energy float64) (*ElectricScooter, error) {
	base, err := newVehicle("Электросамокат", ease, working)
	if err != nil {
		return nil, err
	}
	if err := validateEnergy(energy); err != nil {
		return nil, err
	}
	return &ElectricScooter{vehicle: base, energyUsage: energyUsage{energyKWh: energy}}, nil
}

type ElectricBicycle struct {
	vehicle
	energyUsage
}

func NewElectricBicycle(ease int, working bool, energy float64) (*ElectricBicycle, error) {
	base, err := newVehicle("Электровелосипед", ease, working)
	if err != nil {
		return nil, err
	}
	if err := validateEnergy(energy); err != nil {
		return nil, err
	}
	return &ElectricBicycle{vehicle: base, energyUsage: energyUsage{energyKWh: energy}}, nil
}

type Helmet struct {
	asset
}

func NewHelmet() *Helmet {
	return &Helmet{asset: asset{name: "Шлем"}}
}

type ChargingStation struct {
	asset
	energyUsage
}

func NewChargingStation(energy float64) (*ChargingStation, error) {
	if err := validateEnergy(energy); err != nil {
		return nil, err
	}
	return &ChargingStation{asset: asset{name: "Зарядная станция"}, energyUsage: energyUsage{energyKWh: energy}}, nil
}

type DockStation struct {
	asset
	energyUsage
}

func NewDockStation(energy float64) (*DockStation, error) {
	if err := validateEnergy(energy); err != nil {
		return nil, err
	}
	return &DockStation{asset: asset{name: "Док-станция"}, energyUsage: energyUsage{energyKWh: energy}}, nil
}

type ChargingCabinet struct {
	asset
	energyUsage
}

func NewChargingCabinet(energy float64) (*ChargingCabinet, error) {
	if err := validateEnergy(energy); err != nil {
		return nil, err
	}
	return &ChargingCabinet{asset: asset{name: "Зарядный шкаф"}, energyUsage: energyUsage{energyKWh: energy}}, nil
}

func validateEnergy(energy float64) error {
	if energy < 0 || math.IsNaN(energy) || math.IsInf(energy, 0) {
		return fmt.Errorf("расход энергии должен быть конечным неотрицательным числом")
	}
	return nil
}
