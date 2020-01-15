package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"os"
	"sync"
)

type UserAgentManager struct {
	pp *ring.Ring
}

var singleton *UserAgentManager
var once sync.Once

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func GetUAManager() *UserAgentManager {
	once.Do(func() {
		coffee, _ := readLines(UA_FILE)
		fmt.Printf("[-] Loaded %d User-Agents\n\n", len(coffee))
		var s UserAgentManager
		s.pp = ring.New(len(coffee))
		// populate the ring buffer
		for i := 0; i < s.pp.Len(); i++ {
			s.pp.Value = coffee[i]
			s.pp = s.pp.Next()
		}
		singleton = &s
		//singleton = &UserAgentManager{*ring.New(len(coffee))}
	})
	return singleton
}

func (m *UserAgentManager) GetUserAgent() string {
	m.pp = m.pp.Next()
	return m.pp.Value.(string)
}
