#
# goto

## What is it?
An HTTP server+client testing tool in one. 

## Why?
It's hard to find some of these features together in a single tool

## What Features?
See below

## How to use it?
It's available as a docker image: https://hub.docker.com/repository/docker/uk0000/goto.
Or build it locally on your machine
```
go build -o goto .
```

<br/>

---
# Scenarios

Before we look into detailed features and APIs exposed by the tool, let's look at how this tool can be used in a few scenarios to understand it better.

### Scenario: [Use HTTP client to send requests and track results](scenarios.md#scenario-use-http-client-to-send-requests-and-track-results)

### Scenario: [Run dynamic traffic from K8s pods at startup](scenarios.md#scenario-run-dynamic-traffic-from-k8s-pods-at-startup)

### Scenario: [Capture results from transient pods](scenarios.md#scenario-capture-results-from-transient-pods)

### Scenario: [HTTPS traffic with certificate validation](scenarios.md#scenario-https-traffic-with-certificate-validation)

### Scenario: [Test a client's behavior upon service failure](scenarios.md#scenario-test-a-clients-behavior-upon-service-failure)

### Scenario: [Count number of requests received at each service instance (Pod/VM) for certain headers](scenarios.md#scenario-count-number-of-requests-received-at-each-service-instance-podvm-for-certain-headers)

### Scenario: [Track Request/Connection Timeouts](scenarios.md#scenario-track-requestconnection-timeouts)

<br/>

  <span style="color:red">
  TODO: There are many more possible scenarios to describe here, to show how this tool can be used for various kinds of chaos testing and investigations.
  </span>

<br/>

#
# Features

It's an HTTP client and server built into a single application. 

As a server, it can act as an HTTP proxy that lets you intercept HTTP requests and get some insights (e.g. based on headers) before forwarding it to its destination. But it can also respond to requests as a server all by itself, while still capturing interesting stats and counters that can be used to correlate information against the client.

As a client, it allows sending requests to various destinations and tracking responses by headers and response status code.

The application exposes both client and server features via various management REST APIs as described below. Additionally, it can respond to all undefined URIs with a configurable status code.

First things first, run the application:
```
go run main.go --port 8080
```
Or, build and run
```
go build -o goto .
./goto
```

<br/>


#
# Startup Command Arguments
The application accepts the following command arguments:

