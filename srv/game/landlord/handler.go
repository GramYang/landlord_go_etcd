package landlord

import (
	"encoding/json"
	"landlord_go/proto"
	agentapi "landlord_go/srv/agent/api"
	"landlord_go/srv/game/database"
)

func Login(req *LoginRequest, cid *proto.ClientID) {
	if req.UserName == "" || req.Password == "" {
		log.Errorln("username or password invalid")
		return
	}
	player := &Player{Cid:cid,UserName:req.UserName}
	clientID2Player.Store(cid.ID, player)
	userName2Player.Store(req.UserName, player)
	chat := &ChatMsgResponse{ChatFlag:1, UserName:req.UserName,
		Msg:req.UserName + "骑着母猪大摇大摆溜进游戏室！", TableNum:-1}
	data, _ := json.Marshal(chat)
	agentapi.BroadcastAll(&proto.JsonACK{JsonType:102, Content:data})
	init := new(InitHallResponse)
	tableMap.Range(func(key, value interface{}) bool {
		players := value.(*Table).Players
		var userNames []string
		for _, v := range players {
			userNames = append(userNames, v.UserName)
		}
		hallList = append(hallList, &HallTable{key.(int32), userNames,
			value.(*Table).IsPlay, value.(*Table).PlayerCount == 3})
		return true
	})
	init.HallTables = hallList
	data, _ = json.Marshal(init)
	agentapi.Send(cid, &proto.JsonACK{JsonType:109, Content:data})
}


func InitHall(req *InitHallRequest, cid *proto.ClientID) {
	init := new(InitHallResponse)
	tableMap.Range(func(key, value interface{}) bool {
		players := value.(*Table).Players
		var userNames []string
		for _, v := range players {
			userNames = append(userNames, v.UserName)
		}
		hallList = append(hallList, &HallTable{key.(int32), userNames,
			value.(*Table).IsPlay, value.(*Table).PlayerCount == 3})
		return true
	})
	init.HallTables = hallList
	data, _ := json.Marshal(init)
	agentapi.Send(cid, &proto.JsonACK{JsonType:109, Content:data})
}

func EnterTable(req *EnterTableRequest, cid *proto.ClientID) {
	value1, _ := tableMap.Load(req.TableNum)
	table := value1.(*Table)
	value2, _ := clientID2Player.Load(cid.ID)
	player := value2.(*Player)
	playerCount := table.PlayerCount
	if playerCount < 3 {
		player.TableNum = req.TableNum
		newSeatNum := GetSeatNum(req.TableNum, playerCount)
		player.SeatNum = newSeatNum
		playerMap.Store(newSeatNum, player)
		table.Players = append(table.Players, player)
		table.PlayerCount = playerCount + 1
		table.Cards = GetRandomCards()
		table.TableNum = req.TableNum
		userName2Player.Store(req.UserName, player)
		tablePlayers := RefreshSeatNum2UserName(table)
		enterTable := &EnterTableResponse{true, tablePlayers}
		data, _ := json.Marshal(enterTable)
		WrapMultiSend(table.Players, &proto.JsonACK{JsonType:105, Content:data}, nil)
		for _, v := range hallList {
			if v.TableNum == table.TableNum {
				v.UserNames = append(v.UserNames, req.UserName)
				v.IsPlay = false
				v.IsFull = len(v.UserNames) == 3
			}
		}
		refreshHall := &RefreshHallResponse{hallList}
		data, _ = json.Marshal(refreshHall)
		var players []*Player
		userName2Player.Range(func(key, value interface{}) bool {
			players = append(players, value.(*Player))
			return true
		})
		WrapMultiSend(players, &proto.JsonACK{JsonType:114, Content:data}, cid)
	} else {
		enterTable := &EnterTableResponse{false, nil}
		data, _ := json.Marshal(enterTable)
		agentapi.Send(cid, &proto.JsonACK{JsonType:105, Content:data})
	}
}

func ChatMsg(req *ChatMsgRequest, cid *proto.ClientID) {
	chatMsg := &ChatMsgResponse{req.ChatFlag, req.UserName, req.Msg,
		req.TableNum}
	data, _ := json.Marshal(chatMsg)
	var players []*Player
	if req.ChatFlag == 1 {
		userName2Player.Range(func(key, value interface{}) bool {
			players = append(players, value.(*Player))
			return true
		})
	}
	if req.ChatFlag == 2 {
		tableMap.Range(func(key, value interface{}) bool {
			players = append(players, value.(*Player))
			return true
		})
	}
	WrapMultiSend(players, &proto.JsonACK{JsonType:102, Content:data}, nil)
}

