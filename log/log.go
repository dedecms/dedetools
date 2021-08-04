package log

import (
	"fmt"
	"sync"
	"time"

	"github.com/ahmetalpbalkan/go-cursor"
	"github.com/dedecms/snake"
	"github.com/i582/cfmt/cmd/cfmt"
	"github.com/leaanthony/synx"
)

type Msg struct {
	msgid   string
	start   string
	time    time.Time
	message *synx.String
	num     int
	sum     int
	group   *group
}

// SpinnerGroup is a group of Spinners
type group struct {
	sync.Mutex
	sync.WaitGroup
	spinners []*Msg
	running  bool
	drawn    bool
}

func Start(message string, sum ...int) *Msg {
	l := new(Msg)
	l.message = synx.NewString(cfmt.Sprint(message))
	l.time = time.Now()
	l.start = l.time.Format("[2006-01-02 15:04:05.000000] ")
	l.sum = 0
	if len(sum) > 0 {
		l.sum = sum[0]
		l.num = 0
		cfmt.Print(snake.String("\r").Add(l.start).Add("{{[-]}}::yellow ").Add(message).Add(" [", l.num, "/", l.sum, "]").Get())
	} else {
		cfmt.Print(snake.String("\r").Add(l.start).Add("{{[-]}}::yellow ").Add(message).Get())
	}

	return l
}

func (l *Msg) Time() time.Time {
	return l.time
}

func (l *Msg) GetMessage() string {
	return l.message.GetValue()
}

func (l *Msg) Add() {
	l.num++
	cfmt.Print(snake.String("\r").Add(l.start).Add("{{[-]}}::yellow ").Add(l.message.GetValue()).Add(" [", l.num, "/", l.sum, "]").Get())
}
func (l *Msg) Err(err error) {
	if l.sum > 0 {
		cfmt.Println(snake.String("\r").Add(l.start).Add("{{[X]}}::red ").Add(l.message.GetValue()).Add(" [", l.num, "/", l.sum, "] ").Add(err.Error()).Get())
	} else {
		cfmt.Println(snake.String("\r").Add(l.start).Add("{{[X]}}::red ").Add(l.message.GetValue()).Add(err.Error()).Get())
	}
}

func (l *Msg) Done() {
	if l.sum != 0 {
		cfmt.Println(snake.String("\r").Add(l.start).Add("{{[✓]}}::green ").Add(l.message.GetValue()).Add(" [", l.num, "/", l.sum, "]").Add(" (").Add(time.Since(l.time)).Add(")").Get())
	} else {
		cfmt.Println(snake.String("\r").Add(l.start).Add("{{[✓]}}::green ").Add(l.message.GetValue()).Add(" (").Add(time.Since(l.time)).Add(")").Get())
	}
}

// UpdateMessage updates the spinner message
func (l *Msg) Message(message string) {
	l.message.SetValue(message)
}

// Success marks spinner as success and update message
func (l *Msg) Success(message string) {
	l.Message(cfmt.Sprint(snake.String(l.start).Add("{{[✓]}}::green ").Add("* ").Add(message).Add(" (").Add(time.Since(l.time)).Add(")").Get()))
	l.stop()
}

// Error marks spinner as error and update message
func (l *Msg) Error(message string) {
	l.Message(message)
	l.stop()
}

// Error marks spinner as error and update message
func (l *Msg) IsGroup() bool {
	return l.group != nil
}

func (l *Msg) stop() {
	l.group.redraw()
	l.group.Done()
}

func (l *Msg) sprint() string {
	return fmt.Sprint(l.message.GetValue())
}

// At returns the Spinner at given 0-based index
func (g *group) Get(key int, message string) *Msg {
	l := g.spinners[key]
	if l.msgid == snake.String(message).MD5() {
		return l
	}
	return nil
}

// Start the spinners
func (g *group) Start() {
	g.Add(len(g.spinners))
	g.Lock()
	defer g.Unlock()

	if g.running {
		return
	}
	g.running = true

	go g.redraw()
}

// Stop the spinners
func (g *group) Stop() {
	g.Lock()
	defer g.Unlock()
	g.running = false
}

// Wait for all spinners to finish
func (g *group) Wait() {
	g.WaitGroup.Wait()
	g.Stop()
}

func (g *group) redraw() {
	g.Lock()
	defer g.Unlock()
	if !g.running {
		return
	}
	if g.drawn {
		fmt.Print(cursor.MoveUp(len(g.spinners)))
	}
	for _, spinner := range g.spinners {
		fmt.Print(cursor.ClearEntireLine())
		fmt.Println(spinner.sprint())
	}
	g.drawn = true
}

// Newgroup creates a group
func NewGroup(sum int) *group {
	group := &group{
		spinners: make([]*Msg, sum),
		running:  false,
		drawn:    false,
	}
	return group
}

// Wait for all spinners to finish
func (g *group) Item(key int, message string) {
	now := time.Now()
	start := now.Format("[2006-01-02 15:04:05.000000] ")
	g.spinners[key] = &Msg{
		msgid:   snake.String(message).MD5(),
		message: synx.NewString(cfmt.Sprint(snake.String(start).Add("{{[-]}}::yellow ").Add("* ").Add(message).Get())),
		time:    now,
		start:   start,
		group:   g,
	}
}