<table>
    <thead>
        <tr>
            <th>Argument</th>
            <th>Description</th>
            <th>Default Value</th>
        </tr>
    </thead>
    <tbody>
        <tr>
          <td rowspan="3">--port</td>
          <td>Initial port the server listens on. </td>
          <td rowspan="3">8080</td>
        </tr>
        <tr>
          <td>* Additional ports can be opened by making listener API calls on this port.</td>
        </tr>
        <tr>
          <td>* See [Listeners](#-listeners) feature later in the doc.</td>
        </tr>
        <tr>
          <td rowspan="2">--label</td>
          <td>Label this server instance will use to identify itself. </td>
          <td rowspan="2">Goto-`IPAddress` </td>
        </tr>
        <tr>
          <td>* This is used both for setting `Goto`'s default response headers as well as when registering with registry.</td>
        </tr>
        <tr>
          <td rowspan="3">--registry</td>
          <td>URL of the Goto Registry instance that this instance should connect to. </td>
          <td rowspan="3"> "" </td>
        </tr>
        <tr>
          <td>* This is used to getting initial configs and optionally report results to registry.</td>
        </tr>
        <tr>
          <td>* See [Registry](#registry-features) feature later in the doc.</td>
        </tr>
        <tr>
          <td rowspan="2">--locker</td>
          <td> Whether this instance should report its results back to the Goto Registry instance. </td>
          <td rowspan="2"> false </td>
        </tr>
        <tr>
          <td>* An instance can be asked to report its results to registry in case the  instance is transient, e.g. pods.</td>
        </tr>
        <tr>
          <td rowspan="2">--certs</td>
          <td> Directory path from where to load TLS root certificates. </td>
          <td rowspan="2"> "/etc/certs" </td>
        </tr>
        <tr>
          <td>* The loaded root certificates are used if available, otherwise system default root certs are used.</td>
        </tr>
    </tbody>
</table>


Once the server is up and running, rest of the interactions and configurations are done purely via REST APIs.

<br/>


#
# Client Features
As a client tool, `goto` offers the following features:
- Configure targets to be invoked ahead of time before invocation
- Invoke selective targets or all configured targets in batches
- Multiple concurrent invocations of batches of targets
- Control the number of concurrent requests per target via `replicas` field
- Control the total number of requests per target via `requestCount` field
- Control the minimum wait time after each replica set invocation per target via `delay` field
- Control the minimum duration over which total requests are sent to a client using combination of `requestCount` and `delay`
- 
allows targets to be configured and invoked via REST APIs. Headers can be set to track results for target invocations, and APIs make those results available for consumption as JSON output. The invocation results get accumulated across multiple invocations until cleared explicitly. In addition to keeping the results in the `goto` client instance, those are also stored in locker on registry instance if enabled. (See `--locker` command arg)


#### APIs
|METHOD|URI|Description|
|---|---|---|
| POST      | /client/targets/add                   | Add a target for invocation. [See `Client Target JSON Schema` for Payload](#client-target-json-schema) |
| POST      |	/client/targets/{targets}/remove      | Remove given targets |
| POST      | /client/targets/{targets}/invoke      | Invoke given targets |
| POST      |	/client/targets/invoke/all            | Invoke all targets |
| POST      | /client/targets/{targets}/stop        | Stops a running target |
| POST      | /client/targets/stop/all              | Stops all running targets |
| GET       |	/client/targets/list                  | Get list of currently configured targets |
| GET       |	/client/targets                       | Get list of currently configured targets |
| POST      |	/client/targets/clear                 | Remove all targets |
| PUT, POST |	/client/blocking/set/{flag}           | Set whether calls to invoke will block and receive full target responses  |
| GET       |	/client/blocking                      | Get current state of the blocking flag |
| PUT, POST |	/client/track/headers/add/{headers}   | Add headers for tracking response counts per target |
| PUT, POST |	/client/track/headers/remove/{headers}| Remove headers from tracking set |
| POST      | /client/track/headers/clear           | Remove all tracked headers |
| GET       |	/client/track/headers/list            | Get list of tracked headers |
| GET       |	/client/track/headers                 | Get list of tracked headers |
| GET       |	/client/results                       | Get invocation results in JSON format. See [`Results Schema`](#client-results-schema) |
| POST      | /client/results/{targets}/clear       | Clear previously accumulated invocation results for specific targets |
| POST      | /client/results/clear                 | Clear previously accumulated invocation results |


#### Client Target JSON Schema
|Field|Data Type|Description|
|---|---|---|
| name         | string         | Name for this target |
| method       | string         | HTTP method to use for this target |
| url          | string         | URL for this target   |
| verifyTLS    | bool           | Whether the TLS certificate presented by the target is verified. (Also see `--certs` command arg) |
| headers      | [][]string     | Headers to be sent to this target |
| body         | string         | Request body to use for this target|
| replicas     | int            | Number of parallel invocations to be done for this target |
| requestCount | int            | Number of requests to be made per replicas for this target. The final request count becomes replicas * requestCount  |
| delay        | duration       | Minimum delay to be added per request. The actual added delay will be the max of all the targets being invoked in a given round of invocation, but guaranteed to be greater than this delay |
| sendID       | bool           | Whether or not a unique ID be sent with each client request. If this flag is set, a query param `x-request-id` will be added to each request, which can help with tracing requests on the target servers |
| autoInvoke   | bool           | Whether this target should be invoked as soon as it's added |


#### Client Results Schema
|Field|Data Type|Description|
|---|---|---|
| targetInvocationCounts      | string->int                 | Total requests sent per target |
| targetFirstResponses        | string->time                | Time of first response received from the target |
| targetLastResponses         | string->time                | Time of last response received from the target |
| countsByStatus              | string->int                 | Response counts across all targets grouped by HTTP Status |
| countsByStatusCodes         | string->int                 | Response counts across all targets grouped by HTTP Status Code |
| countsByHeaders             | string->int                 | Response counts across all targets grouped by header names   |
| countsByHeaderValues        | string->string->int         | Response counts across all targets grouped by header names and values |
| countsByTargetStatus        | string->string->int         | Response counts per target grouped by HTTP Status |
| countsByTargetStatusCode    | string->string->int         | Response counts per target grouped by HTTP Status Code |
| countsByTargetHeaders       | string->string->int         | Response counts per target grouped by header names |
| countsByTargetHeaderValues  | string->string->string->int | Response counts per target grouped by header names and header values |


#### Client API and Results Examples

```
#Add target
curl localhost:8080/client/targets/add --data '
{
  "name": "t1",
  "method":	"POST",
  "url": "http://somewhere:8080/foo",
  "headers":[["x", "x1"],["y", "y1"]],
  "body": "{\"test\":\"this\"}",
  "replicas": 2, 
  "requestCount": 2, 
  "delay": "200ms", 
  "sendID": true,
  "autoInvoke": true
}'

#List targets
curl localhost:8080/client/targets

#Remove select target
curl -X POST localhost:8080/client/target/t1,t2/remove

#Clear all configured targets
curl -X POST localhost:8080/client/targets/clear

#Invoke select targets
curl -X POST localhost:8080/client/targets/t2,t3/invoke

#Invoke all targets
curl -X POST localhost:8080/client/targets/invoke/all

#Stop select targets across all running batches
curl -X POST localhost:8080/client/targets/t2,t3/stop

#Stop all targets across all running batches
curl -X POST localhost:8080/client/targets/stop/all

#Set blocking mode
curl -X POST localhost:8080/client/blocking/set/n

#Get blocking mode
curl localhost:8080/client/blocking

#Clear tracked headers
curl -X POST localhost:8080/client/track/headers/clear

#Add headers to track
curl -X PUT localhost:8080/client/track/headers/add/Goto-Host,Via-Goto,x,y,z,foo

#Remove headers from tracking
curl -X PUT localhost:8080/client/track/headers/remove/foo

#Get list of tracked headers
curl localhost:8080/client/track/headers/list

#Clear results
curl -X POST localhost:8080/client/results/clear

#Get results
curl localhost:8080/client/results
```

<details>
<summary>Sample Client Invocation Result (including error reporting example)</summary>
<p>

```json

{
  "targetInvocationCounts": {
    "target1": 40,
    "target2": 20,
    "target3": 600,
    "target4": 300
  },
  "targetFirstResponses": {
    "target1": "2020-06-09T17:54:42.966245-07:00",
    "target2": "2020-06-09T17:54:42.966888-07:00",
    "target3": "2020-06-09T17:54:45.027159-07:00",
    "target4": "2020-06-09T17:54:51.330125-07:00"
  },
  "targetLastResponses": {
    "target1": "2020-06-09T17:54:47.016191-07:00",
    "target2": "2020-06-09T17:54:47.014026-07:00",
    "target3": "2020-06-09T17:54:52.113565-07:00",
    "target4": "2020-06-09T17:54:52.113164-07:00"
  },
  "countsByStatus": {
    "200 OK": 880,
    "403 Forbidden": 40,
    "502 Bad Gateway": 20,
    "503 Service Unavailable": 20
  },
  "countsByStatusCodes": {
    "200": 880,
    "403": 40,
    "502": 20,
    "503": 20
  },
  "countsByHeaders": {
    "foo": 900,
    "goto-host": 960,
    "via-goto": 960,
    "x": 660,
    "y": 960
  },
  "countsByHeaderValues": {
    "foo": {
      "bar1": 600,
      "bar2": 300
    },
    "goto-host": {
      "1.1.1.1": 40
      "2.2.2.2": 20
      "3.3.3.3": 960
    },
    "via-goto": {
      "Server8081": 40,
      "Server8082": 20,
      "Server8083": 900
    },
    "x": {
      "x1": 640,
      "x2": 20
    },
    "y": {
      "x2": 300,
      "y1": 640,
      "y2": 20
    }
  },
  "countsByTargetStatus": {
    "target1": {
      "200 OK": 20,
      "502 Bad Gateway": 20
    },
    "target2": {
      "503 Service Unavailable": 20
    },
    "target3": {
      "200 OK": 570,
      "403 Forbidden": 30
    },
    "target4": {
      "200 OK": 290,
      "403 Forbidden": 10
    }
  },
  "countsByTargetStatusCode": {
    "target1": {
      "200": 20,
      "502": 20
    },
    "target2": {
      "503": 20
    },
    "target3": {
      "200": 570,
      "403": 30
    },
    "target4": {
      "200": 290,
      "403": 10
    }
  },
  "countsByTargetHeaders": {
    "target1": {
      "goto-host": 40,
      "via-goto": 40,
      "x": 40,
      "y": 40
    },
    "target2": {
      "goto-host": 20,
      "via-goto": 20,
      "x": 20,
      "y": 20
    },
    "target3": {
      "foo": 600,
      "goto-host": 600,
      "via-goto": 600,
      "x": 600,
      "y": 600
    },
    "target4": {
      "foo": 300,
      "goto-host": 300,
      "via-goto": 300,
      "y": 300
    }
  },
  "countsByTargetHeaderValues": {
    "target1": {
      "goto-host": {
        "1.1.1.1": 40
      },
      "via-goto": {
        "Server8081": 40
      },
      "x": {
        "x1": 40
      },
      "y": {
        "y1": 40
      }
    },
    "target2": {
      "goto-host": {
        "2.2.2.2": 20
      },
      "via-goto": {
        "Server8082": 20
      },
      "x": {
        "x2": 20
      },
      "y": {
        "y2": 20
      }
    },
    "target3": {
      "foo": {
        "bar1": 600
      },
      "goto-host": {
        "3.3.3.3": 600
      },
      "via-goto": {
        "Server8083": 600
      },
      "x": {
        "x1": 600
      },
      "y": {
        "y1": 600
      }
    },
    "target4": {
      "foo": {
        "bar2": 300
      },
      "goto-host": {
        "3.3.3.3": 300
      },
      "via-goto": {
        "Server8083": 300
      },
      "y": {
        "x2": 300
      }
    }
  }
}

```
</p>
</details>


<br/>

#
# Server Features
The server is useful to be run as a test server for testing some client application, proxy/sidecar, gateway, etc. Or, the server can also be used as a proxy to be put in between a client and a target server application, so that traffic flows through this server where headers can be inspected/tracked before proxying the requests further. The server can add headers, replace request URI with some other URI, add artificial delays to the response, respond with a specific status, monitor request/connection timeouts, etc. The server tracks all the configured parameters, applying those to runtime traffic and building metrics, which can be viewed via various APIs.

<br/>

#
## > Listeners


The server starts with a single http listener on port given to it as command line arg (defaults to 8080). It exposes listener APIs to let you manage additional HTTP listeners (TCP support will come in the future). The ability to launch and shutdown listeners lets you do some chaos testing. All listener ports respond to the same set of API calls, so any of the APIs described below as well as runtime traffic proxying can be done via any active listener.


#### APIs
|METHOD|URI|Description|
|---|---|---|
| POST       | /listeners/add           | Add a listener. [See Payload JSON Schema](#listener-json-schema)|
| POST, PUT  | /listeners/{port}/remove | Remove a listener|
| POST, PUT  | /listeners/{port}/open   | Open an added listener to accept traffic|
| POST, PUT  | /listeners/{port}/close  | Close an added listener|
| GET        | /listeners               | Get a list of listeners |

#### Listener JSON Schema
|Field|Data Type|Description|
|---|---|---|
|label    |string | Label to be applied to the listener. This can also be set/changed via REST API later|
|port     |int    | Port on which the new listener will listen on|
|protocol |string | Currently only `http`. TCP support will come soon.|


#### Listener API Examples:
```
curl localhost:8080/listeners/add --data '{"port":8081, "protocol":"http", "label":"Server-8081"}'

curl -X POST localhost:8080/listeners/8081/remove

curl -X PUT localhost:8080/listeners/8081/open

curl -X PUT localhost:8080/listeners/8081/close

curl localhost:8081/listeners
```

<br/>

#
## > Listener Label

By default, each listener adds a header `Via-Goto: <port>` to each response it sends, where <port> is the port on which the listener is running (default being 8080). A custom label can be added to a listener using the label APIs described below. In addition to `Via-Goto`, each listener also adds another header `Goto-Host` that carries the pod/host name, pod namespace (or `local` if not running as a K8s pod), and pod/host IP address to identify where the response came from.

#### APIs
|METHOD|URI|Description|
|---|---|---|
| POST, PUT | /label/set/{label}  | Set label for this port |
| PUT       | /label/clear        | Remove label for this port |
| GET       | /label              | Get current label of this port |

#### Listener Label API Examples:
```
curl -X PUT localhost:8080/label/set/Server-8080

curl -X PUT localhost:8080/label/clear

curl localhost:8080/label
```

<br/>

#
## > Request Headers Tracking
This feature allows tracking request counts by headers.

#### APIs
|METHOD|URI|Description|
|---|---|---|
|POST     | /request/headers/track/clear									| Remove all tracked headers |
|PUT, POST| /request/headers/track/add/{headers}					| Add headers to track |
|PUT, POST|	/request/headers/track/{headers}/remove				| Remove given headers from tracking |
|GET      | /request/headers/track/{header}/counts				| Get counts for a tracked header |
|PUT, POST| /request/headers/track/counts/clear/{headers}	| Clear counts for given tracked headers |
|POST     | /request/headers/track/counts/clear						| Clear counts for all tracked headers |
|GET      | /request/headers/track/counts									| Get counts for all tracked headers |
|GET      | /request/headers/track/list									  | Get list of tracked headers |
|GET      | /request/headers/track									      | Get list of tracked headers |

#### Request Headers Tracking API Examples:
```
curl -X POST localhost:8080/request/headers/track/clear

curl -X PUT localhost:8080/request/headers/track/add/x,y

curl -X PUT localhost:8080/request/headers/track/remove/x

curl -X POST localhost:8080/request/headers/track/counts/clear/x

curl -X POST localhost:8080/request/headers/track/counts/clear

curl -X POST localhost:8080/request/headers/track/counts/clear

curl localhost:8080/request/headers/track/list
```

#### Request Header Tracking Results Example
```
$ curl localhost:8080/request/headers/track/counts

{
  "x": {
    "RequestCountsByHeaderValue": {
      "x1": 20
    },
    "RequestCountsByHeaderValueAndRequestedStatus": {
      "x1": {
        "418": 20
      }
    },
    "RequestCountsByHeaderValueAndResponseStatus": {
      "x1": {
        "418": 20
      }
    }
  },
  "y": {
    "RequestCountsByHeaderValue": {
      "y1": 20
    },
    "RequestCountsByHeaderValueAndRequestedStatus": {
      "y1": {
        "418": 20
      }
    },
    "RequestCountsByHeaderValueAndResponseStatus": {
      "y1": {
        "418": 20
      }
    }
  }
}
```

<br/>

#
## > Request Timeout
This feature allows tracking request timeouts by headers.

#### APIs
|METHOD|URI|Description|
|---|---|---|
|PUT, POST| /request/timeout/track/headers/{headers}  | Add one or more headers. Requests carrying these headers will be tracked for timeouts and reported |
|PUT, POST| /request/timeout/track/all                | Enable request timeout tracking for all requests |
|POST     |	/request/timeout/track/clear              | Clear timeout tracking configs |
|POST     |	/request/timeout/status                   | Get a report of tracked request timeouts so far |


#### Request Timeout API Examples
```
curl -X POST localhost:8080/request/timeout/track/headers/x,y

curl -X POST localhost:8080/request/timeout/track/headers/all

curl -X POST localhost:8080/request/timeout/track/clear

curl localhost:8080/request/timeout/status
```

<br/>

#
## > URIs
This feature allows tracking request counts by URIs

#### APIs
|METHOD|URI|Description|
|---|---|---|
|GET      |	/request/uri/counts                     | Get request counts for all URIs |
|POST     |	/request/uri/counts/clear               | Clear request counts for all URIs |


#### URI API Examples
```
curl -X POST localhost:8080/request/uri/counts

curl -X POST localhost:8080/request/uri/counts/clear
```

<br/>

#
## > URIs Bypass
This feature allows adding bypass URIs that will not be subject to other configurations, e.g. forced status codes. Request counts are tracked for bypass URIs, and specific status can be configured to respond for bypass URI requests.

#### APIs
|METHOD|URI|Description|
|---|---|---|
|PUT, POST| /request/uri/bypass/add?uri={uri}       | Add a bypass URI |
|PUT, POST| /request/uri/bypass/remove?uri={uri}    | Remove a bypass URI |
|PUT, POST| /request/uri/bypass/clear               | Remove all bypass URIs |
|PUT, POST| /request/uri/bypass/status/set/{status} | Set status code to be returned for bypass URI requests |
|GET      |	/request/uri/bypass/list                | Get list of bypass URIs |
|GET      |	/request/uri/bypass                     | Get list of bypass URIs |
|GET      |	/request/uri/bypass/status              | Get current bypass URI status code |
|GET      |	/request/uri/bypass/counts?uri={uri}    | Get request counts for a given bypass URI |


#### URI Bypass API Examples
```
curl -X POST localhost:8080/request/uri/bypass/clear

curl -X PUT localhost:8080/request/uri/bypass/add\?uri=/foo

curl -X PUT localhost:8081/request/uri/bypass/remove\?uri=/bar

curl -X PUT localhost:8080/request/uri/bypass/status/set/418

curl localhost:8081/request/uri/bypass/list

curl localhost:8080/request/uri/bypass

curl localhost:8080/request/uri/bypass/status

curl localhost:8080/request/uri/bypass/counts\?uri=/foo
```


<br/>

#
## > Response Delay
This feature allows adding a delay to all requests except bypass URIs and proxy requests. Delay is specified as duration, e.g. 1s

#### APIs
|METHOD|URI|Description|
|---|---|---|
| PUT, POST | /response/delay/set/{delay} | Set a delay for non-management requests (i.e. runtime traffic) |
| PUT, POST | /response/delay/clear       | Remove currently set delay |
| GET       |	/response/delay             | Get currently set delay |

#### Response Delay API Examples

```
curl -X POST localhost:8080/response/delay/clear

curl -X PUT localhost:8080/response/delay/set/2s

curl localhost:8080/response/delay
```

<br/>

#
## > Response Headers
This feature allows adding custom response headers to all responses sent by the server.

#### APIs
|METHOD|URI|Description|
|---|---|---|
| PUT, POST | /response/headers/add/{header}/{value}  | Add a custom header to be sent with all responses |
| PUT, POST | /response/headers/remove/{header}       | Remove a previously added custom response header |
| POST      |	/response/headers/clear                 | Remove all configured custom response headers |
| GET       |	/response/headers/list                  | Get list of configured custom response headers |
| GET       |	/response/headers                       | Get list of configured custom response headers |

#### Response Headers API Examples
```
curl -X POST localhost:8080/response/headers/clear

curl -X POST localhost:8080/response/headers/add/x/x1

curl localhost:8080/response/headers/list

curl -X POST localhost:8080/response/headers/remove/x

curl localhost:8080/response/headers
```


<br/>

#
## > Response Payload
This feature allows setting custom response payload to be sent with server responses. Response payload can be set for all requests (default), for specific URIs, or for specific headers. If response is set for all three, URI response payload gets highest priority if matched with request URI, followed by payload for matching request headers, and otherwise default payload is used as fallback if configured. If no custom payload is configured, the request continues with its normal processing, in which case it may receive the "catch all" echo response.

#### APIs
|METHOD|URI|Description|
|---|---|---|
| POST | /response/payload/set/default  | Add a custom payload to be sent with all resopnses |
| POST | /response/payload/set/uri?uri={uri}  | Add a custom payload to be sent for requests matching the given URI. URI can contain placeholders |
| POST | /response/payload/set/header/{header}  | Add a custom payload to be sent for requests matching the given header name |
| POST | /response/payload/set/header/{header}/value/{value}  | Add a custom payload to be sent for requests matching the given header name and value |
| POST | /response/payload/clear  | Clear all configured custom response payloads |
| GET  |	/response/payload                      | Get configured custom payloads |

#### Response Payload API Examples
```
curl -X POST localhost:8080/response/payload/set/default --data '{"test": "default payload"}'

curl -X POST localhost:8080/response/payload/set/uri?uri=/foo/{f}/bar{b} --data '{"test": "uri was /foo/{}/bar/{}"}'

curl -X POST localhost:8080/response/payload/set/header/foo --data '{"test": "header was foo"}'

curl -X POST localhost:8080/response/payload/set/header/foo/value/bar --data '{"test": "header was foo with value bar"}'

curl -X POST localhost:8080/response/payload/clear

curl localhost:8080/response/payload
```

<br/>

#
## > Response Status
This feature allows setting a forced response status for all requests except bypass URIs. Server also tracks number of status requests received (via /status URI) and number of responses send per status code.

#### APIs
|METHOD|URI|Description|
|---|---|---|
| PUT, POST | /response/status/set/{status}     | Set a forced response status that all non-proxied and non-management requests will be responded with |
| PUT, POST |	/response/status/clear            | Remove currently configured forced response status, so that all subsequent calls will receive their original deemed response |
| PUT, POST | /response/status/counts/clear     | Clear counts tracked for response statuses |
| GET       |	/response/status/counts/{status}  | Get request counts for a given status |
| GET       |	/response/status/counts           | Get request counts for all response statuses so far |
| GET       |	/response/status                  | Get the currently configured forced response status |

#### Response Status API Examples
```
curl -X POST localhost:8080/response/status/counts/clear

curl -X POST localhost:8080/response/status/clear

curl -X PUT localhost:8080/response/status/set/502

curl -X PUT localhost:8080/response/status/set/0

curl -X POST localhost:8080/response/status/counts/clear

curl localhost:8080/response/status/counts

curl localhost:8080/response/status/counts/502
```

#### Response Status Tracking Result Example
```
{
  "countsByRequestedStatus": {
    "418": 20
  },
  "countsByReportedStatus": {
    "200": 15,
    "202": 4,
    "208": 5,
    "418": 20
  }
}
```

<br/>

#
## > Status API
This URI allows client to ask for a specific status as response code. The given status is reported back, except when forced status is configured in which case the forced status is sent as response.

#### APIs
|METHOD|URI|Description|
|---|---|---|
| GET       |	/status/{status}                  | This call either receives the given status, or the forced response status if one is set |

#### Status API Examples
```
curl -I  localhost:8080/status/418
```

<br/>

#
## > Echo API
This URI echoes back the headers and payload sent by client. The response is also subject to any forced response status and will carry custom headers if any are configured.

#### APIs
|METHOD|URI|Description|
|---|---|---|
| GET       |	/echo                  | Sends response back with request headers and body, with added custom response headers and forced status |

#### Echo API Examples
```
curl -I  localhost:8080/echo
```

<br/>

#
## > CatchAll

Any request that doesn't match any of the defined management APIs, and also doesn't match any proxy targets, gets treated by a catch-all response that sends HTTP 200 response by default (unless an override response code is set)


<br/>
<br/>

#
# Proxy Features

`Goto` proxy feature allows targets to be configured that are triggered based on matching criteria against requests. The targets can also be invoked manually for testing the configuration. However, the real fun happens when the proxy targets are matched with runtime traffic based on the match criteria specified in a proxy target's spec (based on headers, URIs, and query parameters), and one or more matching targets get invoked for a given request.

#### APIs
|METHOD|URI|Description|
|---|---|---|
|POST     |	/request/proxy/targets/add              | Add target for proxying requests [see `Proxy Target JSON Schema`](#proxy-target-json-schema) |
|PUT, POST| /request/proxy/targets/{target}/remove  | Remove a proxy target |
|PUT, POST| /request/proxy/targets/{target}/enable  | Enable a proxy target |
|PUT, POST| /request/proxy/targets/{target}/disable | Disable a proxy target |
|POST     |	/request/proxy/targets/{targets}/invoke | Invoke proxy targets by name |
|POST     |	/request/proxy/targets/invoke/{targets} | Invoke proxy targets by name |
|GET      |	/request/proxy/targets/counts           | Get proxy target match/invocation stats, by uri, header and query params |
|POST     |	/request/proxy/targets/counts/clear     | Remove all proxy target match/invocation stats |
|POST     |	/request/proxy/targets/clear            | Remove all proxy targets |
|GET 	    |	/request/proxy/targets                  | List all proxy targets |


#### Proxy Target JSON Schema
|Field|Data Type|Description|
|---|---|---|
| name          | string                                | Name for this target |
| url           | string                                | URL for the target. Request's URI or Override URI gets added to the URL for each proxied request. |
| sendID        | bool           | Whether or not a unique ID be sent with each request. If this flag is set, a query param `x-request-id` will be added to each request, which can help with tracing requests on the target servers |
| replaceURI    | string                                | URI to be used in place of the original request URI.|
| addHeaders    | `[][]string`                            | Additional headers to add to the request before proxying |
| removeHeaders | `[]string `                             | Headers to remove from the original request before proxying |
| addQuery      | `[][]string`                            | Additional query parameters to add to the request before proxying |
| removeQuery   | `[]string`                              | Query parameters to remove from the original request before proxying |
| match        | JSON     | Match criteria based on which runtime traffic gets proxied to this target. See [JSON Schema](#proxy-target-match-criteria-json-schema) and [detailed explanation](#proxy-target-match-criteria) below |
| replicas     | int      | Number of parallel replicated calls to be made to this target for each matched request. This allows each request to result in multiple calls to be made to a target if needed for some test scenarios |
| enabled       | bool     | Whether or not the proxy target is currently active |

#### Proxy Target Match Criteria JSON Schema
|Field|Data Type|Description|
|---|---|---|
| headers | `[][]string`  | Headers names and optional values to match against request headers |
| uris    | `[]string`    | URIs with optional {placeholders} to match against request URI |
| query   | `[][]string`  | Query parameters with optional values to match against request query |


#### Proxy Target Match Criteria
Proxy target match criteria specify the URIs, headers and query parameters, matching either of which will cause the request to be proxied to the target.

- URIs: specified as a list of URIs, with `{foo}` to be used for variable portion of a URI. E.g., `/foo/{f}/bar/{b}` will match URIs like `/foo/123/bar/abc`, `/foo/something/bar/otherthing`, etc. The variables are captured under the given labels (f and b in previous example). If the target is configured with `replaceURI` to proxy the request to a different URI than the original request, the `replaceURI` can refer to those capturing variables using the syntax described in this example:
  
  ```
  curl http://goto:8080/request/proxy/targets/add --data \
  '{"name": "target1", "url":"http://somewhere", \
  "match":{"uris":["/foo/{x}/bar/{y}"]}, \
  "replaceURI":"/abc/{y:.*}/def/{x:.*}", \
  "enabled":true, "sendID": true}'
  ```
  
  This target will be triggered for requests with the pattern `/foo/<somex>/bar/<somey>` and the request will be forwarded to the target as `http://somewhere/abc/somey/def/somex`, where the values `somex` and `somey` are extracted from the original request and injected into the replacement URI.

  URI match `/` has the special behavior of matching all traffic.

<br/>

- Headers: specified as a list of key-value pairs, with the ability to capture values in named variables and reference those variables in the `addHeaders` list. A target is triggered if any of the headers in the match list are present in the request (headers are matched using OR instead of AND). The variable to capture header value is specified as `{foo}`, and can be referenced in the `addHeaders` list again as `{foo}`. This example will make it clear:

  ```
  curl http://goto:8080/request/proxy/targets/add --data \
  '{"name": "target2", "url":"http://somewhere", \
  "match":{"headers":[["foo", "{x}"], ["bar", "{y}"]]}, \
  "addHeaders":[["abc","{x}"], ["def","{y}"]], "removeHeaders":["foo"], \
  "enabled":true, "sendID": true}'
  ```

  This target will be triggered for requests carrying headers `foo` or `bar`. On the proxied request, additional headers will be set: `abc` with value copied from `foo`, an `def` with value copied from `bar`. Also, header `foo` will be removed from the proxied request.

<br/>

- Query: specified as a list of key-value pairs, with the ability to capture values in named variables and reference those variables in the `addQuery` list. A target is triggered if any of the query parameters in the match list are present in the request (matched using OR instead of AND). The variable to capture query parameter value is specified as `{foo}`, and can be referenced in the `addQuery` list again as `{foo}`. Example:

    ```
  curl http://goto:8080/request/proxy/targets/add --data \
  '{"name": "target3", "url":"http://somewhere", \
  "match":{"query":[["foo", "{x}"], ["bar", "{y}"]]}, \
  "addQuery":[["abc","{x}"], ["def","{y}"]], "removeQuery":["foo"], \
  "enabled":true, "sendID": true}'
  ```

  This target will be triggered for requests with carrying query params `foo` or `bar`. On the proxied request, query param `foo` will be removed, and additional query params will be set: `abc` with value copied from `foo`, an `def` with value copied from `bar`. For incoming request `http://goto:8080?foo=123&bar=456` gets proxied as `http://somewhere?abc=123&def=456&bar=456`. 

<br/>

#### Request Proxying API Examples:
```
curl -X POST localhost:8080/request/proxy/targets/clear

curl localhost:8081/request/proxy/targets/add --data '{"name": "t1", \
"match":{"uris":["/x/{x}/y/{y}"], "query":[["foo", "{f}"]]}, \
"url":"http://localhost:8083", \
"replaceURI":"/abc/{y:.*}/def/{x:.*}", \
"addHeaders":[["z","z1"]], \
"addQuery":[["bar","{f}"]], \
"removeQuery":["foo"], \
"replicas":1, "enabled":true, "sendID": true}'

curl localhost:8081/request/proxy/targets/add --data '{"name": "t2", \
"match":{"headers":[["foo"]]}, \
"url":"http://localhost:8083", \
"replaceURI":"/echo", \
"addHeaders":[["z","z2"]], \
"replicas":1, "enabled":true, "sendID": false}'

curl localhost:8082/request/proxy/targets/add --data '{"name": "t3", \
"match":{"headers":[["x", "{x}"], ["y", "{y}"]], "uris":["/foo"]}, \
"url":"http://localhost:8083", \
"replaceURI":"/echo", \
"addHeaders":[["z","{x}"], ["z","{y}"]], \
"removeHeaders":["x", "y"], \
"replicas":1, "enabled":true, "sendID": true}'

curl -X PUT localhost:8080/request/proxy/targets/t1/remove

curl -X PUT localhost:8080/request/proxy/targets/t2/disable

curl -X PUT localhost:8080/request/proxy/targets/t2/enable

curl -v -X POST localhost:8080/request/proxy/targets/t1/invoke

curl localhost:8080/request/proxy/targets
```

<br/>



#
# Trigger Features

`Goto` allow targets to be configured that are triggered based on response status. The triggers can be invoked manually for testing, but their real value is when they get triggered based on response status. Even more valuable when the request was proxied to another upstream service, in which case the trigger is based on the response status of the upstream service.

#### APIs
|METHOD|URI|Description|
|---|---|---|
|POST     |	/response/trigger/add              | Add a trigger target. See [Trigger Target JSON Schema](#trigger-target-json-schema) |
|PUT, POST| /response/trigger/{target}/remove  | Remove a trigger target |
|PUT, POST| /response/trigger/{target}/enable  | Enable a trigger target |
|PUT, POST| /response/trigger/{target}/disable | Disable a trigger target |
|POST     |	/response/trigger/{targets}/invoke | Invoke trigger targets by name for manual testing |
|POST     |	/response/trigger/clear            | Remove all trigger targets |
|GET 	    |	/response/trigger/list             | List all trigger targets |


#### Trigger Target JSON Schema
|Field|Data Type|Description|
|---|---|---|
| name          | string                                | Name for this target |
| method        | string                                | HTTP method to use for this target |
| url           | string                                | URL for the target. |
| headers       | `[][]string`                          | request headers to send with this trigger request |
| body          | `string`                              | request body to send with this trigger request |
| sendID        | bool           | Whether or not a unique ID be sent with each request. If this flag is set, a query param `x-request-id` will be added to each request, which can help with tracing requests on the target servers |
| enabled       | bool     | Whether or not the trigger is currently active |
| triggerOnResponseStatuses | []int     | List of response statuses for which this target will be triggered |


<br/>

#### Trigger API Examples:
```
curl -X POST localhost:8080/response/trigger/clear

curl localhost:8080/response/trigger/add --data '{
	"name": "t1", 
	"method":"POST", 
	"url":"http://localhost:8082/response/status/clear", 
	"headers":[["foo", "bar"],["x", "x1"],["y", "y1"]], 
	"body": "{\"test\":\"this\"}", 
	"sendId": true, 
	"enabled": true, 
	"triggerOnResponseStatuses": [502, 503]
}'

curl -X POST localhost:8080/response/trigger/t1/remove

curl -X POST localhost:8080/response/trigger/t1/enable

curl -X POST localhost:8080/response/trigger/t1/disable

curl -X POST localhost:8080/response/trigger/t1/invoke

curl localhost:8080/response/trigger/list

```

#
# Jobs Features

`Goto` allow jobs to be configured that can be run manually or auto-start upon addition. Two kinds of jobs are supported:
- HTTP requests to be made to some target URL
- Command execution on local OS
The job results can be retrieved via API from the `goto` instance, and also stored in locker on registry instance if enabled. (See `--locker` command arg)

#### Jobs APIs
|METHOD|URI|Description|
|---|---|---|
| POST  |	/jobs/add           | Add a job. See [Job JSON Schema](#job-json-schema) |
| POST  | /jobs/{jobs}/remove | Remove given jobs by name |
| POST  | /jobs/clear         | Remove all jobs |
| POST  | /jobs/{jobs}/run    | Run given jobs |
| POST  | /jobs/run/all       | Run all configured jobs |
| POST  | /jobs/{jobs}/stop   | Stop given jobs if running |
| POST  | /jobs/stop/all      | Stop all running jobs |
| GET   | /jobs/{job}/results | Get results for the given job |
| GET   | /jobs/              | Get a list of all configured jobs |


#### Job JSON Schema
|Field|Data Type|Description|
|---|---|---|
| id            | string        | ID for this job |
| task          | JSON          | Task to be executed for this job. Can be an [HTTP Task](#job-http-task-json-schema) or [Command Task](#job-command-task-json-schema) |
| delay         | duration      | Minimum delay to be added per iteration of the job. Actual delay may be higher than this. |
| count         | int           | Number of times this job should be executed during a single invocation |
| maxResults    | int           | Number of results to be retained from all the executions of this job during an invocation |
| keepFirst     | bool          | Indicates whether the first invocation result should be retained, reducing the slots for capturing remaining results by (maxResults-1) |


#### Job HTTP Task JSON Schema
|Field|Data Type|Description|
|---|---|---|
| name         | string         | Name for this target |
| method       | string         | HTTP method to use for this target |
| url          | string         | URL for this target   |
| headers      | [][]string     | Headers to be sent to this target |
| body         | string         | Request body to use for this target|
| replicas     | int            | Number of parallel invocations to be done for this target |
| requestCount | int            | Number of requests to be made per replicas for this target. The final request count becomes replicas * requestCount  |
| delay        | duration       | Minimum delay to be added per request. The actual added delay will be the max of all the targets being invoked in a given round of invocation, but guaranteed to be greater than this delay |
| sendId       | bool           | Whether or not a unique ID be sent with each client request. If this flag is set, a query param `x-request-id` will be added to each request, which can help with tracing requests on the target servers |
| parseJSON    | bool           | Indicates whether the response payload is expected to be JSON and hence not to treat it as text (to avoid escaping quotes in JSON) |


#### Job Command Task JSON Schema
|Field|Data Type|Description|
|---|---|---|
| cmd       | string         | Command to be executed on the OS. Use `sh` as command if shell features are to be used (e.g. pipe) |
| args      | []string       | Arguments to be passed to the OS command |


<br/>

#### Job APIs Examples:
```
curl -X POST http://localhost:8080/jobs/clear

curl localhost:8080/jobs/add --data '
{ 
"id": "job1",
"task": {
	"name": "job1",
	"method":	"POST",
	"url": "http://localhost:8081/echo",
	"headers":[["x", "x1"],["y", "y1"]],
	"body": "{\"test\":\"this\"}",
	"replicas": 1, "requestCount": 1, 
	"delay": "200ms",
	"parseJSON": true
	},
"auto": false,
"count": 10,
"keepFirst": true,
"maxResults": 5,
"delay": "1s"
}'

curl localhost:8080/jobs/add --data '
{ 
"id": "job2",
"task": {"cmd": "sh", "args": ["-c", "date +%s; echo Hello; sleep 1;"]},
"auto": false,
"count": 10,
"keepFirst": true,
"maxResults": 5,
"delay": "1s"
}'

curl -X POST http://localhost:8080/jobs/job1,job2/remove

curl http://localhost:8080/jobs

curl -X POST http://localhost:8080/jobs/job1,job2/run

curl -X POST http://localhost:8080/jobs/run/all

curl -X POST http://localhost:8080/jobs/job1,job2/stop

curl -X POST http://localhost:8080/jobs/stop/all

curl -X POST http://localhost:8080/jobs/job1/results
```

<br/>


#
# Registry Features

Any `goto` instance can act as a registry of other `goto` instances, and other worker `goto` instances can be configured to register themselves with the registry. You can pick any instance as registry and pass its URL to other instances as a command line argument, which tells other instances to register themselves with the given registry at startup.

A `goto` instance can be passed command line arguments '`--registry <url>`' to point it to the `goto` instance acting as a registry. When a `goto` instance receives this command line argument, it invokes the registration API on the registry instance passing its `label` and `IP:Port` to the registry server. The `label` a `goto` instance uses can also be passed to it as a command line argument '`--label <label>`'. Multiple worker `goto` instances can register using the same label but different IP addresses, which would be the case for pods of the same deployment in K8s. The worker instances that register with a registry instance at startup, also deregister themselves with the registry upon shutdown.

By registering a worker instance to a registry instance, we get a few benefits:
1. You can pre-register a list of invocation targets and jobs at the registry instance that should be handed out to the worker instances. These targets/jobs are registered by labels, and the worker instances receive the matching targets+jobs for the labels they register with.
2. The targets and jobs registered at the registry can also be marked for `auto-invocation`. When a worker instance receives a target/job from registry at startup that's marked for auto-invocation, it immediately invokes that target/job at startup. Additionally, the target/job is retained in the worker instance for later invocation via API as well.
3. In addition to sending targets/jobs to worker instances at the time of registration, the registry instance also pushes targets/jobs to the worker instances as and when more targets/jobs get added to the registry. This has the added benefit of just using the registry instance as the single point of configuration, where you add targets/jobs and those get pushed to all worker instances. Removal of targets/jobs from the registry also gets pushed, so the targets/jobs get removed from the corresponding worker instances. Even targets/jobs that are pushed later can be marked for `auto-invocation`, and the worker instances that receive the target/job will invoke it immediately upon receipt.

#### Registry APIs
|METHOD|URI|Description|
|---|---|---|
| POST      | /registry/peers/add     | Register a worker instance (referred to as peer). See [Peer JSON Schema](#peer-json-schema)|
| POST, PUT | /registry/peers/{peer}/remove/{address} | Deregister a peer by its label and IP address |
| POST      | /registry/peers/clear   | Remove all registered peers|
| GET       | /registry/peers         | Get all registered peers |
| POST      | /registry/peers/{peer}/locker/store/{key} | Store any arbitrary value for the given key in the locker of the given peer |
| POST      | /registry/peers/{peer}/locker/remove/{key} | Remove stored data for the given key from the locker of the given peer |
| GET       | /registry/peers/{peer}/locker | Get locker's data for the given peer |
| POST      | /registry/peers/{peer}/locker/clear | Clear the locker for the given peer |
| POST      | /registry/peers/lockers/clear | Clear all lockers |
| GET       | /registry/peers/targets | Get all registered targets for all peers |
| POST      | /registry/peers/{peer}/targets/add | Add a target to be sent to a peer. See [Peer Target JSON Schema](#peer-target-json-schema) |
| POST, PUT | /registry/peers/{peer}/targets/{targets}/remove | Remove given targets for a peer |
| POST      | /registry/peers/{peer}/targets/clear   | Remove all targets for a peer|
| GET       | /registry/peers/{peer}/targets   | Get all targets of a peer |
| POST, PUT | /registry/peers/{peer}/targets/{targets}/invoke | Invoke given targets on the given peer |
| POST, PUT | /registry/peers/{peer}/targets/invoke/all | Invoke all targets on the given peer |
| GET       | /registry/peers/jobs | Get all registered jobs for all peers |
| POST      | /registry/peers/{peer}/jobs/add | Add a job to be sent to a peer. See [Peer Job JSON Schema](#peer-job-json-schema) |
| POST, PUT | /registry/peers/{peer}/jobs/{jobs}/remove | Remove given jobs for a peer |
| POST      | /registry/peers/{peer}/jobs/clear   | Remove all jobs for a peer|
| GET       | /registry/peers/{peer}/jobs   | Get all jobs of a peer |
| POST, PUT | /registry/peers/{peer}/jobs/{jobs}/invoke | Invoke given jobs on the given peer |
| POST, PUT | /registry/peers/{peer}/jobs/invoke/all | Invoke all jobs on the given peer |

#### Peer JSON Schema
|Field|Data Type|Description|
|---|---|---|
| Name      | string | Name/Label of a peer |
| Namespace | string | Namespace of the peer instance (if available, else `local`) |
| Pod       | string | Pod/Hostname of the peer instance |
| Address   | string | IP address of the peer instance |

#### Peer Target JSON Schema
** Same as [Client Target JSON Schema](#client-target-json-schema)

#### Peer Job JSON Schema
** Same as [Jobs JSON Schema](#job-json-schema)

<br/>

#### Registry APIs Examples:
```
curl -X POST http://localhost:8080/registry/peers/clear

curl localhost:8080/registry/peers/add --data '
{ 
"name": "peer1",
"namespace": "test",
"pod": "podXYZ",
"address":	"1.1.1.1:8081"
}'
curl -X POST http://localhost:8080/registry/peers/peer1/remove/1.1.1.1:8081

curl localhost:8080/registry/peers

curl -X POST http://localhost:8080/registry/peers/peer1/targets/clear

curl localhost:8080/registry/peers/peer1/targets/add --data '
{ 
"name": "t1",
"method":	"POST",
"url": "http://somewhere/foo",
"headers":[["x", "x1"],["y", "y1"]],
"body": "{\"test\":\"this\"}",
"replicas": 2, 
"requestCount": 2, 
"delay": "200ms", 
"sendID": true,
"autoInvoke": true
}'

curl -X POST http://localhost:8080/registry/peers/peer1/targets/t1,t2/remove

curl http://localhost:8080/registry/peers/peer1/targets

curl -X POST http://localhost:8080/registry/peers/peer1/targets/t1,t2/invoke

curl -X POST http://localhost:8080/registry/peers/peer1/targets/invoke/all

curl -X POST http://localhost:8080/registry/peers/peer1/jobs/clear

curl localhost:8080/registry/peers/peer1/jobs/add --data '
{ 
"id": "job1",
"task": {
	"name": "job1",
	"method":	"POST",
	"url": "http://somewhere/echo",
	"headers":[["x", "x1"],["y", "y1"]],
	"body": "{\"test\":\"this\"}",
	"replicas": 1, 
  "requestCount": 1, 
	"delay": "200ms",
	"parseJSON": true
},
"auto": true,
"count": 10,
"keepFirst": true,
"maxResults": 5,
"delay": "1s"
}'

curl localhost:8080/registry/peers/peer1/jobs/add --data '
{ 
"id": "job2",
"task": {"cmd": "sh", "args": ["-c", "date +%s; echo Hello; sleep 1;"]},
"auto": true,
"count": 10,
"keepFirst": true,
"maxResults": 5,
"delay": "1s"
}'

curl -X POST http://localhost:8080/registry/peers/peer1/jobs/job1,job2/remove

curl http://localhost:8080/registry/peers/jobs

curl http://localhost:8080/registry/peers/peer1/jobs

curl -X POST http://localhost:8080/registry/peers/peer1/jobs/job1,job2/invoke

curl -X POST http://localhost:8080/registry/peers/peer1/jobs/invoke/all

```
<br/>

<details>
<summary>Registry Locker Store Example</summary>
<p>

```
    {
      "peer1": {
        "client": {
          "Data": "{\"targetInvocationCounts\":{\"t11\":400,\"t12\":400},...",
          "FirstReported": "2020-06-09T18:28:17.877231-07:00",
          "LastReported": "2020-06-09T18:28:29.955605-07:00"
        },
        "client_1": {
          "Data": "{\"targetInvocationCounts\":{\"t11\":400},\"target...",
          "FirstReported": "2020-06-09T18:28:17.879187-07:00",
          "LastReported": "2020-06-09T18:28:29.958954-07:00"
        },
        "client_2": {
          "Data": "{\"targetInvocationCounts\":{\"t12\":400}...",
          "FirstReported": "2020-06-09T18:28:17.889567-07:00",
          "LastReported": "2020-06-09T18:28:29.945121-07:00"
        },
        "job_job1_1": {
          "Data": "[{\"Index\":\"1.1\",\"Finished\":false,\"Data\":{...}]",
          "FirstReported": "2020-06-09T18:28:17.879195-07:00",
          "LastReported": "2020-06-09T18:28:27.529454-07:00"
        },
        "job_job2_2": {
          "Data": "[{\"Index\":\"2.1\",\"Finished\":false,\"Data\":\"1...}]",
          "FirstReported": "2020-06-09T18:28:18.985445-07:00",
          "LastReported": "2020-06-09T18:28:37.428542-07:00"
        }
      },
      "peer2": {
        "client": {
          "Data": "{\"targetInvocationCounts\":{\"t22\":4}...}",
          "FirstReported": "2020-06-09T18:28:19.782433-07:00",
          "LastReported": "2020-06-09T18:28:20.023149-07:00"
        },
        "client_1": {
          "Data": "{\"targetInvocationCounts\":{\"t22\":4}...}",
          "FirstReported": "2020-06-09T18:28:19.91232-07:00",
          "LastReported": "2020-06-09T18:28:20.027295-07:00"
        },
        "job_job1_1": {
          "Data": "[{\"Index\":\"1.1\",\"Finished\":false,\"ResultTime\":\"2020...\",\"Data\":\"...}]",
          "FirstReported": "2020-06-09T18:28:19.699578-07:00",
          "LastReported": "2020-06-09T18:28:22.778416-07:00"
        },
        "job_job2_2": {
          "Data": "[{\"Index\":\"2.1\",\"Finished\":false,\"ResultTime\":\"2020-0...\",\"Data\":\"...}]",
          "FirstReported": "2020-06-09T18:28:20.79828-07:00",
          "LastReported": "2020-06-09T18:28:59.698923-07:00"
        }
      }
    }
```

</p>
</details>
