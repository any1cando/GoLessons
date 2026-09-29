package main

import "fmt"

func main() {
	channelOfNotifications := make(chan string, 3)
	channelOfNotifications <- "Extra Message: Alert!"
	channelOfNotifications <- "Nothing serious"
	channelOfNotifications <- "Nothing serious"

	close(channelOfNotifications)

	for message := range channelOfNotifications {
		fmt.Println(message)
	}
}
