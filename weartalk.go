package main

import "strconv"

type WearTalk struct {
	UID      int
	NickName string
	Sex      int8
	Key      string
}

func (wt *WearTalk) getNickName() string {
	switch wt.Sex {
	case 0:
		return wt.NickName + "♂"

	case 1:
		return wt.NickName + "♀"

	default:
		return wt.NickName
	}
}

func (wt *WearTalk) getUID() string {
	return "TaMP" + strconv.Itoa(wt.UID)
}
