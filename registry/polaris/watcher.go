package grpcpolaris

import (
	"fmt"
	"sync"
	"time"

	"github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/pkg/model"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

type GrpcPolaris struct {
	servers               []string
	namespace             string
	service               string
	consumer              polaris.ConsumerAPI
	resolverAddresses     []resolver.Address
	resolverAddressesChan chan []resolver.Address
	wg                    sync.WaitGroup
}

func newGrpcPolaris(servers []string, namespace string, service string) (p *GrpcPolaris, err error) {
	var consumer polaris.ConsumerAPI
	if consumer, err = polaris.NewConsumerAPIByAddress(servers...); err != nil {
		return
	}
	p = &GrpcPolaris{
		servers:               servers,
		namespace:             namespace,
		service:               service,
		consumer:              consumer,
		resolverAddressesChan: make(chan []resolver.Address, 10),
	}
	return
}

func (p *GrpcPolaris) Close() {
	p.consumer.Destroy()
	p.wg.Wait()
}

func (p *GrpcPolaris) Watch() chan []resolver.Address {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			var (
				resolverAddresses []resolver.Address
				eventChannel      <-chan model.SubScribeEvent
				err               error
			)
			if resolverAddresses, eventChannel, err = p.WatchService(); err != nil {
				grpclog.Errorf("GrpcPolaris WatchService: %v", err.Error())
				time.Sleep(time.Second)
				continue
			}
			if len(resolverAddresses) > 0 && !isSameAddrs(p.resolverAddresses, resolverAddresses) {
				p.resolverAddresses = resolverAddresses
				p.resolverAddressesChan <- cloneAddresses(resolverAddresses)
			}
			select {
			case <-eventChannel:
			case <-time.After(time.Minute): // 定期触发一次
			}
		}
	}()
	return p.resolverAddressesChan
}

// WatchService 订阅服务消息
func (p *GrpcPolaris) WatchService() (resolverAddresses []resolver.Address, eventChannel <-chan model.SubScribeEvent, err error) {
	var (
		request  *polaris.WatchServiceRequest = new(polaris.WatchServiceRequest)
		response *model.WatchServiceResponse
	)
	request.Key.Namespace = p.namespace
	request.Key.Service = p.service
	if response, err = p.consumer.WatchService(request); err != nil {
		return
	}
	var (
		list []model.Instance = response.GetAllInstancesResp.GetInstances()
		item model.Instance
	)
	for _, item = range list {
		resolverAddresses = append(resolverAddresses, resolver.Address{
			Addr: fmt.Sprintf("%s:%d", item.GetHost(), item.GetPort()),
		})
	}
	eventChannel = response.EventChannel
	return
}

func isSameAddrs(addrList1, addrList2 []resolver.Address) bool {
	if len(addrList1) != len(addrList2) {
		return false
	}
	var (
		found bool
		addr1 resolver.Address
		addr2 resolver.Address
	)
	for _, addr1 = range addrList1 {
		found = false
		for _, addr2 = range addrList2 {
			if addr1.Addr == addr2.Addr {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func cloneAddresses(in []resolver.Address) (out []resolver.Address) {
	out = make([]resolver.Address, len(in))
	copy(out, in)
	return out
}
