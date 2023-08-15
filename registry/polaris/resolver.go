package grpcpolaris

import (
	"sync"

	"google.golang.org/grpc/resolver"
)

type grpcpolarisResolver struct {
	target  resolver.Target
	cc      resolver.ClientConn
	watcher *GrpcPolaris
	wg      sync.WaitGroup
}

func (r *grpcpolarisResolver) start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		var resolverAddresses []resolver.Address
		for resolverAddresses = range r.watcher.Watch() {
			r.cc.UpdateState(resolver.State{Addresses: resolverAddresses})
		}
	}()
}

func (r *grpcpolarisResolver) ResolveNow(o resolver.ResolveNowOptions) {
}

func (r *grpcpolarisResolver) Close() {
	r.watcher.Close()
	r.wg.Wait()
}
