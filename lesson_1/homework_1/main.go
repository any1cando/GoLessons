package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Door Создали структуру Дверь, чтобы делать переход между улицей/коридором, т.к. на этом завязаны тесты
type Door struct {
	IsOpen bool
}

// Room Создали структуру Комната, в которой лежит название комнаты, предметы, которые в ней лежат, информация, которую
// можно узнать, находясь в этой комнате и набор мест, куда можно из неё пойти
type Room struct {
	Name               string
	Items              map[string][]string
	InformationAround  string
	InformationOnEnter string
	EmptyInformation   string
	AdditionalLookInfo func(*Player) string
	PlacesToGo         []*Room
	Door               *Door
}

// Player Создали структуру Игрока, который имеет текущую локацию, бул поле на наличие рюкзака и предметы, которые
// находятся в этом рюкзаке
type Player struct {
	Location    *Room
	HasBackpack bool
	Items       []string
}

func (player *Player) ApplyItem(item string, destination string) string {

	if !slices.Contains(player.Items, item) {
		return "нет предмета в инвентаре - " + item
	}

	if destination != "дверь" || player.Location.Door == nil {
		return "не к чему применить"
	}

	if item != "ключи" {
		return "не к чему применить"
	}
	player.Location.Door.IsOpen = true // Достаточно одной двери, т.к. у них общий указатель
	return "дверь открыта"

}

func (player *Player) TakeItem(itemWeWannaTake string) string {

	var allItemsInLocation []string
	var placeWithItem string
	var itemsInPlace []string

	for place, items := range player.Location.Items {
		for _, item := range items {
			allItemsInLocation = append(allItemsInLocation, item)
			if item == itemWeWannaTake {
				// сохраняем ключ и сам слайс предметов из комнаты, чтобы потом его отредактировать
				placeWithItem = place
				itemsInPlace = items
			}
		}
	}

	if !slices.Contains(allItemsInLocation, itemWeWannaTake) {
		return "нет такого"
	}

	if !player.HasBackpack {
		return "некуда класть"
	}

	player.Items = append(player.Items, itemWeWannaTake)
	indexOfItemWeWannaDelete := slices.Index(itemsInPlace, itemWeWannaTake)
	itemsInPlace = slices.Delete(itemsInPlace, indexOfItemWeWannaDelete, indexOfItemWeWannaDelete+1)
	if len(itemsInPlace) == 0 {
		delete(player.Location.Items, placeWithItem)
	} else {
		player.Location.Items[placeWithItem] = itemsInPlace
	}
	return "предмет добавлен в инвентарь: " + itemWeWannaTake
}

func (player *Player) PutOnItem(itemPutOn string) string {
	if player.HasBackpack && itemPutOn == "рюкзак" {
		return "рюкзак уже надет"
	}

	var allItemsInLocation []string

	for _, items := range player.Location.Items {
		for _, item := range items {
			allItemsInLocation = append(allItemsInLocation, item)
		}
	}

	if slices.Contains(allItemsInLocation, itemPutOn) {
		if itemPutOn == "рюкзак" {
			player.HasBackpack = true
			var placeWithBackpack string
			var sliceHasBackpack []string
		OuterLoop:
			for place, items := range player.Location.Items {
				for _, item := range items {
					if item == "рюкзак" {
						placeWithBackpack = place
						sliceHasBackpack = items
						break OuterLoop
					}
				}
			}
			// находим индекс рюкзак в слайсе, из которого хотим удалить
			indexOfBackpack := slices.Index(sliceHasBackpack, "рюкзак")
			// дальше удаляем из этого слайса этот предмет и обновляем слайс
			sliceHasBackpack = slices.Delete(sliceHasBackpack, indexOfBackpack, indexOfBackpack+1)
			// перезаписываем результат обновленных вещей в тот слайс, из которого удаляли элемент в изначальной
			//структуре, перед этим проверив, что там еще что-то осталось
			if len(sliceHasBackpack) == 0 {
				// то есть если кроме рюкзака там ничего и не было, то и сносим нахуй этот пустой слайс
				delete(player.Location.Items, placeWithBackpack)
			} else {
				player.Location.Items[placeWithBackpack] = sliceHasBackpack
			}

			return "вы надели: рюкзак"
		} else {
			return fmt.Sprintf("нельзя надеть %s", itemPutOn)
		}
	}

	return "нет такого предмета"
}

// TODO: ПОСМОТРЕТЬ СЮДА, ДОБАВИТЬ ОБРАБОТКУ С ДВЕРЬЮ!!!

func (player *Player) GoToLocation(newRoom *Room) string {
	checkRoom := slices.Contains(player.Location.PlacesToGo, newRoom)

	if !checkRoom {
		return "нет пути в " + newRoom.Name
	}

	isStreetDoor := (player.Location.Name == "коридор" && newRoom.Name == "улица") ||
		(player.Location.Name == "улица" && newRoom.Name == "коридор")

	if isStreetDoor && !newRoom.Door.IsOpen {
		return "дверь закрыта"
	}

	player.Location = newRoom
	var namesOfRoomsToGo []string

	for _, room := range player.Location.PlacesToGo {
		namesOfRoomsToGo = append(namesOfRoomsToGo, room.Name)
	}

	firstPartOfString := player.Location.InformationOnEnter + ". "
	var secondPartOfString string
	if player.Location.Name == "улица" {
		secondPartOfString = "можно пройти - домой"
	} else {
		secondPartOfString = "можно пройти - " + strings.Join(namesOfRoomsToGo, ", ")
	}
	return firstPartOfString + secondPartOfString
}

