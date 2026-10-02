package main

import (
	"fmt"
	"maps"
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
		callFunctionSelectMessages := func(users ...User) {
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
			if len(channelForMaxUsers) == GetMessagesMaxUsersBatch {
				wgForSelectMessages.Add(1)
				go callFunctionSelectMessages(<-channelForMaxUsers, <-channelForMaxUsers)
			}
		} else if len(channelForMaxUsers) == 1 {
			wgForSelectMessages.Add(1)
			go callFunctionSelectMessages(<-channelForMaxUsers)
		} else {
			break
		}
	}

	wgForSelectMessages.Wait()

	// 	out - MsgID
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	mutexForSpam := &sync.Mutex{}
	conditionalMutex := sync.NewCond(mutexForSpam)
	counterForRunningCoroutines := 0
	waitGroup := &sync.WaitGroup{}

	// TODO: Как подсчитывать количество запущенных корутин именно в этой функции? Мьютекс?
	for messageId := range in {

		mutexForSpam.Lock()
		for counterForRunningCoroutines == HasSpamMaxAsyncRequests {
			conditionalMutex.Wait()
		}
		counterForRunningCoroutines++
		mutexForSpam.Unlock()
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			messageCheckOnSpam, err := HasSpam(messageId.(MsgID))
			if err != nil {
				// ошибка, что-то с ней делаем
				mutexForSpam.Lock()
				counterForRunningCoroutines--
				conditionalMutex.Signal()
				mutexForSpam.Unlock()
				return
			}

			mutexForSpam.Lock()
			counterForRunningCoroutines--
			conditionalMutex.Signal()
			mutexForSpam.Unlock()

			out <- MsgData{messageId.(MsgID), messageCheckOnSpam}
		}()
	}

	waitGroup.Wait()

	// out - MsgData
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	allMessagesDataTrue := map[MsgID]bool{}
	allMessagesDataFalse := map[MsgID]bool{}

	for messageData := range in {

		if messageData.(MsgData).HasSpam {
			allMessagesDataTrue[messageData.(MsgData).ID] = messageData.(MsgData).HasSpam
		} else {
			allMessagesDataFalse[messageData.(MsgData).ID] = messageData.(MsgData).HasSpam
		}
	}

	keysForTrue := slices.Collect(maps.Keys(allMessagesDataTrue))
	slices.Sort(keysForTrue)

	for _, messageId := range keysForTrue {
		finalString := fmt.Sprintf("true %v", messageId)
		out <- finalString
	}

	keysForFalse := slices.Collect(maps.Keys(allMessagesDataFalse))
	slices.Sort(keysForFalse)

	for _, messageId := range keysForFalse {
		finalString := fmt.Sprintf("false %v", messageId)
		out <- finalString
	}

	// out - string
}
