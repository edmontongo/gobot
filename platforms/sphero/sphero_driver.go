package sphero

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
        "log"
	"github.com/edmontongo/gobot"
)

type packet struct {
	header   []uint8
	body     []uint8
	checksum uint8
}

type SpheroDriver struct {
	gobot.Driver
	seq             uint8
	asyncResponse   [][]uint8
	syncResponse    [][]uint8
	packetChannel   chan *packet
	responseChannel chan []uint8
}

type Collision struct {
	// Normalized impact components (direction of the collision event):
	X, Y, Z int16
	// Thresholds exceeded by X (1h) and/or Y (2h) axis (bitmask):
	Axis byte
	// Power that cross threshold Xt + Xs:
	XMagnitude, YMagnitude int16
	// Sphero's speed when impact detected:
	Speed uint8
	// Millisecond timer
	Timestamp uint32
}

// OK this is sorta bad because it is hardcoded
// response of all 3 masks
type Locator struct {
	// did 02
	// dlen x0b
 	// XPOS, YPOS, XVEL, YVEL, SOG
        Xpos, Ypos, Accel, Xvel, Yvel int16
}

const LocatorMask2  uint32 = 0x0C000000
const AccelOneMask2 uint32 = 0x02000000
const VelocityMask2 uint32 = 0x01800000

func NewSpheroDriver(a *SpheroAdaptor, name string) *SpheroDriver {
	s := &SpheroDriver{
		Driver: *gobot.NewDriver(
			name,
			"SpheroDriver",
			a,
		),
		packetChannel:   make(chan *packet, 1024),
		responseChannel: make(chan []uint8, 1024),
	}

	s.AddEvent("collision")
	s.AddEvent("locator")
	s.AddCommand("SetRGB", func(params map[string]interface{}) interface{} {
		r := uint8(params["r"].(float64))
		g := uint8(params["g"].(float64))
		b := uint8(params["b"].(float64))
		s.SetRGB(r, g, b)
		return nil
	})

	s.AddCommand("Roll", func(params map[string]interface{}) interface{} {
		speed := uint8(params["speed"].(float64))
		heading := uint16(params["heading"].(float64))
		s.Roll(speed, heading)
		return nil
	})

	s.AddCommand("Stop", func(params map[string]interface{}) interface{} {
		s.Stop()
		return nil
	})

	s.AddCommand("GetRGB", func(params map[string]interface{}) interface{} {
		return s.GetRGB()
	})

	s.AddCommand("SetBackLED", func(params map[string]interface{}) interface{} {
		level := uint8(params["level"].(float64))
		s.SetBackLED(level)
		return nil
	})

	s.AddCommand("SetHeading", func(params map[string]interface{}) interface{} {
		heading := uint16(params["heading"].(float64))
		s.SetHeading(heading)
		return nil
	})
	s.AddCommand("SetStabilization", func(params map[string]interface{}) interface{} {
		on := params["heading"].(bool)
		s.SetStabilization(on)
		return nil
	})

	return s
}

func (s *SpheroDriver) adaptor() *SpheroAdaptor {
	return s.Adaptor().(*SpheroAdaptor)
}

func (s *SpheroDriver) Init() bool {
	return true
}

