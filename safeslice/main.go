package safeSlice

type SafeSlice interface {
	Append(interface{})
	At(int) interface{}
	Close() []interface{}
	Delete(int)
	Len() int
	Update(int, UpdateFunc)
}

type UpdateFunc func(interface{}) interface{}

type safeSlice chan commandData

func (sl safeSlice) run() {
	list := []interface{}{}
	for command := range sl {
		switch command.action {
		case appendSlice:
			list = append(list, command.value)
		case atSlice:
			if 0 <= command.idx && command.idx < len(list) {
				command.result <- list[command.idx]
			} else {
				command.result <- nil
			}
		case deleteSlice:
			if command.idx < len(list) {
				list = append(list[:command.idx], list[command.idx+1:]...)
			}
		case lenSlice:
			command.result <- len(list)
		case updateSlice:
			if 0 <= command.idx && command.idx < len(list) {
				command.updater(list[command.idx])
			}
		case closeSlice:
			close(sl)
			command.data <- list
		}
	}
}

type commandData struct {
	action  commandAction
	idx     int
	value   interface{}
	result  chan<- interface{}
	data    chan<- []interface{}
	updater UpdateFunc
}

type commandAction int

const (
	appendSlice commandAction = iota
	atSlice
	closeSlice
	deleteSlice
	lenSlice
	updateSlice
)

func (sl safeSlice) Append(value interface{}) {
	sl <- commandData{action: appendSlice, value: value}
}

func (sl safeSlice) At(index int) (value interface{}) {
	reply := make(chan interface{})
	sl <- commandData{action: atSlice, idx: index, result: reply}
	return <-reply
}

func (sl safeSlice) Close() (slice []interface{}) {
	data := make(chan []interface{})
	sl <- commandData{action: closeSlice, data: data}
	return <-data
}

func (sl safeSlice) Delete(index int) {
	sl <- commandData{action: deleteSlice, idx: index}
}

func (sl safeSlice) Len() int {
	reply := make(chan interface{})
	sl <- commandData{action: lenSlice, result: reply}
	return (<-reply).(int)
}

func (sl safeSlice) Update(index int, updater UpdateFunc) {
	sl <- commandData{action: updateSlice, updater: updater}
}

func New() SafeSlice {
	sl := make(safeSlice)
	go sl.run()
	return sl
}