func Ready(req *ReadyRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	table.ReadyCount++
	if req.IsReady {
		readyResponse := &ReadyResponse{true}
		data, _ := json.Marshal(readyResponse)
		agentapi.Send(cid, &proto.JsonACK{JsonType:113, Content:data})
	}
	if table.ReadyCount == 3 {
		table.IsWait = false
		table.IsGrab = true
		log.Infoln("房间当前准备人数：", table.ReadyCount)
		totalCards := table.Cards
		cardsMap := make(map[int32][]int32)
		cardsMap[0] = totalCards[:17]
		cardsMap[1] = totalCards[17:34]
		cardsMap[2] = totalCards[34:51]
		players := table.Players
		var threeCards []int32
		threeCards = append(threeCards, totalCards[51])
		threeCards = append(threeCards, totalCards[52])
		threeCards = append(threeCards, totalCards[53])
		table.ThreeCards = threeCards
		landlordNum := GetRandomLandlord(table.TableNum)
		for i:=0; i<3; i++ {
			table.PlayersCardsOut[players[i].SeatNum] = 17
			grabLandlord := &GrabLandlordResponse{landlordNum,
				RefreshSeatNum2UserName(table), threeCards, cardsMap[int32(i)]}
			data, _ := json.Marshal(grabLandlord)
			agentapi.Send(players[i].Cid, &proto.JsonACK{JsonType:108, Content:data})
		}
		table.ReadyCount = 0
	}
}

func CancelReady(req *CancelReadyRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	table.ReadyCount--
	if req.IsCancelReady {
		cancelReady := &CancelReadyResponse{true}
		data, _ := json.Marshal(cancelReady)
		agentapi.Send(cid, &proto.JsonACK{JsonType:100,Content:data})
	}
}

func GiveUpLandlord(req *GiveUpLandlordRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	table.PassLandlordCount++
	log.Infoln("累计放弃地主次数：", table.PassLandlordCount)
	giveUpLandlord := &GiveUpLandlordResponse{GetRightRivalSeatNum(req.SeatNum, table.Players)}
	data, _ := json.Marshal(giveUpLandlord)
	WrapMultiSend(table.Players, &proto.JsonACK{JsonType:107,Content:data}, nil)
	if table.PassLandlordCount == 3 {
		log.Infoln("三次扔地主自动结束抢地主")
		table.PassLandlordCount = 0
		landlord := GetRightRivalSeatNum(req.SeatNum, table.Players)
		for k := range table.PlayersCardsOut {
			if k == landlord {
				table.PlayersCardsOut[k] = 20
			}
		}
		endGrabLandlord := &EndGrabLandlordResponse{landlord,table.ThreeCards}
		data, _ = json.Marshal(endGrabLandlord)
		WrapMultiSend(table.Players, &proto.JsonACK{JsonType:104,Content:data}, nil)
	}
}

func EndGrabLandlord(req *EndGrabLandlordRequest, cid *proto.ClientID) {
	landlord := req.MeSeatNum
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	for k := range table.PlayersCardsOut {
		if k == landlord {
			table.PlayersCardsOut[k] = 20
		}
	}
	endGrabLandlord := &EndGrabLandlordResponse{landlord,table.ThreeCards}
	data, _ := json.Marshal(endGrabLandlord)
	WrapMultiSend(table.Players, &proto.JsonACK{JsonType:104,Content:data}, nil)
}

func LandlordMultipleWager(req *LandlordMultipleWagerRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	table.WagerMultipleNum = req.MultipleNum
	if req.MultipleNum == 1 {
		multipleWager := &MultipleWagerResponse{1}
		data, _ := json.Marshal(multipleWager)
		WrapMultiSend(table.Players, &proto.JsonACK{JsonType:112,Content:data}, nil)
	} else {
		landlordMultipleWager := &LandlordMultipleWagerResponse{req.MultipleNum}
		data, _ := json.Marshal(landlordMultipleWager)
		WrapMultiSend(table.Players, &proto.JsonACK{JsonType:110,Content:data}, cid)
	}
}

func MultipleWager(req *MultipleWagerRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	table.AnswerMultipleNum++
	table.AgreedMultipleResult += req.Agreed
	var multipleWager1 *MultipleWagerResponse
	if table.AnswerMultipleNum == 2 {
		if table.AgreedMultipleResult < 2 {
			multipleWager1 = &MultipleWagerResponse{1}
		} else {
			multipleWager1 = &MultipleWagerResponse{table.WagerMultipleNum}
		}
		data, _ := json.Marshal(multipleWager1)
		WrapMultiSend(table.Players, &proto.JsonACK{JsonType:112,Content:data}, nil)
		table.AgreedMultipleResult = 0
		table.WagerMultipleNum = 1
		table.AnswerMultipleNum = 0
	}
}

