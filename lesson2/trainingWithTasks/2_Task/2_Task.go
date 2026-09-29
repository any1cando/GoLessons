// 2. Очередь уведомлений — буферизированный канал
//Нужно сначала накопить несколько уведомлений, а затем вывести их.
//- Создай строковый канал с буфером на три сообщения.
//- В main отправь в него три уведомления до запуска получателя.
//- Закрой канал.
//- Прочитай все сообщения через range и выведи.
//Затем увеличь число сообщений до четырёх, оставив буфер прежним. Объясни, почему программа перестала работать.

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
