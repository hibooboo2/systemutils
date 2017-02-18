package activity

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"syscall"
	"unsafe"
)

func NewDevices() ([]*InputDevice, error) {
	var ret []*InputDevice

	if err := checkRoot(); err != nil {
		return ret, err
	}

	for i := 0; i < MAX_FILES; i++ {
		buff, err := ioutil.ReadFile(fmt.Sprintf(INPUTS, i))
		if err != nil {
			break
		}
		ret = append(ret, newInputDeviceReader(buff, i))
	}

	return ret, nil
}

func checkRoot() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	if u.Uid != "0" {
		return fmt.Errorf("Cannot read device files. Are you running as root?")
	}
	return nil
}

func newInputDeviceReader(buff []byte, id int) *InputDevice {
	rd := bufio.NewReader(bytes.NewReader(buff))
	rd.ReadLine()
	dev, _, _ := rd.ReadLine()
	splt := strings.Split(string(dev), "=")

	return &InputDevice{
		Id:   id,
		Name: splt[1],
	}
}

func NewKeyLogger(dev *InputDevice) *KeyLogger {
	return &KeyLogger{
		dev: dev,
	}
}

func (t *KeyLogger) Read() (chan InputEvent, error) {
	ret := make(chan InputEvent, 512)

	if err := checkRoot(); err != nil {
		close(ret)
		return ret, err
	}

	fd, err := os.Open(fmt.Sprintf(DEVICE_FILE, t.dev.Id))
	if err != nil {
		close(ret)
		return ret, fmt.Errorf("Error opening device file:", err)
	}

	go func() {

		tmp := make([]byte, eventsize)
		event := InputEvent{}
		for {

			n, err := fd.Read(tmp)
			if err != nil {
				panic(err)
				close(ret)
				break
			}
			if n <= 0 {
				continue
			}

			if err := binary.Read(bytes.NewBuffer(tmp), binary.LittleEndian, &event); err != nil {
				panic(err)
			}

			ret <- event

		}
	}()
	return ret, nil
}

func (t *InputEvent) KeyString() string {
	return keyCodeMap[t.Code]
}

var keyCodeMap map[uint16]string

func init() {
	keyCodeMap = map[uint16]string{
		1:   "ESC",
		2:   "1",
		3:   "2",
		4:   "3",
		5:   "4",
		6:   "5",
		7:   "6",
		8:   "7",
		9:   "8",
		10:  "9",
		11:  "0",
		12:  "-",
		13:  "=",
		14:  "BS",
		15:  "TAB",
		16:  "Q",
		17:  "W",
		18:  "E",
		19:  "R",
		20:  "T",
		21:  "Y",
		22:  "U",
		23:  "I",
		24:  "O",
		25:  "P",
		26:  "[",
		27:  "]",
		28:  "ENTER",
		29:  "L_CTRL",
		30:  "A",
		31:  "S",
		32:  "D",
		33:  "F",
		34:  "G",
		35:  "H",
		36:  "J",
		37:  "K",
		38:  "L",
		39:  ";",
		40:  "'",
		41:  "`",
		42:  "L_SHIFT",
		43:  "\\",
		44:  "Z",
		45:  "X",
		46:  "C",
		47:  "V",
		48:  "B",
		49:  "N",
		50:  "M",
		51:  ",",
		52:  ".",
		53:  "/",
		54:  "R_SHIFT",
		55:  "*",
		56:  "L_ALT",
		57:  "SPACE",
		58:  "CAPS_LOCK",
		59:  "F1",
		60:  "F2",
		61:  "F3",
		62:  "F4",
		63:  "F5",
		64:  "F6",
		65:  "F7",
		66:  "F8",
		67:  "F9",
		68:  "F10",
		69:  "NUM_LOCK",
		70:  "SCROLL_LOCK",
		71:  "HOME",
		72:  "UP_8",
		73:  "PGUP_9",
		74:  "-",
		75:  "LEFT_4",
		76:  "5",
		77:  "RT_ARROW_6",
		78:  "+",
		79:  "END_1",
		80:  "DOWN",
		81:  "PGDN_3",
		82:  "INS",
		83:  "DEL",
		84:  "",
		85:  "",
		86:  "",
		87:  "F11",
		88:  "F12",
		89:  "",
		90:  "",
		91:  "",
		92:  "",
		93:  "",
		94:  "",
		95:  "",
		96:  "R_ENTER",
		97:  "R_CTRL",
		98:  "/",
		99:  "PRT_SCR",
		100: "R_ALT",
		101: "",
		102: "Home",
		103: "Up",
		104: "PgUp",
		105: "Left",
		106: "Right",
		107: "End",
		108: "Down",
		109: "PgDn",
		110: "Insert",
		111: "Del",
		112: "",
		113: "",
		114: "",
		115: "",
		116: "",
		117: "",
		118: "",
		119: "Pause",
	}
}

const (
	INPUTS        = "/sys/class/input/event%d/device/uevent"
	DEVICE_FILE   = "/dev/input/event%d"
	MAX_FILES     = 255
	MAX_NAME_SIZE = 256
)

//event types
const (
	EV_SYN       = 0x00
	EV_KEY       = 0x01
	EV_REL       = 0x02
	EV_ABS       = 0x03
	EV_MSC       = 0x04
	EV_SW        = 0x05
	EV_LED       = 0x11
	EV_SND       = 0x12
	EV_REP       = 0x14
	EV_FF        = 0x15
	EV_PWR       = 0x16
	EV_FF_STATUS = 0x17
	EV_MAX       = 0x1f
)

var eventsize = int(unsafe.Sizeof(InputEvent{}))

type KeyLogger struct {
	dev *InputDevice
}

type InputDevice struct {
	Id   int
	Name string
}

type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}
