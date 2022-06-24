package scheduler

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log"
	"time"

	"github.com/cockroachdb/pebble"
	uid "github.com/google/uuid"
)

func init() {
	gob.Register(&Task{})
}

//Scheduler 任务调度器
//db-持久化DB,使用Pebble来存(rocksdb)
//running-是否在运行
//close-关闭信号
type Scheduler struct {
	db      *pebble.DB
	running bool
	close   chan struct{}
}

//Task 任务详情
//timeout-任务本身的执行超时时间,失败会重新执行
//interval-如果>0则为重复的间隔任务
type Task struct {
	Id   string
	Name string

	Timeout  time.Duration
	Interval time.Duration

	removed bool
	ctx     context.Context
	cancel  context.CancelFunc
	s       *Scheduler
}

//Stop 停止某个任务
func (t *Task) Stop() {
	t.removed = true
}

//run 内置函数,执行任务
//todo 未实现Name<->Function对应关系,模型已出,不是重点
func (t *Task) run() {
	log.Println("run task", t.Id, t.Timeout, t.Name, t.Interval)

	go func(task *Task) {
		defer task.cancel()
		time.Sleep(task.Interval)
		//TODO
	}(t)

	select {
	case <-t.ctx.Done():
		if t.ctx.Err() == context.Canceled {
			log.Println("task done", t.Id, t.Timeout, t.Name, t.Interval)
			if t.Interval <= 0 || t.removed {
				if err := t.s.db.Delete([]byte(t.Id), &pebble.WriteOptions{}); err != nil {
					log.Println("remove task failed", t, err)
				}
				return
			}
		} else {
			log.Println("task timeout", t.Id, t.Timeout, t.Name, t.Interval)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	t.cancel = cancel
	t.ctx = ctx
	go t.run()
}

//New 新建一个调度器,随便起个名字,支持持久化
func New(name string) (*Scheduler, error) {
	db, err := pebble.Open(name, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &Scheduler{db, false, make(chan struct{})}, nil
}

//Add 新增一个任务
//name-与Function对应,暂未实现,没必要
//timeout-任务本身的执行超时时间
//repeatInterval-反复执行的间隔时间(是否为重复任务,默认不填)
func (s *Scheduler) Add(name string, timeout time.Duration, repeatInterval ...time.Duration) (*Task, error) {
	var buf bytes.Buffer
	var err error
	var interval time.Duration

	if len(repeatInterval) > 0 {
		interval = repeatInterval[0]
	}

	if interval < 0 {
		return nil, errors.New("interval must >= 0")
	}

	if name == "" {
		return nil, errors.New("name was nil")
	}

	t := s.newTask(uid.NewString(), name, timeout, interval)

	err = gob.NewEncoder(&buf).Encode(t)

	if err != nil {
		return nil, err
	}

	err = s.db.Set([]byte(t.Id), buf.Bytes(), &pebble.WriteOptions{})

	if s.running {
		go t.run()
	}

	return t, err
}

//Remove 删除并停止任务
func (s *Scheduler) Remove(t *Task) error {
	if t == nil {
		return errors.New("task was nil")
	}

	t.removed = true
	t.Interval = 0
	t.cancel()

	return s.db.Delete([]byte(t.Id), &pebble.WriteOptions{})
}

//List 列出所有任务
func (s *Scheduler) List() []*Task {
	iter := s.db.NewIter(&pebble.IterOptions{})

	var tasks []*Task

	for iter.First(); iter.Valid(); iter.Next() {
		t := &Task{s: s}
		if err := gob.NewDecoder(bytes.NewReader(iter.Value())).Decode(t); err != nil {
			log.Println("load task failed", iter.Key(), iter.Value(), err)
		}
		t.ctx, t.cancel = context.WithTimeout(context.Background(), t.Timeout)
		tasks = append(tasks, t)
	}
	_ = iter.Close()

	return tasks
}

//Run 跑起来~
func (s *Scheduler) Run() {
	if s.running {
		return
	}

	s.running = true

	for _, t := range s.List() {
		go t.run()
	}

	<-s.close
}

//Stop 终止调度器
func (s *Scheduler) Stop() {
	close(s.close)
}

//newTask 内置函数
func (s *Scheduler) newTask(id, name string, timeout, interval time.Duration) *Task {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &Task{id, name, timeout, interval, false, ctx, cancel, s}
}