func (s *SpheroDriver) Start() bool {
	go func() {
		for {
			packet := <-s.packetChannel
			s.write(packet)
		}
	}()

	go func() {
		for {
			response := <-s.responseChannel
			s.syncResponse = append(s.syncResponse, response)
		}
	}()

	go func() {
		for {
			header := s.readHeader()
			// log.Printf("header: %x\n", header)
			if header != nil && len(header) != 0 {
				body := s.readBody(header[4])
				// log.Printf("body: %x\n", body)
				if header[1] == 0xFE {
					async := append(header, body...)
					s.asyncResponse = append(s.asyncResponse, async)
				} else {
					s.responseChannel <- append(header, body...)
				}
			}
		}
	}()

	go func() {
		for {
			var evt []uint8
			for len(s.asyncResponse) != 0 {
				evt, s.asyncResponse = s.asyncResponse[len(s.asyncResponse)-1], s.asyncResponse[:len(s.asyncResponse)-1]
				if evt[2] == 0x07 {
					s.handleCollisionDetected(evt)
				} else if evt[2] == 0x03 {
					s.handleLocator(evt)
				} else {
					log.Printf("Async: %v\n", evt[2])
				}


			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	s.configureDefaultCollisionDetection()
	s.enableStopOnDisconnect()

	return true
}

func (s *SpheroDriver) Halt() bool {
	gobot.Every(10*time.Millisecond, func() {
		s.Stop()
	})
	time.Sleep(1 * time.Second)
	return true
}

func (s *SpheroDriver) SetRGB(r uint8, g uint8, b uint8) {
	s.packetChannel <- s.craftPacket([]uint8{r, g, b, 0x01}, 0x02, 0x20)
}

func (s *SpheroDriver) GetRGB() []uint8 {
	buf := s.getSyncResponse(s.craftPacket([]uint8{}, 0x02, 0x22))
	if len(buf) == 9 {
		return []uint8{buf[5], buf[6], buf[7]}
	}
	return []uint8{}
}

func (s *SpheroDriver) SetBackLED(level uint8) {
	s.packetChannel <- s.craftPacket([]uint8{level}, 0x02, 0x21)
}

func (s *SpheroDriver) SetHeading(heading uint16) {
	s.packetChannel <- s.craftPacket([]uint8{uint8(heading >> 8), uint8(heading & 0xFF)}, 0x02, 0x01)
}

func (s *SpheroDriver) SetStabilization(on bool) {
	b := uint8(0x01)
	if on == false {
		b = 0x00
	}
	s.packetChannel <- s.craftPacket([]uint8{b}, 0x02, 0x02)
}

func (s *SpheroDriver) Roll(speed uint8, heading uint16) {
	s.packetChannel <- s.craftPacket([]uint8{speed, uint8(heading >> 8), uint8(heading & 0xFF), 0x01}, 0x02, 0x30)
}

func (s *SpheroDriver) Stop() {
	s.Roll(0, 0)
}

// ConfigureCollisionDetectionRaw allows custom collision detection sensitivity.
// see: http://orbotixinc.github.io/Sphero-Docs/docs/collision-detection/index.html
// deadTime - post-collision dead time in 10ms increments
func (s *SpheroDriver) ConfigureCollisionDetectionRaw(xThreshold, xSpeed, yThreshold, ySpeed, deadTime uint8) {
	// Meth 0x01 to enable, 0x00 to disable
	s.packetChannel <- s.craftPacket([]uint8{0x01, xThreshold, xSpeed, yThreshold, ySpeed, deadTime}, 0x02, 0x12)
}

// See http://orbotixinc.github.io/Sphero-Docs/docs/sphero-api/bootloader-and-sphero.html
//DID,CID,SEQ,DLEN,N,M,MASK,PCNT,MASK2
//x02,x11,8bit,0eh,16bit,16bit,32bit,8bit,32bit
//        ff fc 02 11 23 0e 00 c8 00 01 00 00 00 00 00 0f 80 00 00 c8 
func (s *SpheroDriver) ConfigureDataStreaming(seq uint8, n, m uint16, mask uint32, pcnt uint8, mask2 uint32) {
	// Meth 0x01 to enable, 0x00 to disable
	s.packetChannel <- s.specialCraftPacket([]uint8{ //seq, 0x0e, 
		uint8(n >> 8 & 0xFF), uint8(n & 0xFF),
		uint8(m >> 8 & 0xFF), uint8(m & 0xFF),
		uint8(mask >> 24 & 0xFF), uint8(mask >> 16 & 0xFF),
		uint8(mask >> 8 & 0xFF), uint8(mask & 0xFF),
		pcnt,
		uint8(mask2 >> 24 & 0xFF), uint8(mask2 >> 16 & 0xFF),
		uint8(mask2 >> 8 & 0xFF), uint8(mask2 & 0xFF)},
		0xFC,
		0x02, //DID
		0x11)  //CID
}

// 2 is good for updates per second
func (s *SpheroDriver) ConfigureLocatorStreaming(updatesPerSecond int) { 
	mask2 := ( LocatorMask2 | AccelOneMask2 | VelocityMask2 )
	//mask2 := uint32(0xaa800000)
	updates := uint16(400 / updatesPerSecond)
	log.Printf("Sending Data Streaming %v Mask2\n", mask2)
	s.ConfigureDataStreaming(0xAB, updates, 1, 0, 0, mask2)
}



func (s *SpheroDriver) configureDefaultCollisionDetection() {
	s.ConfigureCollisionDetectionRaw(0x40, 0x40, 0x50, 0x50, 0x60)
}

func (s *SpheroDriver) enableStopOnDisconnect() {
	s.packetChannel <- s.craftPacket([]uint8{0x00, 0x00, 0x00, 0x01}, 0x02, 0x37)
}

func (s *SpheroDriver) handleCollisionDetected(data []uint8) {
	// 22 = 5 byte async header + 16 bytes of data + 1 byte checksum
	if len(data) == 22 && data[4] == 17 {
		checksum := data[len(data)-1]
		if checksum == calculateChecksum(data[2:len(data)-1]) {
			buffer := bytes.NewBuffer(data[5:])

			var collision Collision
			binary.Read(buffer, binary.BigEndian, &collision)
			gobot.Publish(s.Event("collision"), collision)
			return
		}
	}
	gobot.Publish(s.Event("collision"), data)
}

func (s *SpheroDriver) handleLocator(data []uint8) {
	// 22 = 5 byte async header + 16 bytes of data + 1 byte checksum
	checksum := data[len(data)-1]
	if checksum == calculateChecksum(data[2:len(data)-1]) {
		// SOP+SOP+ID+DLEN+DLEN
		buffer := bytes.NewBuffer(data[5:])		
		var locator Locator
		binary.Read(buffer, binary.BigEndian, &locator)
		gobot.Publish(s.Event("locator"), locator)
		return
	} else {
		log.Printf("handleLocator failed to parse %v\n", data)
	}
}



func (s *SpheroDriver) getSyncResponse(packet *packet) []byte {
	s.packetChannel <- packet
	for i := 0; i < 500; i++ {
		for key := range s.syncResponse {
			if s.syncResponse[key][3] == packet.header[4] && len(s.syncResponse[key]) > 6 {
				var response []byte
				response, s.syncResponse = s.syncResponse[len(s.syncResponse)-1], s.syncResponse[:len(s.syncResponse)-1]
				return response
			}
		}
		time.Sleep(100 * time.Microsecond)
	}

	return []byte{}
}


func (s *SpheroDriver) specialCraftPacket(body []uint8, sop2 byte, did byte, cid byte) *packet {
	packet := new(packet)
	packet.body = body
	dlen := len(packet.body) + 1
	packet.header = []uint8{0xFF, sop2, did, cid, s.seq, uint8(dlen)}
	packet.checksum = s.calculateChecksum(packet)
	return packet
}

func (s *SpheroDriver) craftPacket(body []uint8, did byte, cid byte) *packet {
	packet := new(packet)
	packet.body = body
	dlen := len(packet.body) + 1
	packet.header = []uint8{0xFF, 0xFF, did, cid, s.seq, uint8(dlen)}
	packet.checksum = s.calculateChecksum(packet)
	return packet
}



func (s *SpheroDriver) write(packet *packet) {
	buf := append(packet.header, packet.body...)
	buf = append(buf, packet.checksum)
	length, err := s.adaptor().sp.Write(buf)
	if err != nil {
		fmt.Println(s.Name, err)
		s.adaptor().Disconnect()
		fmt.Println("Reconnecting to SpheroDriver...")
		s.adaptor().Connect()
		return
	} else if length != len(buf) {
		fmt.Println("Not enough bytes written", s.Name)
	}
	s.seq++
}

func (s *SpheroDriver) calculateChecksum(packet *packet) uint8 {
	buf := append(packet.header, packet.body...)
	return calculateChecksum(buf[2:])
}

func calculateChecksum(buf []byte) uint8 {
	var calculatedChecksum uint16
	for i := range buf {
		calculatedChecksum += uint16(buf[i])
	}
	return uint8(^(calculatedChecksum % 256))
}

func (s *SpheroDriver) readHeader() []uint8 {
	return s.readNextChunk(5)
}

func (s *SpheroDriver) readBody(length uint8) []uint8 {
	return s.readNextChunk(int(length))
}

func (s *SpheroDriver) readNextChunk(length int) []uint8 {
	var read = make([]uint8, length)
	var bytesRead = 0

	for bytesRead < length {
		time.Sleep(1 * time.Millisecond)
		n, err := s.adaptor().sp.Read(read[bytesRead:])
		if err != nil {
			// log.Printf("readNextChunk: %v\n", err)
			return nil
		}
		bytesRead += n
		// log.Printf("readNextChunk: %d %x\n", n, read[:bytesRead])
	}
	return read
}
