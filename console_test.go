package main

import (
	"bytes"
	"strings"
	"testing"

	"vyshkat/park"
)

func TestConsole(t *testing.T) {
	container := &Container{}
	firstFleet := container.Fleet()
	secondFleet := container.Fleet()
	if firstFleet != secondFleet {
		t.Fatal("контейнер должен возвращать один парк")
	}
	input := "\nreport\nbeginners\nbicycle 5 true\nscooter 7 true\nelectric-scooter 6 true 1.5\nelectric-bicycle 8 true 2\nhelmet\nstation 0.5\ndock 0.25\ncabinet 0.75\nelectric-scooter 8 false 1\nreport\nbeginners\nhelp\nexit\nhelmet\n"
	var output bytes.Buffer
	console := NewConsole(container.Fleet(), strings.NewReader(input), &output)
	if err := console.Run(); err != nil {
		t.Fatal(err)
	}
	if len(container.Fleet().Items()) != 8 || container.Fleet().TotalEnergyKWh() != 5 {
		t.Fatal("консоль неверно изменила парк")
	}
	for _, expected := range []string{"транспорт неисправен", "Транспорт: 4", "Вещи: 4", "Всего объектов: 8", "Энергия за сутки: 5 кВт·ч", "Транспорта для новичков нет", "2 Самокат 7", "3 Электросамокат 6", "4 Электровелосипед 8", "3 Электросамокат — 1.5 кВт·ч/сутки", "6 Зарядная станция — 0.5 кВт·ч/сутки", "7 Док-станция — 0.25 кВт·ч/сутки", "8 Зарядный шкаф — 0.75 кВт·ч/сутки"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("в выводе нет %q", expected)
		}
	}
}

func TestInvalidConsoleInput(t *testing.T) {
	commands := []string{
		"unknown", "report extra", "beginners extra", "help extra", "exit extra",
		"bicycle", "bicycle text true", "bicycle 6 yes", "bicycle 11 true",
		"scooter 6 true 1", "scooter 0 true", "electric-scooter 6 true", "electric-scooter 6 true text", "electric-scooter 6 true NaN",
		"electric-bicycle 6 true", "electric-bicycle 6 true text", "electric-bicycle 6 true -1",
		"helmet extra", "station", "station text", "station -1", "dock", "dock text", "dock -1", "cabinet", "cabinet text", "cabinet -1",
	}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			fleet := park.NewFleet(park.ServiceCenter{})
			fleet.Add(park.NewHelmet())
			var output bytes.Buffer
			console := NewConsole(fleet, strings.NewReader(command+"\n"), &output)
			if err := console.Run(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), "Ошибка:") || len(fleet.Items()) != 1 || fleet.Items()[0].InventoryNumber() != 1 {
				t.Fatal("неверная команда должна давать ошибку и сохранять парк")
			}
		})
	}
}

func TestConsoleRejectsEveryBrokenTransport(t *testing.T) {
	for _, kind := range []string{"scooter", "bicycle", "electric-scooter", "electric-bicycle"} {
		t.Run(kind, func(t *testing.T) {
			command := kind + " 8 false"
			if strings.HasPrefix(kind, "electric-") {
				command += " 1"
			}
			fleet := park.NewFleet(park.ServiceCenter{})
			var output bytes.Buffer
			console := NewConsole(fleet, strings.NewReader(command+"\nhelmet\nexit\n"), &output)
			if err := console.Run(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), "транспорт неисправен") || fleet.TransportCount() != 0 || len(fleet.Items()) != 1 || fleet.Items()[0].InventoryNumber() != 1 {
				t.Fatal("неисправный транспорт не должен занимать место и номер в парке")
			}
		})
	}
}
