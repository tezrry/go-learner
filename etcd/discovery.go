package etcd

import (
	etcd "github.com/kitex-contrib/registry-etcd"
	"log"
)

func main() {

	_, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}) // r should not be reused.
	if err != nil {
		log.Fatal(err)
	}
	//// /docs/tutorials/framework-exten/registry/#integrate-into-kitex
	//server, err := echo.NewServer(new(EchoImpl), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "echo"}), server.WithRegistry(r))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = server.Run()
	//if err != nil {
	//	log.Fatal(err)
	//}

}
