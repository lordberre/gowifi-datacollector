package queue

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

const (
	MAX_QUEUE_SIZE = 32
	queueDir       = "/tmp/gowifi_datacollector/"
)

type Queue struct {
	content        [MAX_QUEUE_SIZE][]byte
	readHead       int
	writeHead      int
	len            int
	persistentFile string
}

func NewQueue() Queue {
	cmd := exec.Command("mkdir", "-p", queueDir)
	cmd.Wait()
	_, err := cmd.Output()
	if err != nil {
		log.Errorf("[NewQueue()] %s\n", err.Error())
	}
	queue := Queue{
		persistentFile: queueDir + "queue.json",
	}
	queue.ReadPersistent()
	return queue
}

func (q *Queue) GetLength() int {
	return q.len
}

func (q *Queue) Push(e []byte) bool {
	if q.len >= MAX_QUEUE_SIZE {
		return false
	}
	q.content[q.writeHead] = e
	q.writeHead = (q.writeHead + 1) % MAX_QUEUE_SIZE
	q.len++
	return true
}

func (q *Queue) Pop() ([]byte, bool) {
	if q.len <= 0 {
		return nil, false
	}
	result := q.content[q.readHead]
	q.content[q.readHead] = nil
	q.readHead = (q.readHead + 1) % MAX_QUEUE_SIZE
	q.len--
	return result, true
}

func (q *Queue) WritePersistent() {
	file, err := os.Create(q.persistentFile)

	if err != nil {
		log.Fatalf("failed creating queue file: %s", err)
	}

	defer file.Close()

	if q.len == 0 {
		return
	}

	w := bufio.NewWriter(file)

	for _, data := range q.content {
		if len(data) == 0 {
			continue
		}
		_, _ = w.WriteString(string(data) + "\n")
	}

	w.Flush()
}

func (q *Queue) ReadPersistent() {
	file, err := os.Open(q.persistentFile)
	if err != nil {
		log.Infof("failed reading queue file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	queue := make([][]byte, 0)
	for scanner.Scan() {
		queue = append(queue, scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if len(queue) > MAX_QUEUE_SIZE {
		queue = queue[:MAX_QUEUE_SIZE]
	}
	for i, item := range queue {
		q.content[i] = item
	}
	fmt.Printf("Loaded %d previously queued items\n", len(queue))
}
