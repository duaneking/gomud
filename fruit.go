package main

import ("mud"
	"fmt")

type Fruit struct {
//	mud.Persister
	mud.PhysicalObject
	mud.TimeListener
	universe *mud.Universe
	room *mud.Room
	name string
	ping chan int
	stage LifeStage
	lastChange int
	visible bool
	hasMadePlant bool
}

var fruitStages map[int]LifeStage

func init() {
	fruitStages = make(map[int]LifeStage)
	underripe := LifeStage{StageNo: 0, Name: "underripe", StageChangeDelay: 10000}
	ripe := LifeStage{StageNo: 1, Name: "ripe", StageChangeDelay: 40000}
	rotten := LifeStage{StageNo: 2, Name: "rotten", StageChangeDelay: 10000}
        pit := LifeStage{StageNo: 3, Name: "pit", StageChangeDelay: 10000}
	defunct := LifeStage{StageNo: 4, Name: "defunct", StageChangeDelay: -1}
	addLs(underripe, fruitStages)
	addLs(ripe, fruitStages)
	addLs(rotten, fruitStages)
	addLs(pit, fruitStages)
	addLs(defunct, fruitStages)
}

func (f Fruit) Visible() bool { return f.visible }
func (f Fruit) Carryable() bool { return true }
func (f Fruit) TextHandles() []string {
	return []string{f.name}
}
func (f Fruit) Description() string {
	return fmt.Sprintf("A(n) %s %s", f.stage.Name, f.name);
}

func (f *Fruit) SetRoom(r *mud.Room) { f.room = r }
func (f Fruit) Room() *mud.Room { return f.room }

type FruitTasteStimulus struct {
	mud.Stimulus
	f *Fruit
}

func (f *Fruit) Ping() chan int { return f.ping }
func (f *Fruit) Age(now int) {
	if(f.stage.StageChangeDelay > 0) {
		mud.Log("Age next stage clause, room =",f.Room())
		nextStage := (f.stage.StageNo + 1)
		f.stage = fruitStages[nextStage]
		f.lastChange = now
	} else if !f.hasMadePlant {
		mud.Log("Age MakePlant clause, room =",f.Room())
		p := MakePlant(f.universe, f.name)
		f.visible = false
		f.Room().AddChild(p)
		p.SetRoom(f.Room())
		f.hasMadePlant = true
	}
}

func (f *Fruit) UpdateTimeLoop() {
	for {
		now := <- f.ping
		if now > (f.lastChange + f.stage.StageChangeDelay) {
			f.Age(now)
		}
	}
}

func MakeFruit(u *mud.Universe, name string) *Fruit {
	f := new(Fruit)
	f.universe = u
	f.name = name
	f.ping = make(chan int)
	f.stage = fruitStages[0]
	f.visible = true

//	u.Persistents = append(u.Persistents, f)
	u.TimeListeners = append(u.TimeListeners, f)

	go f.UpdateTimeLoop()

	return f
}