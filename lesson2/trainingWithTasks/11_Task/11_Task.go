//11. Запрос с ограничением времени — context.WithTimeout
//Сервис должен прекратить работу, если запрос выполняется слишком долго.
//- Напиши функцию loadData, принимающую контекст и длительность имитации загрузки.
//- Функция возвращает (string, error).
//- Если загрузка закончилась — верни "данные загружены" и nil.
//- Если контекст завершился раньше — верни пустую строку и ctx.Err().
//- Ожидание внутри функции должно прерываться при отмене контекста.
//- Вызови её с тайм-аутом 200 ms для загрузок длительностью 50 ms и 500 ms.
//Дополнение: повтори с context.WithDeadline, задав конкретный момент завершения.

package main

import (
	context2 "context"
	"fmt"
	"time"
)

func loadData(ctx context2.Context, durationTime time.Duration) (string, error) {

	select {
	case <-time.After(durationTime):
		return "Данные загружены", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	myContextWithTimeout, cancel := context2.WithTimeout(context2.Background(), time.Millisecond*200)
	defer cancel() // вызываем для очистки ресурсов
	duration := time.Millisecond * 500
	answer, err := loadData(myContextWithTimeout, duration)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(answer)
}
