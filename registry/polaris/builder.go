package grpcpolaris

import (
	"google.golang.org/grpc/resolver"
)

const scheme string = "grpcpolaris"

type grpcpolarisBuilder struct {
	servers   []string
	namespace string
	service   string
}

func (b *grpcpolarisBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, err error) {
	var watcher *GrpcPolaris
	if watcher, err = newGrpcPolaris(b.servers, b.namespace, b.service); err != nil {
		return
	}
	var r *grpcpolarisResolver = &grpcpolarisResolver{
		target:  target,
		cc:      cc,
		watcher: watcher,
	}
	r.start()
	return r, nil
}

func (_ *grpcpolarisBuilder) Scheme() string {
	return scheme
}

func RegisterResolver(servers []string, namespace string, service string) {
	resolver.Register(&grpcpolarisBuilder{
		servers:   servers,
		namespace: namespace,
		service:   service,
	})
}
