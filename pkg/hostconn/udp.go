/* hostconn/udp.go
 * Report current judge state to some host, in udp
 */

package hostconn

import (
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/erjiaqing/PCIJudger2/pkg/hostconn/message"
	"github.com/erjiaqing/PCIJudger2/pkg/util"
	"github.com/sirupsen/logrus"
)

type UDP struct {
	socket *net.UDPConn
	uid    string
}

func NewUDP(ip string, port int, uid string) *UDP {
	if ip == "" || port == 0 {
		return &UDP{}
	}
	if uid == "" {
		uid = util.RandSeq(12)
	}
	udpaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		logrus.Fatalf("Failed to resolve address of udp host: %v", err)
	}
	addr, err := net.DialUDP("udp", nil, udpaddr)
	if err != nil {
		logrus.Fatalf("Failed to create connection to udp host: %v", err)
	}
	return &UDP{
		socket: addr,
		uid:    uid,
	}
}

func (c *UDP) SendStatus(state string, progress int) {
	if c == nil || c.socket == nil {
		return
	}
	msg := &message.StateMessage{}
	msg.Uid = c.uid
	msg.State = state
	msg.Progress = int32(progress)

	data, err := proto.Marshal(msg)
	if err != nil {
		logrus.Errorf("Failed to marshal message: %v", data)
	}
	c.socket.Write([]byte(data))
}
