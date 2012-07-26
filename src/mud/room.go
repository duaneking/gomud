package mud

import ("fmt"
	"strconv")

type RoomID int
type RoomSide int

type RoomConnCreator func() *SimpleRoomConnection
type RoomConnector func(a *Room, b *Room) *SimpleRoomConnection

const (
	SideA RoomSide = iota
	SideB
)

type RoomExitInfo struct {
	exit RoomConnection
	exitSide RoomSide
}

type Room struct {
	Persister
	id int
	text string
	players map[int]Player
	perceivers []Perceiver
	Persistents []Persister
	physObjects []PhysicalObject
	exits []RoomExitInfo
	stimuliBroadcast chan Stimulus
	interactionQueue chan InterObjectAction
	universe *Universe
}

type RoomConnection interface {
	RoomA() *Room
	RoomB() *Room
	AExitName() string
	BExitName() string
}

type SimpleRoomConnection struct {
	RoomConnection
	roomA, roomB *Room
	aExitName, bExitName string
}

func (rc SimpleRoomConnection) RoomA() *Room { return rc.roomA }
func (rc SimpleRoomConnection) RoomB() *Room { return rc.roomB }
func (rc SimpleRoomConnection) AExitName() string { return rc.aExitName }
func (rc SimpleRoomConnection) BExitName() string { return rc.bExitName }

func SimpleRoomConnectCreator(a string, b string) RoomConnCreator {
	return func() *SimpleRoomConnection {
		conn := new(SimpleRoomConnection)
		conn.aExitName, conn.bExitName = a, b
		return conn
	}
}

var EastWestRoomConnection = SimpleRoomConnectCreator("east","west")
var NorthSouthRoomConnection = SimpleRoomConnectCreator("north","south")
var UpDownRoomConnection = SimpleRoomConnectCreator("up","down")

func ConnectWithConnCreator(exitGen RoomConnCreator) RoomConnector {
	return func(a *Room, b *Room) *SimpleRoomConnection {
		roomConn := exitGen()
		roomConn.roomA = a
		roomConn.roomB = b
		reiA := RoomExitInfo{exitSide: SideA, exit: roomConn}
		reiB := RoomExitInfo{exitSide: SideB, exit: roomConn}
		a.exits = append(a.exits, reiA)
		b.exits = append(b.exits, reiB)
		return roomConn
	}
}

var ConnectEastWest = ConnectWithConnCreator(EastWestRoomConnection)
var ConnectNorthSouth = ConnectWithConnCreator(NorthSouthRoomConnection)
var ConnectUpDown = ConnectWithConnCreator(UpDownRoomConnection)

func (r *Room) ActionQueue() {
	for {
		action := <- r.interactionQueue
		action.Exec()
	}
}

func (r *Room) FanOutBroadcasts() {
	for {
		broadcast := <- r.stimuliBroadcast
		for _,p := range r.perceivers { 
			p.StimuliChannel() <- broadcast 
		}
	}
}

func (r *RoomExitInfo) Name() string {
	if(r.exitSide == SideA) {
		return r.exit.AExitName()
	} else {
		return r.exit.BExitName()
	}
	return ""
}

func (r *RoomExitInfo) OtherSide() *Room {
	if(r.exitSide == SideA) {
		return r.exit.RoomB()
	} else {
		return r.exit.RoomA()
	}
	return nil
}

func (r *Room) Describe(toPlayer *Player) string {
	roomText := r.text
	objectsText := r.DescribeObjects(toPlayer)
	playersText := r.DescribePlayers(toPlayer)
	
	return roomText + Divider() + objectsText + Divider() + playersText
}

func (r *Room) DescribeObjects(toPlayer *Player) string {
	objTextBuf := "Sitting here is/are:\n"
	for _,obj := range r.physObjects {
		if obj != nil && obj.Visible() {
			objTextBuf += obj.Description()
			objTextBuf += "\n"
		}
	}
	return objTextBuf
}

func (r *Room) DescribePlayers(toPlayer *Player) string {
	objTextBuf := "Other people present:\n"
	for _,player := range r.players {
		if player.id != toPlayer.id {
			objTextBuf += player.name
			objTextBuf += "\n"
		}
	}
	return objTextBuf
}

func (r *Room) AddPhysObj(p PhysicalObject) {
	r.physObjects = append(r.physObjects, p)
	fmt.Println("[add] new physObjects = ",r.physObjects)
}

func (r *Room) AddPerceiver(p Perceiver) {
	r.perceivers = append(r.perceivers, p)
	fmt.Println("[add] new perceivers = ",r.perceivers)
}

func (r *Room) RemovePerceiver(p Perceiver) {
	for i,listP := range(r.perceivers) {
		if p == listP {
			if len(r.perceivers) > 1 {
				perceivers := append(r.perceivers[:i],r.perceivers[i+1:]...)
				r.perceivers = perceivers
			} else {
				r.perceivers = []Perceiver{}
			}
			fmt.Println("[rm] new perceivers = ",r.perceivers)
			break
		}
	}
}

func (r *Room) AddPersistent(p Persister) {
	r.Persistents = append(r.Persistents, p)
}

func (r *Room) RemovePersistent(p Persister) {
	for i,listP := range(r.Persistents) {
		if p == listP {
			if len(r.Persistents) > 1 {
				perceivers := append(r.Persistents[:i],r.Persistents[i+1:]...)
				r.Persistents = perceivers
			} else {
				r.Persistents = []Persister{}
			}
			fmt.Println("[rm] new Persistents = ",r.Persistents)
			break
		}
	}
}

func (r *Room) Broadcast(s Stimulus) {
	r.stimuliBroadcast <- s
}

func (r *Room) PersistentValues() map[string]interface{} {
	vals := make(map[string]interface{})
	if(r.id > 0) {
		vals["id"] = strconv.Itoa(r.id)
	}
	vals["persisters"] = r.Persistents
	return vals
}

func (r *Room) Save() string {
	outID := r.universe.Store.SaveStructure("room",r.PersistentValues())
	if(r.id == 0) {
		r.id, _ = strconv.Atoi(outID)
	}
	return outID
}

func NewBasicRoom(universe *Universe, rid int, rtext string, physObjs []PhysicalObject) *Room {
	r := Room{id: rid, text: rtext, universe: universe}
	r.stimuliBroadcast = make(chan Stimulus, 10)
	r.interactionQueue = make(chan InterObjectAction, 10)
	r.players = make(map[int]Player)
	r.perceivers = []Perceiver{}
	r.Persistents = []Persister{}
	r.physObjects = physObjs
	r.exits = []RoomExitInfo{}
	universe.Rooms[r.id] = &r
	universe.Persistents = append(universe.Persistents, &r)

	return &r
}