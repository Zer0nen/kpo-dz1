package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"vyshkat/park"
)

type Console struct {
	fleet *park.Fleet
	input *bufio.Scanner
	out   io.Writer
}

func NewConsole(fleet *park.Fleet, input io.Reader, out io.Writer) *Console {
	return &Console{fleet: fleet, input: bufio.NewScanner(input), out: out}
}

func (c *Console) Run() error {
	fmt.Fprintln(c.out, "ВышКат — учёт парка")
	c.help()
	for {
		fmt.Fprint(c.out, "> ")
		if !c.input.Scan() {
			return c.input.Err()
		}
		args := strings.Fields(c.input.Text())
		if len(args) == 0 {
			continue
		}
		if len(args) == 1 && args[0] == "exit" {
			return nil
		}
		if err := c.execute(args); err != nil {
			fmt.Fprintln(c.out, "Ошибка:", err)
		}
	}
}

func (c *Console) execute(args []string) error {
	command := args[0]
	switch command {
	case "help", "report", "beginners":
		if len(args) != 1 {
			return fmt.Errorf("команда %s не принимает параметры", command)
		}
		switch command {
		case "help":
			c.help()
		case "report":
			items := c.fleet.Items()
			fmt.Fprintln(c.out, "Транспорт:", c.fleet.TransportCount())
			fmt.Fprintln(c.out, "Вещи:", len(items)-c.fleet.TransportCount())
			fmt.Fprintln(c.out, "Всего объектов:", len(items))
			fmt.Fprintln(c.out, "Энергия за сутки:", c.fleet.TotalEnergyKWh(), "кВт·ч")
			for _, item := range items {
				fmt.Fprint(c.out, item.InventoryNumber(), " ", item.Name())
				if consumer, ok := item.(park.EnergyConsumer); ok {
					fmt.Fprint(c.out, " — ", consumer.DailyEnergyKWh(), " кВт·ч/сутки")
				}
				fmt.Fprintln(c.out)
			}
		case "beginners":
			items := c.fleet.BeginnerTransport()
			if len(items) == 0 {
				fmt.Fprintln(c.out, "Транспорта для новичков нет")
			}
			for _, item := range items {
				fmt.Fprintln(c.out, item.InventoryNumber(), item.Name(), item.BeginnerScore())
			}
		}
		return nil
	}

	var item park.InventoryItem
	var err error
	switch command {
	case "bicycle", "scooter", "electric-scooter", "electric-bicycle":
		electric := command == "electric-scooter" || command == "electric-bicycle"
		count := 3
		if electric {
			count = 4
		}
		if len(args) != count {
			return fmt.Errorf("используйте формат из help")
		}
		ease, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("простота должна быть целым числом")
		}
		if args[2] != "true" && args[2] != "false" {
			return fmt.Errorf("исправность задаётся как true или false")
		}
		working := args[2] == "true"
		if electric {
			energy, parseErr := strconv.ParseFloat(args[3], 64)
			if parseErr != nil {
				return fmt.Errorf("расход энергии должен быть числом")
			}
			if command == "electric-scooter" {
				item, err = park.NewElectricScooter(ease, working, energy)
			} else {
				item, err = park.NewElectricBicycle(ease, working, energy)
			}
		} else if command == "bicycle" {
			item, err = park.NewBicycle(ease, working)
		} else {
			item, err = park.NewScooter(ease, working)
		}
		if err != nil {
			return err
		}
	case "helmet":
		if len(args) != 1 {
			return fmt.Errorf("команда helmet не принимает параметры")
		}
		item = park.NewHelmet()
	case "station", "dock", "cabinet":
		if len(args) != 2 {
			return fmt.Errorf("формат: %s <кВт·ч за сутки>", command)
		}
		energy, parseErr := strconv.ParseFloat(args[1], 64)
		if parseErr != nil {
			return fmt.Errorf("расход энергии должен быть числом")
		}
		switch command {
		case "station":
			item, err = park.NewChargingStation(energy)
		case "dock":
			item, err = park.NewDockStation(energy)
		case "cabinet":
			item, err = park.NewChargingCabinet(energy)
		}
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("неизвестная команда; введите help")
	}

	number, err := c.fleet.Add(item)
	if err != nil {
		return err
	}
	fmt.Fprintln(c.out, "Принято:", number, item.Name())
	return nil
}

func (c *Console) help() {
	fmt.Fprintln(c.out, "bicycle <простота 1–10> <исправен true/false>")
	fmt.Fprintln(c.out, "scooter <простота 1–10> <исправен true/false>")
	fmt.Fprintln(c.out, "electric-scooter <простота 1–10> <исправен true/false> <кВт·ч за сутки>")
	fmt.Fprintln(c.out, "electric-bicycle <простота 1–10> <исправен true/false> <кВт·ч за сутки>")
	fmt.Fprintln(c.out, "helmet — добавить шлем")
	fmt.Fprintln(c.out, "station <кВт·ч за сутки> — добавить зарядную станцию")
	fmt.Fprintln(c.out, "dock <кВт·ч за сутки> — добавить док-станцию")
	fmt.Fprintln(c.out, "cabinet <кВт·ч за сутки> — добавить зарядный шкаф")
	fmt.Fprintln(c.out, "report — количество, энергия и весь инвентарь")
	fmt.Fprintln(c.out, "beginners — транспорт для новичков")
	fmt.Fprintln(c.out, "help — справка; exit — выход")
}
