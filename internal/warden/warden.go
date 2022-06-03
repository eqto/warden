package warden

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/eqto/every"
	"github.com/eqto/go-json"
	log "github.com/eqto/go-logger"
	"github.com/eqto/service"
)

var (
	counterMap  = make(map[string]int)
	counterLock = sync.RWMutex{}
)

func Run() error {
	defer service.HandlePanic()
	go run()
	every.Minutes().Do(func(c *every.Context) {
		run()
	})
	service.Wait()

	return nil
}

func run() {
	defer service.HandlePanic()
	entries, e := os.ReadDir(`configs`)
	if e != nil {
		log.W(e)
		return
	}
	for _, dir := range entries {
		if !dir.IsDir() && strings.HasSuffix(dir.Name(), `.json`) {
			go processConfig(dir.Name())
		}
	}
}

func readCounter(name string) int {
	counterLock.RLock()
	defer counterLock.RUnlock()
	return counterMap[name]
}

func writeCounter(name string, value int) {
	counterLock.Lock()
	defer counterLock.Unlock()
	counterMap[name] = value
}

func processConfig(filename string) {
	defer service.HandlePanic()
	js, e := json.ParseFile(`configs/` + filename)
	if e != nil {
		log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
		return
	}
	name := js.GetString(`name`)
	if name == `` {
		log.W(fmt.Sprintf(`Empty name, file: %s`, filename))
		return
	}

	out, e := exec.Command(`bash`, `-c`, fmt.Sprintf(`ps ax | grep -E "\W*%s(\s+|$)" | wc -l`, name)).Output()
	if e != nil {
		log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
		return
	}
	count, e := strconv.Atoi(strings.TrimSpace(string(out)))
	if e != nil {
		log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
		return
	}
	prevCount := readCounter(name)

	writeCounter(name, count)

	if prevCount != count {
		body := js.GetJSONObject(`notify.body`)
		strBody := strings.ReplaceAll(body.ToString(), `[count]`, strconv.Itoa(count))
		strBody = strings.ReplaceAll(strBody, `[name]`, name)

		resp, e := http.Post(js.GetString(`notify.url`), js.GetString(`notify.content_type`), strings.NewReader(strBody))
		if e != nil {
			log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
			return
		}
		bodyResp, e := io.ReadAll(resp.Body)
		if e != nil {
			log.D(string(bodyResp))
			log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
			return
		}
	}
}
