package main

import "sync"

func RunPipeline(cmds ...cmd) {
	waitGroup := &sync.WaitGroup{}
	// создать два канала с буфером
	in := make(chan interface{}, 5)

	for _, cmd := range cmds { // cmds - это слайс команд, поэтому берем и индекс команды, и саму команду.
		out := make(chan interface{}, 5)

		// Каждый раз выходной канал мы обновляем. На первой итерации там будет пусто,
		// потому что мы еще ничего не выводили, а на остальных итерациях мы будем сохранять результат в входной канал
		//после всей проделанной работы

		waitGroup.Add(1)
		go func(in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			cmd(in, out)
		}(in, out, waitGroup)
		in = out
	}
	waitGroup.Wait()
}

// TODO: Нужно отрисовать флоу того, как это все протекает на схеме визуально, потому что я
// TODO: не понимаю, кто / кого / за кем вызывает и в какой последовательности

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	var user User
	for {
		emailUser, ok := <-in
		if ok {
			user = GetUser(emailUser.(string))
			out <- user
		} else {
			close(out)
			break
		}
	}
	// 	out - User
}

//func SelectMessages(in, out chan interface{}) {
//	// 	in - User
//	// 	out - MsgID
//}
//
//func CheckSpam(in, out chan interface{}) {
//	// in - MsgID
//	// out - MsgData
//}
//
//func CombineResults(in, out chan interface{}) {
//	// in - MsgData
//	// out - string
//}
