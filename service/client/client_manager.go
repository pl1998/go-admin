/**
  @author:panliang
  @data:2022/5/27
  @note
**/
package client

import (
	"fmt"
	"im-services/pkg/coroutine_poll"
	"im-services/pkg/logger"
	"sync"
)

type ImClientManager struct {
	ImClientMap      map[int64]*ImClient
	BroadcastChannel chan []byte
	PrivateChannel   chan []byte
	GroupChannel     chan []byte
	Register         chan *ImClient
	Unregister       chan *ImClient
	MutexKey         sync.RWMutex //读写锁
}

var (
	ImManager = ImClientManager{
		ImClientMap:      make(map[int64]*ImClient),
		BroadcastChannel: make(chan []byte),
		PrivateChannel:   make(chan []byte),
		GroupChannel:     make(chan []byte),
		Register:         make(chan *ImClient),
		Unregister:       make(chan *ImClient),
	}
)

type ClientManagerInterface interface {
	// 设置客户端信息
	SetClient(client *ImClient)
	// 删除客户端信息
	DelClient(client *ImClient)
	// 启动服务
	Start()
	// 消息投递到指定客户端
	ImSend(message []byte, client *ImClient)
	// 私聊信息消费
	LaunchPrivateMessage(msg_byte []byte)
	// 群聊信息消费
	LaunchGroupMessage(msg_byte []byte)
	// 消费离线消息
	ConsumingOfflineMessages(client *ImClient)
	// 向好友广播在线状态
	RadioUserOnlineStatus(client *ImClient)
	// 获取在线人数
	GetOnlineNumber() int
}

func (manager *ImClientManager) SetClient(client *ImClient) {
	manager.MutexKey.Lock()
	defer manager.MutexKey.Unlock()
	logger.Logger.Info(fmt.Sprintf("客户端链接:%d", client.ID))
	manager.ImClientMap[client.ID] = client

}

func (manager *ImClientManager) DelClient(client *ImClient) {
	manager.MutexKey.Lock()
	client.Close()
	defer manager.MutexKey.Unlock()
	delete(manager.ImClientMap, client.ID)
}

func (manager *ImClientManager) Start() {
	for {
		select {
		case client := <-ImManager.Register:
			// 设置客户端 拉去离线消息 推送在线状态
			manager.SetClient(client)
			manager.ConsumingOfflineMessages(client)
			//manager.RadioUserOnlineStatus(client)
		case client := <-ImManager.Unregister:
			manager.DelClient(client)
			logger.Logger.Debug(fmt.Sprintf("离线的客户端%s:", client.ID))

		case message := <-ImManager.PrivateChannel:
			coroutine_poll.AntsPool.Submit(func() {
				manager.LaunchPrivateMessage(message)
			})
		case groupMessage := <-ImManager.GroupChannel:
			coroutine_poll.AntsPool.Submit(func() {
				manager.LaunchPrivateMessage(groupMessage)
			})

		}

	}
}

func (manager *ImClientManager) ImSend(message []byte, client *ImClient) {
	data, ok := manager.ImClientMap[client.ID]
	if ok {
		data.Send <- message
	}
}

func (manager *ImClientManager) GetOnlineNumber() int {
	manager.MutexKey.RLock()
	defer manager.MutexKey.RUnlock()
	return len(manager.ImClientMap)
}
