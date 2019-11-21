package hubmodel

import "github.com/davyxu/cellnet"

var (
	chanByName = map[string][]cellnet.Session{}
)

func AddSubscriber(name string, ses cellnet.Session) {
	list, _ := chanByName[name]
	list = append(list, ses)
	chanByName[name] = list
}

func RemoveSubscriber(ses cellnet.Session, callback func(chanName string)) {
	var found bool
	for {
		found = false
	Refound:
		for name, list := range chanByName {
			for index, libSes := range list {
				if libSes == ses {
					callback(name)
					list = append(list[:index], list[index + 1:]...)
					if len(list) == 0 {
						delete(chanByName, name)
					} else {
						chanByName[name] = list
					}
					found = true
					goto Refound //再查一次，下次没有libSes == ses就不走这里了
 				}
			}
		}
		if !found {
			break
		}
	}
}

func VisitSubscriber(name string, callback func(ses cellnet.Session) bool) (count int) {
	if list, ok := chanByName[name]; ok {
		for _, ses := range list {
			count++
			if !callback(ses) {
				return
			}
		}
	}
	return
}