package main

import (
	"slices"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	waitGroup := &sync.WaitGroup{}
	in := make(chan interface{}, 1)

	for _, cmd := range cmds { // cmds - это слайс команд, поэтому берем и индекс команды, и саму команду.
		out := make(chan interface{}, 1)

		// Каждый раз выходной канал мы обновляем. На первой итерации там будет пусто,
		// потому что мы еще ничего не выводили, а на остальных итерациях мы будем сохранять результат во входной канал
		// после всей проделанной работы, а потом обновлять выходящий

		waitGroup.Add(1)
		go func(in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			cmd(in, out)
			close(out)
		}(in, out, waitGroup)
		in = out
	}
	waitGroup.Wait()
	// TODO: А будет ли какое-то возвращаемое значение у функции RunPipeline??
}

// TODO: Нужно отрисовать флоу того, как это все протекает на схеме визуально, потому что я
// TODO: не понимаю, кто / кого / за кем вызывает и в какой последовательности

func SelectUsers(in, out chan interface{}) {
	wgForSelectUsers := &sync.WaitGroup{}
	mutexForSelectUsers := &sync.Mutex{}
	// 	in - string
	var uniqueUsers []User

	for emailUser := range in {
		wgForSelectUsers.Add(1)
		go func() {
			defer wgForSelectUsers.Done()
			user := GetUser(emailUser.(string))
			mutexForSelectUsers.Lock()
			if !slices.Contains(uniqueUsers, user) {
				uniqueUsers = append(uniqueUsers, user)
				mutexForSelectUsers.Unlock()
				out <- user
			} else {
				mutexForSelectUsers.Unlock()
			}
		}()
	}
	wgForSelectUsers.Wait()
	// 	out - User
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User

	wgForSelectMessages := &sync.WaitGroup{}
	channelForMaxUsers := make(chan User, 2)

	for {
		user, ok := <-in
		callFunction := func(users ...User) {
			defer wgForSelectMessages.Done()
			messagesWeGot, err := GetMessages(users...)
			if err != nil {
				// в случае ошибки - выводим ошибку
				return
			}
			for _, message := range messagesWeGot {
				out <- message
			}
		}
		if ok {
			channelForMaxUsers <- user.(User)
			if len(channelForMaxUsers) == 2 {
				wgForSelectMessages.Add(1)
				go callFunction(<-channelForMaxUsers, <-channelForMaxUsers)
			}
		} else if len(channelForMaxUsers) == 1 {
			wgForSelectMessages.Add(1)
			go callFunction(<-channelForMaxUsers)
		} else {
			break
		}
	}

	wgForSelectMessages.Wait()

	// 	out - MsgID
}

//func CheckSpam(in, out chan interface{}) {
//	// in - MsgID
//	// out - MsgData
//}
//
//func CombineResults(in, out chan interface{}) {
//	// in - MsgData
//	// out - string
//}
