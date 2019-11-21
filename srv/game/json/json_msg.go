package json

import (
	"github.com/davyxu/cellnet"
	"landlord_go/proto"
	agentapi "landlord_go/srv/agent/api"
	"landlord_go/srv/game/landlord"
	"reflect"
)

func init() {
	proto.Handle_Game_JsonREQ = agentapi.HandleBackendMessage(func(ev cellnet.Event, cid proto.ClientID) {
		switch msg := ev.Message().(type) {
		case *proto.JsonREQ:
			log.Infof("receive json message: %v content: %+v\n", landlord.Code2bean[msg.JsonType], string(msg.Content))
			req := reflect.New(landlord.Code2bean[msg.JsonType]).Interface()
			switch msg.JsonType {
			case 21:
				landlord.Login(req.(*landlord.LoginRequest), &cid)
			case 19:
				landlord.InitHall(req.(*landlord.InitHallRequest), &cid)
			case 15:
				landlord.EnterTable(req.(*landlord.EnterTableRequest), &cid)
			case 12:
				landlord.ChatMsg(req.(*landlord.ChatMsgRequest), &cid)
			case 23:
				landlord.Ready(req.(*landlord.ReadyRequest), &cid)
			case 10:
				landlord.CancelReady(req.(*landlord.CancelReadyRequest), &cid)
			case 18:
				landlord.GiveUpLandlord(req.(*landlord.GiveUpLandlordRequest), &cid)
			case 14:
				landlord.EndGrabLandlord(req.(*landlord.EndGrabLandlordRequest), &cid)
			case 20:
				landlord.LandlordMultipleWager(req.(*landlord.LandlordMultipleWagerRequest), &cid)
			case 22:
				landlord.MultipleWager(req.(*landlord.MultipleWagerRequest), &cid)
			case 11:
				landlord.CardsOut(req.(*landlord.CardsOutRequest), &cid)
			case 13:
				landlord.EndGame(req.(*landlord.EndGameRequest), &cid)
			case 17:
				landlord.ExitSeat(req.(*landlord.ExitSeatRequest), &cid)
			case 16:
				landlord.ExitHall(req.(*landlord.ExitHallRequest), &cid)
			case 24:
				landlord.UserInfo(req.(*landlord.UserInfoRequest), &cid)
			case 25:
				landlord.GameResult(req.(*landlord.GameResultRequest), &cid)
			}
		case *proto.ClientClosedACK:
			log.Infof("client with id: %s has exit", msg.ID)
			landlord.ExitOrException(&cid)
		}
	})
}