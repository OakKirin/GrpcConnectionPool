package GrpcConnectionPool

import (
	"math/rand"
	"sync"
)

var poolOnce sync.Once

//连接池管理
type ConnectionPool struct {
	//key为服务名
	pool map[string]map[string]chan *grpcConnection
	//健康服务列表及其ip:port
	serviceMap map[string][]string
}

func (this *ConnectionPool) newConnectionPool() {
	this.pool = make(map[string]map[string]chan *grpcConnection)
	this.serviceMap = make(map[string][]string)
}

//初始化连接池和健康服务列表监控
func InitConnectionPool() *ConnectionPool {
	pool := &ConnectionPool{}
	poolOnce.Do(pool.newConnectionPool)
	return pool
}

func (this *ConnectionPool) createNewConnChan(address, service string) chan *grpcConnection {
	c := make(chan *grpcConnection, ActiveConn)
	for i := 1; i <= ActiveConn; i++ {
		c <- NewGrpcConnection(address, service)
	}
	return c
}

func (this *ConnectionPool) PopConnection(service string) *grpcConnection {
	r := rand.Int31n(int32(len(this.serviceMap[service])))
	c := <-this.pool[service][this.serviceMap[service][r]]
	return c
}

func (this *ConnectionPool) PutConnection(conn *grpcConnection) {
	this.pool[conn.service][conn.address] <- conn
}

func (this *ConnectionPool) notice(service string, serviceAddressList []string) {
	if value, ok := this.pool[service]; ok {
		for i := 0; i < len(serviceAddressList); i++ {
			if _, ok := value[serviceAddressList[i]]; !ok {
				value[serviceAddressList[i]] = this.createNewConnChan(serviceAddressList[i], service)
			}
		}
	} else {
		this.pool[service] = make(map[string]chan *grpcConnection)
		for i := 0; i < len(serviceAddressList); i++ {
			this.pool[service][serviceAddressList[i]] = this.createNewConnChan(serviceAddressList[i], service)
		}
	}
	this.serviceMap[service] = serviceAddressList
}