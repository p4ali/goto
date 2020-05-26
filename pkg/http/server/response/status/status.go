package status

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"goto/pkg/http/server/intercept"
	"goto/pkg/util"

	"github.com/gorilla/mux"
)

type PortStatus struct {
  alwaysReportStatus      int
  alwaysReportStatusCount int
  countsByRequestedStatus map[int]int
  countsByReportedStatus  map[int]int
  lock                    sync.RWMutex
}

var (
  Handler       util.ServerHandler     = util.ServerHandler{"status", SetRoutes, Middleware}
  portStatusMap map[string]*PortStatus = map[string]*PortStatus{}
  statusLock    sync.RWMutex
)

func SetRoutes(r *mux.Router, parent *mux.Router) {
  statusRouter := r.PathPrefix("/status").Subrouter()
  util.AddRoute(statusRouter, "/set/{status}", setStatus, "PUT", "POST")
  util.AddRoute(statusRouter, "/counts/clear", clearStatusCounts, "PUT", "POST")
  util.AddRoute(statusRouter, "/counts/{status}", getStatusCount, "GET")
  util.AddRoute(statusRouter, "/counts", getStatusCount, "GET")
  util.AddRoute(statusRouter, "/clear", setStatus, "PUT", "POST")
  util.AddRoute(statusRouter, "", getStatus, "GET")
  util.AddRoute(parent, "/status/{status}", getStatus, "GET")
}

func getOrCreatePortStatus(r *http.Request) *PortStatus {
  listenerPort := util.GetListenerPort(r)
  portStatus := portStatusMap[listenerPort]
  if portStatus == nil {
    portStatus = &PortStatus{countsByRequestedStatus: map[int]int{}, countsByReportedStatus: map[int]int{}}
    portStatusMap[listenerPort] = portStatus
  }
  return portStatus
}

func setStatus(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  status := strings.Split(vars["status"], ":")
  statusLock.Lock()
  defer statusLock.Unlock()
  portStatus := getOrCreatePortStatus(r)
  portStatus.alwaysReportStatusCount = -1
  portStatus.alwaysReportStatus = 200
  if len(status[0]) > 0 {
    s, _ := strconv.ParseInt(status[0], 10, 32)
    if s > 0 {
      portStatus.alwaysReportStatus = int(s)
      portStatus.alwaysReportStatusCount = 0
      if len(status) > 1 {
        s, _ := strconv.ParseInt(status[1], 10, 32)
        portStatus.alwaysReportStatusCount = int(s)
      }
    }
  }
  msg := ""
  if portStatus.alwaysReportStatusCount > 0 {
    msg = fmt.Sprintf("Will respond with forced status: %d times %d", portStatus.alwaysReportStatus, portStatus.alwaysReportStatusCount)
  } else if portStatus.alwaysReportStatusCount == 0 {
    msg = fmt.Sprintf("Will respond with forced status: %d forever", portStatus.alwaysReportStatus)
  } else {
    msg = fmt.Sprintf("Will respond with given status")
  }
  util.AddLogMessage(msg, r)
  w.WriteHeader(http.StatusAccepted)
  fmt.Fprintln(w, msg)
}

func computeResponseStatus(originalStatus int, portStatus *PortStatus) int {
  responseStatus := originalStatus
  if portStatus.alwaysReportStatus > 0 && portStatus.alwaysReportStatusCount >= 0 {
    responseStatus = portStatus.alwaysReportStatus
    if portStatus.alwaysReportStatusCount > 0 {
      if portStatus.alwaysReportStatusCount == 1 {
        portStatus.alwaysReportStatusCount = -1
      } else {
        portStatus.alwaysReportStatusCount--
      }
    }
  }
  return responseStatus
}

func getStatus(w http.ResponseWriter, r *http.Request) {
  statusLock.Lock()
  defer statusLock.Unlock()
  requestedStatus, _ := util.GetIntParam(r, "status", 200)
  portStatus := getOrCreatePortStatus(r)
  portStatus.countsByRequestedStatus[requestedStatus]++
  util.AddLogMessage(fmt.Sprintf("Requested status: [%d]", requestedStatus), r)
  reportedStatus := computeResponseStatus(requestedStatus, portStatus)
  w.WriteHeader(reportedStatus)
}

func getStatusCount(w http.ResponseWriter, r *http.Request) {
  statusLock.RLock()
  defer statusLock.RUnlock()
  if portStatus := portStatusMap[util.GetListenerPort(r)]; portStatus != nil {
    if status, present := util.GetIntParam(r, "status"); present {
      util.AddLogMessage(fmt.Sprintf("Status: %d, Request count: %d, Response count: %d",
        status, portStatus.countsByRequestedStatus[status], portStatus.countsByReportedStatus[status]), r)
      w.WriteHeader(http.StatusOK)
      fmt.Fprintf(w, "{\"status\": %d, \"requestCount\": %d, \"responseCount\": %d}\n",
        status, portStatus.countsByRequestedStatus[status], portStatus.countsByReportedStatus[status])
    } else {
      util.AddLogMessage("Reporting count for all statuses", r)
      countsByRequestedStatus := util.ToJSON(portStatus.countsByRequestedStatus)
      countsByReportedStatus := util.ToJSON(portStatus.countsByReportedStatus)
      msg := fmt.Sprintf("{\"countsByRequestedStatus\": %s, \"countsByReportedStatus\": %s}",
        countsByRequestedStatus, countsByReportedStatus)
      util.AddLogMessage(msg, r)
      w.WriteHeader(http.StatusOK)
      fmt.Fprintln(w, msg)
    }
  } else {
    w.WriteHeader(http.StatusNoContent)
    fmt.Fprintln(w, "No data")
  }
}

func clearStatusCounts(w http.ResponseWriter, r *http.Request) {
  statusLock.Lock()
  defer statusLock.Unlock()
  portStatus := getOrCreatePortStatus(r)
  portStatus.countsByRequestedStatus = map[int]int{}
  portStatus.countsByReportedStatus = map[int]int{}
  util.AddLogMessage("Clearing status counts", r)
  w.WriteHeader(http.StatusAccepted)
  fmt.Fprintln(w, "Status counts cleared")
}

func Middleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if !util.IsAdminRequest(r) {
      crw := intercept.NewInterceptResponseWriter(w, true)
      next.ServeHTTP(crw, r)
      statusLock.Lock()
      defer statusLock.Unlock()
      portStatus := getOrCreatePortStatus(r)
      crw.StatusCode = computeResponseStatus(crw.StatusCode, portStatus)
      portStatus.countsByReportedStatus[crw.StatusCode]++
      util.AddLogMessage(fmt.Sprintf("Reporting status: [%d]", crw.StatusCode), r)
      crw.Proceed()
    } else {
      next.ServeHTTP(w, r)
    }
  })
}