func CardsOut(req *CardsOutRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	var cardsOut *CardsOutResponse
	if req.IsPass {
		table.ContinuousPass++
		if table.ContinuousPass >= 2 {
			cardsOut = &CardsOutResponse{req.IsPass,true,req.FromSeatNum,
				req.ToSeatNum,table.LastCardsOut,table.ThrowOutCards,
				table.PlayersCardsOut}
			table.ContinuousPass = 0
		} else {
			cardsOut = &CardsOutResponse{req.IsPass, false, req.FromSeatNum,
				req.ToSeatNum,table.LastCardsOut,table.ThrowOutCards,
				table.PlayersCardsOut}
		}
	} else {
		table.ContinuousPass = 0
		table.LastCardsOut = req.CardsOut
		for k := range req.CardsOut {
			table.ThrowOutCards[req.CardsOut[k] % 20]++
		}
		table.PlayersCardsOut[req.FromSeatNum] -= int32(len(req.CardsOut))
		cardsOut = &CardsOutResponse{req.IsPass, false, req.FromSeatNum,
			req.ToSeatNum, req.CardsOut, table.ThrowOutCards,
		table.PlayersCardsOut}
	}
	data, _ := json.Marshal(cardsOut)
	WrapMultiSend(table.Players, &proto.JsonACK{JsonType:101,Content:data}, nil)
}

func EndGame(req *EndGameRequest, cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	endGame1 := &EndGameResponse{req.WinnerSeatNum}
	data, _ := json.Marshal(endGame1)
	WrapMultiSend(table.Players, &proto.JsonACK{JsonType:103,Content:data},nil)
}

func ExitSeat(req *ExitSeatRequest, cid *proto.ClientID) {
	value1, _ := playerMap.Load(req.YourSeatNum)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	for k, v := range table.Players {
		if v == player {
			table.Players = append(table.Players[:k], table.Players[k+1:]...)
		}
	}
	table.PlayerCount--
	table.IsPlay = false
	table.IsGrab = false
	table.IsWait = true
	exitSeat1 := &ExitSeatResponse{player.UserName,req.YourSeatNum,
		RefreshSeatNum2UserName(table)}
	data, _ := json.Marshal(exitSeat1)
	WrapMultiSend(table.Players, &proto.JsonACK{JsonType:106,Content:data}, cid)
	for _, v := range hallList {
		if v.TableNum == table.TableNum {
			v.IsFull = false
			v.IsPlay = false
			for x, y := range v.UserNames {
				if y == player.UserName {
					v.UserNames = append(v.UserNames[:x], v.UserNames[x+1:]...)
				}
			}
		}
	}
	playerMap.Delete(req.YourSeatNum)
	refreshHall1 := &RefreshHallResponse{hallList}
	data, _ = json.Marshal(refreshHall1)
	agentapi.BroadcastAll(&proto.JsonACK{JsonType:114,Content:data})
}

func ExitHall(req *ExitHallRequest, cid *proto.ClientID) {
	userName2Player.Delete(req.UserName)
}

func UserInfo(req *UserInfoRequest, cid *proto.ClientID) {
	res, err := database.GetUserInfo(req.UserName)
	if err != nil {
		return
	}
	userInfo := &UserInfoResponse{res["name"],res["avatar"],res["win"],
		res["lose"],res["money"]}
	data, _ := json.Marshal(userInfo)
	agentapi.Send(cid, &proto.JsonACK{JsonType:115,Content:data})
}

func GameResult(req *GameResultRequest, cid *proto.ClientID) {
	var gameResult1 *GameResultResponse
	if req.Result {
		res := database.Win(req.UserName, req.Password, req.Money)
		gameResult1 = &GameResultResponse{res}
	} else {
		res := database.Lose(req.UserName, req.Password, req.Money)
		gameResult1 = &GameResultResponse{res}
	}
	data, _ := json.Marshal(gameResult1)
	agentapi.Send(cid, &proto.JsonACK{JsonType:116,Content:data})
}

func ExitOrException(cid *proto.ClientID) {
	value1, _ := clientID2Player.Load(cid.ID)
	player := value1.(*Player)
	value2, _ := tableMap.Load(player.TableNum)
	table := value2.(*Table)
	playerMap.Delete(player.SeatNum)
	for k, v := range table.Players {
		if v == player {
			table.Players = append(table.Players[k:], table.Players[:k+1]...)
		}
	}
	table.PlayerCount--
	table.IsPlay = false
	table.IsGrab = false
	table.IsWait = true
	exitSeat1 := &ExitSeatResponse{player.UserName,
		player.SeatNum,RefreshSeatNum2UserName(table)}
	data, _ := json.Marshal(exitSeat1)
	WrapMultiSend(table.Players,&proto.JsonACK{JsonType:106,Content:data}, cid)
	for _, v := range hallList {
		if v.TableNum == table.TableNum {
			v.IsFull = false
			v.IsPlay = false
			for x, y := range v.UserNames {
				if y == player.UserName {
					v.UserNames = append(v.UserNames[x:], v.UserNames[:x+1]...)
				}
			}
		}
	}
	refreshHall1 := &RefreshHallResponse{hallList}
	data, _ = json.Marshal(refreshHall1)
	agentapi.BroadcastAll(&proto.JsonACK{JsonType:114,Content:data})
	userName2Player.Delete(player.UserName)
	clientID2Player.Delete(cid.ID)
}