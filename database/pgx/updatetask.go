package pgx

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type DBUpdateTask struct {
	lock        sync.Mutex
	delayTimer  *time.Timer
	changes     map[string]interface{}
	Table       string
	Where       map[string]interface{}
	UpdateDelay time.Duration
	*Scheme
}

func (task *DBUpdateTask) AddTask(column string, value interface{}) {
	task.lock.Lock()
	defer task.lock.Unlock()

	if task.changes == nil {
		task.changes = map[string]interface{}{}
	}
	task.changes[column] = value

	if task.UpdateDelay <= 0 {
		task.UpdateDelay = time.Minute
	}
	if task.delayTimer != nil {
		task.delayTimer.Stop()
	}
	task.delayTimer = time.AfterFunc(task.UpdateDelay, func() {
		task.Update()
	})
}

func (task *DBUpdateTask) Update() (err error) {
	task.lock.Lock()
	defer task.lock.Unlock()

	if len(task.Where) == 0 || len(task.changes) == 0 {
		return
	}

	var sets []string
	var values []interface{}
	for col, val := range task.changes {
		sets = append(sets, fmt.Sprintf(`"%s"=?`, col))
		values = append(values, val)
	}

	whereSql, whereValues := ParseWhere(task.Where, nil)
	_, err = task.Exec(SQLFormat(`UPDATE "%s"."%s" SET %s %s`, task.String(), task.Table, strings.Join(sets, ","), whereSql), append(values, whereValues...)...)
	if err == nil {
		task.changes = nil
	}

	return
}
