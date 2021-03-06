package delay

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"goto/pkg/util"

	"github.com/gorilla/mux"
)

var (
  Handler     util.ServerHandler       = util.ServerHandler{"delay", SetRoutes, Middleware}
  delayByPort map[string]time.Duration = map[string]time.Duration{}
  delayCountByPort map[string]int = map[string]int{}
  delayLock   sync.RWMutex
)

func SetRoutes(r *mux.Router, parent *mux.Router, root *mux.Router) {
  delayRouter := r.PathPrefix("/delay").Subrouter()
  util.AddRoute(delayRouter, "/set/{delay}", setDelay, "POST", "PUT")
  util.AddRoute(delayRouter, "/clear", setDelay, "POST", "PUT")
  util.AddRoute(delayRouter, "", getDelay, "GET")
}

func setDelay(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  delayParam := strings.Split(vars["delay"], ":")
  listenerPort := util.GetListenerPort(r)
  delayLock.Lock()
  defer delayLock.Unlock()
  delayCountByPort[listenerPort] = -1
  delayByPort[listenerPort] = 0
  msg := ""
  if len(delayParam[0]) > 0 {
    if delay, err := time.ParseDuration(delayParam[0]); err == nil {
      delayByPort[listenerPort] = delay
      if delay > 0 {
        delayCountByPort[listenerPort] = 0
      }
      if len(delayParam) > 1 {
        times, _ := strconv.ParseInt(delayParam[1], 10, 32)
        delayCountByPort[listenerPort] = int(times)
      }
      if delayCountByPort[listenerPort] > 0 {
        msg = fmt.Sprintf("Will delay next %d requests with %s", delayCountByPort[listenerPort], delayByPort[listenerPort])
      } else if delayCountByPort[listenerPort] == 0 {
        msg = fmt.Sprintf("Will delay requests with %s until reset", delayByPort[listenerPort])
      } else {
        msg = "Delay cleared"
      }
    } else {
      msg = "Invalid delay param"
    }
  } else {
    msg = "Delay cleared"
  }
  w.WriteHeader(http.StatusAccepted)
  util.AddLogMessage(msg, r)
  fmt.Fprintln(w, msg)
}

func getDelay(w http.ResponseWriter, r *http.Request) {
  delayLock.RLock()
  defer delayLock.RUnlock()
  delay := delayByPort[util.GetListenerPort(r)]
  msg := fmt.Sprintf("Current delay: %s\n", delay.String())
  w.WriteHeader(http.StatusOK)
  util.AddLogMessage(msg, r)
  fmt.Fprintln(w, msg)
}

func Middleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    delayLock.RLock()
    listenerPort := util.GetListenerPort(r)
    delay := delayByPort[listenerPort]
    delayCount := delayCountByPort[listenerPort]
    delayLock.RUnlock()
    if delay > 0 && delayCount >= 0 && !util.IsAdminRequest(r) {
      util.AddLogMessage(fmt.Sprintf("Delaying for = %s", delay.String()), r)
      if delayCount > 0 {
        if delayCount == 1 {
          delayCount = -1
          delayByPort[listenerPort] = 0
        } else {
          delayCount--
        }
        util.AddLogMessage(fmt.Sprintf("Remaining delay count = %d", delayCount), r)
        delayCountByPort[listenerPort] = delayCount
      }
      time.Sleep(delay)
    }
    next.ServeHTTP(w, r)
  })
}
