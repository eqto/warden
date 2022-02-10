package warden

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/eqto/every"
	"github.com/eqto/go-json"
	log "github.com/eqto/go-logger"
	"github.com/eqto/service"
)

var counterMap = make(map[string]int)

func Run() error {
	run()
	every.Minutes().Do(func(c *every.Context) {
		run()
	})
	service.Wait()

	return nil
}

func run() {
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

func processConfig(filename string) {
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

	out, e := exec.Command(`bash`, `-c`, fmt.Sprintf(`ps ax | grep "[%s]%s" | wc -l`, name[:1], name[1:])).Output()
	if e != nil {
		log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
		return
	}
	count, e := strconv.Atoi(strings.TrimSpace(string(out)))
	if e != nil {
		log.W(fmt.Sprintf(`%v, file: %s`, e, filename))
		return
	}
	prevCount := counterMap[name]

	counterMap[name] = count
	if prevCount != count {
		body := js.GetJSONObject(`notify.body`)
		strBody := strings.ReplaceAll(body.ToString(), `[count]`, strconv.Itoa(count))

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
