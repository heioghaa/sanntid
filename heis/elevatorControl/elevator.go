package elevatorControl

import (
	"../liftio"
	"../liftnet"
	"../localQueue"
	"encoding/json"
	"log"
	"os"
	"time"
)

var myID int
var isIdle = true
var lastOrder = uint(0)
var toNetwork = make(chan liftnet.Message, 10)
var fromNetwork = make(chan liftnet.Message, 10)
var floorOrder = make(chan uint, 5) // floor orders to io
var setLight = make(chan liftio.Light, 5)
var liftStatus liftio.FloorStatus


func RunElevator() {

	buttonPress := make(chan liftio.Button, 5) // button presses from io
	status := make(chan liftio.FloorStatus, 5) // the lifts status

	myID = liftnet.Init(&toNetwork, &fromNetwork)
	liftio.Init(&floorOrder, &setLight, &status, &buttonPress)
	readQueueFromFile() // if no prev queue: "queue.txt doesn't exitst" entered in log
	liftStatus := <-status
	ticker1 := time.NewTicker(10 * time.Millisecond).C
	ticker2 := time.NewTicker(5 * time.Millisecond).C
	log.Println("Up and running. My is is: ", myID, "Penalty time is: ", (myID%10)*10)
	for {
		select {
		case button := <-buttonPress:
			newKeypress(button)
		case liftStatus = <-status:
			runQueue(liftStatus)
		case message := <-fromNetwork:
			newMessage(message)
			orderLight(message)
		case <-ticker1:
			checkTimeout()
		case <-ticker2:
			runQueue(liftStatus)
		}
	}
}

// Called from RunElevator 
func newKeypress(button liftio.Button) {
	switch button.Button {
	case liftio.Up:
		log.Println("Request up button pressed.", button.Floor)
		addMessage(button.Floor, true)
		setOrderLight(button.Floor, true, true)
	case liftio.Down:
		log.Println("Request down button ressed.", button.Floor)
		addMessage(button.Floor, false)
		setOrderLight(button.Floor, false, true)
	case liftio.Command:
		log.Println("Command button pressed.", button.Floor)
		addCommand(button.Floor)
	case liftio.Stop:
		log.Println("Stop button pressed")
		// action optional
	case liftio.Obstruction:
		log.Println("Obstruction")
		// action optional
	}

}

// Called from RunElevator
func runQueue(liftStatus liftio.FloorStatus) {
	floor := liftStatus.Floor
	if liftStatus.Running {
		if liftStatus.Direction {
			floor++
		} else {
			floor--
		}
	}
	order, direction := localQueue.GetOrder(floor, liftStatus.Direction)
	if liftStatus.Floor == order && liftStatus.Door {
		removeFromQueue(order, direction)
		lastOrder = 0
		liftStatus.Door = true
		time.Sleep(20 * time.Millisecond)
		if order, _ := localQueue.GetOrder(liftStatus.Floor, liftStatus.Direction); order == 0 {
			isIdle = true
		}
	} else {
		isIdle = false
		if lastOrder != order && !liftStatus.Door {
			lastOrder = order
			floorOrder <- order
		}
	}
}

// Called from runQueue
func removeFromQueue(floor uint, direction bool) {
	log.Println("Removing from queue", floor, direction)
	localQueue.DeleteLocalOrder(floor, direction)
	delMessage(floor, direction)
	setLight <- liftio.Light{floor, liftio.Command, false}
	setOrderLight(floor, direction, false)

}

// called from run loop 
func orderLight(message liftnet.Message) {
	switch message.Status {
	case liftnet.Done:
		setOrderLight(message.Floor, message.Direction, false)
	case liftnet.New:
		setOrderLight(message.Floor, message.Direction, true)
	case liftnet.Accepted:
		setOrderLight(message.Floor, message.Direction, true)
	}
}

// Called from orderLight
func setOrderLight(floor uint, direction bool, on bool) {
	if direction {
		setLight <- liftio.Light{floor, liftio.Up, on}
	} else {
		setLight <- liftio.Light{floor, liftio.Down, on}
	}
}

// Called by RunElevator and ReadQueuFromFile
func addCommand(floor uint) {
	localQueue.AddLocalCommand(floor)
	setLight <- liftio.Light{floor, liftio.Command, true}
}

// Called by RunElevator
func readQueueFromFile() {
	input, err := os.Open(localQueue.BackupFile)
	if err != nil {
		log.Println("Error in opening file: ", err)
		return
	}
	defer input.Close()
	byt := make([]byte, 23)
	dat, err := input.Read(byt)
	if err != nil {
		log.Println("Error in reading file: ", err)
		return
	}
	log.Println("Read ", dat, " bytes from file ")
	var cmd []bool
	if err := json.Unmarshal(byt, &cmd); err != nil {
		log.Println(err)
	}
	for i, val := range cmd {
		if val {
			addCommand(uint(i + 1))
		}
	}
}