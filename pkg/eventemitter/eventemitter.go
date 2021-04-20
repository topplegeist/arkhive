package eventemitter

import (
	"errors"
	"reflect"
)

type EventEmitter struct {
	messagesType reflect.Type
	subscribers  []*Subscriber
}

func (eventEmitter *EventEmitter) Emit(message interface{}) error {
	if eventEmitter.messagesType == nil && len(eventEmitter.subscribers) > 0 {
		return errors.New("no event emitter message type")
	} else if len(eventEmitter.subscribers) == 0 {
		return nil
	}
	if reflect.TypeOf(message) != eventEmitter.messagesType {
		panic("Emitting wrong data type, expected '" + eventEmitter.messagesType.Name() + "'")
	}
	for _, subscriber := range eventEmitter.subscribers {
		subscriber.enqueue(message)
	}
	return nil
}

func (eventEmitter *EventEmitter) Subscribe(callback interface{}) {
	callbackType := reflect.TypeOf(callback)
	if callbackType.Kind() != reflect.Func {
		panic("Callback is not a function")
	}
	if callbackType.NumIn() != 1 {
		panic("Callback function should accept one event as argument")
	}
	callbackMessageType := callbackType.In(0)
	if eventEmitter.messagesType == nil {
		eventEmitter.messagesType = callbackMessageType
	} else if eventEmitter.messagesType != callbackMessageType {
		panic("Event emitter and callback data type are different, expected '" + eventEmitter.messagesType.Name() + "'")
	}
	subscriber := newSubscriber(eventEmitter.messagesType)

	(*subscriber).callback = callback
	eventEmitter.subscribers = append(eventEmitter.subscribers, subscriber)
}

type Subscriber struct {
	inputQueue reflect.Value
	callback   interface{}
}

func newSubscriber(messagesType reflect.Type) *Subscriber {
	instance := new(Subscriber)
	channelType := reflect.ChanOf(reflect.BothDir, messagesType)
	instance.inputQueue = reflect.MakeChan(channelType, 1)
	go func() {
		for {
			if message, ok := instance.inputQueue.Recv(); ok {
				reflect.ValueOf(instance.callback).Call([]reflect.Value{message})
			}
		}
	}()
	return instance
}

func (subscriber *Subscriber) enqueue(message interface{}) {
	subscriber.inputQueue.Send(reflect.ValueOf(message))
}
