package radius
import (
	"net"
	"log"
)
const AUTH_PORT = 1812
const ACCOUNTING_PORT = 1813

type Server struct {
	addr string
	secret string
	services map[string]Service
}

type Service interface {
    Authenticate(request *Packet) (*Packet,error)
}

type PasswordService struct {}

func(p *PasswordService)Authenticate(request *Packet) (*Packet,error){
	npac := request.Reply()
	npac.Code = AccessReject
	npac.AVPs = append(npac.AVPs, AVP{Type:ReplyMessage, Value:[]byte("you dick!")})
	return npac,nil
}
func NewServer(addr string, secret string) *Server {
	return &Server{addr,secret,make(map[string]Service)}
}

func (s *Server) RegisterService(serviceAddr string,  handler Service){
	s.services[serviceAddr] = handler	
}

func (s *Server) ListenAndServe() error {
	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	var b [512]byte
	for {
		n, addr, err := conn.ReadFrom(b[:])
		if err != nil {
			return err
		}
	
		p := b[:n]
		pac := &Packet{server:s}
		err = pac.Decode(p)
		if err != nil {
			return err
		}
		
		ips:= pac.Attributes(NASIPAddress)
		
		if len(ips) != 1 {
			continue
		}
	
		ss := net.IP(ips[0].Value[0:4])
		
		service, ok := s.services[ss.String()]
		if !ok {
			log.Println("recieved request for unknown service: ",ss)
			continue
		
			//reject
		}
		npac,err := service.Authenticate(pac)
		if err != nil {
			return err
		}
		err = npac.Send(conn, addr)
		if err != nil {
			return err
		}
	}
	return nil
}	