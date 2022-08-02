/**
  @author:panliang
  @data:2022/7/30
  @note
**/
package grpcMessage

import (
	"context"
	"im-services/internal/api/requests"
	"im-services/internal/enum"
	"im-services/internal/service/client"
	messageHandler "im-services/internal/service/message"
	"im-services/pkg/date"
	"im-services/pkg/logger"
)

// ImGrpcMessage 实现 ImMessageServer 接口
type ImGrpcMessage struct {
}

func (ImGrpcMessage) mustEmbedUnimplementedImMessageServer() {}

// ReceivesGrpcPrivateMessage 接收消息
func (ImGrpcMessage) SendMessageHandler(c context.Context, request *SendMessageRequest) (*SendMessageResponse, error) {
	logger.Logger.Error(request.Message)
	params := requests.PrivateMessageRequest{
		MsgId:       date.TimeUnixNano(),
		MsgCode:     enum.WsChantMessage,
		MsgClientId: request.MsgClientId,
		FormID:      request.FormId,
		ToID:        request.ToId,
		ChannelType: int(request.ChannelType),
		MsgType:     int(request.MsgType),
		Message:     request.Message,
		SendTime:    date.NewDate(),
		Data:        request.Data,
	}

	var handler messageHandler.MessageHandler

	msgString := handler.GetGrpcPrivateChatMessages(params)

	switch request.ChannelType {
	case 1:
		client.ImManager.PrivateChannel <- []byte(msgString)
	case 2:
		client.ImManager.GroupChannel <- []byte(msgString)
	}
	return &SendMessageResponse{Code: 200, Message: "Success"}, nil
}
