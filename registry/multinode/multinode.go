package multinode

import (
	"strings"

	"google.golang.org/grpc/resolver"
)

const scheme string = "multinode"

type multinodeBuilder struct{}

func (_ *multinodeBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, err error) {
	var r *multinodeResolver = &multinodeResolver{
		target: target,
		cc:     cc,
	}
	r.start()
	return r, nil
}

func (_ *multinodeBuilder) Scheme() string {
	return scheme
}

type multinodeResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *multinodeResolver) start() {
	var (
		addrlist          []string = strings.Split(r.target.Endpoint, ",")
		addr              string
		resolverAddresses []resolver.Address
	)
	for _, addr = range addrlist {
		resolverAddresses = append(resolverAddresses, resolver.Address{Addr: addr})
	}
	r.cc.UpdateState(resolver.State{Addresses: resolverAddresses})
}

func (r *multinodeResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *multinodeResolver) Close() {}

func init() {
	resolver.Register(&multinodeBuilder{})
}
