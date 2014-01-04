package main

type SafeMap interface {
	Insert(string, interface{})
	Delete(string)
	Find(string) (interface{}, bool)
	Update(string, UpdateFunc)
	Close() map[string]interface{}
}

type UpdateFunc func(interface{}, bool) interface{}

type safeMap chan commandData

type commandData struct {
	action  commandAction
	key     string
	value   interface{}
	result  chan<- interface{}
	data    chan<- map[string]interface{}
	updater UpdateFunc
}

type commandAction int

const (
	remove commandAction = iota
	end
	find
	insert
	length
	update
)

func (sm SafeMap) Insert(key string, value interface{}) {
	sm <- commandData{action: insert, key: key, value: value}
}

func (sm SafeMap) Delete(key string) {
	sm <- commandData{action: delete, key: key}
}

type findResult struct {
	value interface{}
	found bool
}

func (sm SafeMap) Find(key string) (value interface{}, found bool) {
	reply := make(chan interface{})
	sm <- commandData{action: find, key: key, result: reply}
	result := (<-reply).(findResult)
	return result.value, result.found
}

func (sm SafeMap) Len() int {
	reply := make(chan interface{})
	sm <- commandData{action: length, result: reply}
	return (<-reply).(int)
}

func (sm SafeMap) Update(key string, updater UpdateFunc) {
	sm <- commandData{action: find, key: key, updater: updater}
}

func (sm SafeMap) Close() map[string]interface{} {
	reply := make(chan map[string]interface{})
	sm <- commandData{action: end, data: reply}
	return <-reply
}

func New() SafeMap {
	sm := make(safeMap)
	go sm.run()
	return sm
}

func (sm safeMap) run() {
	store := make(map[string]interface{})
	for command := range sm {
		switch command.action {
		case insert:
			store[command.key] = command.value
		case remove:
			delete(store, command.key)
		case find:
			value, found := store[command.key]
			command.result <- findResult{value, found}
		case length:
			command.result <- len(store)
		case update:
			value, found := store[command.key]
			store[command.key] = command.updater(value, found)
		case end:
			close(sm)
			command.data <- store
		}
	}
}