func (player *Player) LookAround() string {
	var firstPartOfString string
	var secondPartOfString string

	if len(player.Location.Items) > 0 {
		var totalInfoAboutItems []string
		var places []string

		for place := range player.Location.Items {
			places = append(places, place)
		}

		sort.Strings(places) // надо ли сортировать этот список мест?

		for _, place := range places {
			items := strings.Join(player.Location.Items[place], ", ")

			stringAboutItem := fmt.Sprintf("на %s: %s", place, items)
			totalInfoAboutItems = append(totalInfoAboutItems, stringAboutItem)
		}

		firstPartOfString = player.Location.InformationAround +
			strings.Join(totalInfoAboutItems, ", ")
	} else {
		firstPartOfString = player.Location.EmptyInformation
	}

	if player.Location.AdditionalLookInfo != nil {
		firstPartOfString += player.Location.AdditionalLookInfo(player)
	}

	var namesOfRoomsToGo []string

	for _, room := range player.Location.PlacesToGo {
		namesOfRoomsToGo = append(namesOfRoomsToGo, room.Name)
	}

	if player.Location.Name == "улица" {
		secondPartOfString = ". можно пройти - домой"
	} else {
		secondPartOfString = ". можно пройти - " + strings.Join(namesOfRoomsToGo, ", ")
	}

	return firstPartOfString + secondPartOfString
}

// Создали переменную, в которую будем загружать комнаты в функции initGame()
var rooms []*Room
var player Player

func main() {

	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/

}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все комнаты
		если что-то было - оно корректно перезатирается
	*/

	// кухня, коридор, комната, улица
	rooms = nil

	kitchenRoom := Room{
		Name: "кухня",
		Items: map[string][]string{
			"столе": {"чай"},
		},
		InformationAround:  "ты находишься на кухне, ",
		InformationOnEnter: "кухня, ничего интересного",
		EmptyInformation:   "ты находишься на кухне",
		// PlacesToGo:        {corridorRoom},
	}

	var corridorRoom = Room{
		Name:               "коридор",
		Items:              map[string][]string{}, // пустая мапа так и задумана - в коридоре ничего не лежит
		InformationAround:  "ничего интересного",
		InformationOnEnter: "ничего интересного",
		EmptyInformation:   "ничего интересного",
		// PlacesToGo:        {kitchenRoom, streetRoom, roomRoom},
	}

	var streetRoom = Room{
		Name:               "улица",
		Items:              map[string][]string{},
		InformationAround:  "на улице весна",
		InformationOnEnter: "на улице весна",
		EmptyInformation:   "на улице весна",
		// PlacesToGo:        {corridorRoom},
	}

	var roomRoom = Room{
		Name: "комната",
		Items: map[string][]string{
			"столе": {"ключи", "конспекты"},
			"стуле": {"рюкзак"},
		},
		InformationAround:  "",
		InformationOnEnter: "ты в своей комнате",
		EmptyInformation:   "пустая комната",
		// PlacesToGo:        {corridorRoom},
	}

	player = Player{
		Location:    &kitchenRoom,
		HasBackpack: false,
		Items:       []string{},
	}

	roomRoom.PlacesToGo = []*Room{&corridorRoom}
	kitchenRoom.PlacesToGo = []*Room{&corridorRoom}
	streetRoom.PlacesToGo = []*Room{&corridorRoom}
	corridorRoom.PlacesToGo = []*Room{&kitchenRoom, &roomRoom, &streetRoom}

	kitchenRoom.AdditionalLookInfo = func(player *Player) string {
		if player.HasBackpack {
			return ", надо идти в универ"
		}
		return ", надо собрать рюкзак и идти в универ"
	}

	streetDoor := Door{
		IsOpen: false,
	}
	corridorRoom.Door = &streetDoor
	streetRoom.Door = &streetDoor

	rooms = append(rooms, &roomRoom, &corridorRoom, &kitchenRoom, &streetRoom)
}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	var usedRoom *Room
	partsOfCommand := strings.Fields(command)

	if len(partsOfCommand) == 0 {
		return "неизвестная команда"
	}

	switch partsOfCommand[0] {
	case "идти":

		if len(partsOfCommand) < 2 {
			return "не хватает аргумента"
		}

		for _, room := range rooms {
			if room.Name == partsOfCommand[1] {
				usedRoom = room
				break
			}
		}
		if usedRoom == nil {
			return "нет такой комнаты"
		}
		return player.GoToLocation(usedRoom)

	case "осмотреться":
		return player.LookAround()

	case "надеть":
		if len(partsOfCommand) < 2 {
			return "не хватает аргумента"
		}
		return player.PutOnItem(partsOfCommand[1])

	case "взять":
		if len(partsOfCommand) < 2 {
			return "не хватает аргумента"
		}
		return player.TakeItem(partsOfCommand[1])

	case "применить":
		if len(partsOfCommand) < 3 {
			return "не хватает аргумента"
		}
		return player.ApplyItem(partsOfCommand[1], partsOfCommand[2])

	default:
		return "неизвестная команда"
	}
}
