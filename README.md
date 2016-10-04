# m2http

m2http is a application that forwards messages from MQTT to an HTTP
endpoint using [rbforwarder](https://github.com/redBorder/rbforwarder)
package.

You can modify messages before send them via HTTP on the callback:

```go
mqttHandler := NewMQTTHandler(loadMQTTConfig(),
  func(client MQTT.Client, msg MQTT.Message) {
    opts := map[string]interface{}{
      "http_endpoint": msg.Topic(),
    }

    // Do something with the message
    data := msg.Payload()

    f.Produce(data, opts, nil)
  },
)
```

## Installing

To install this application ensure you have the `GOPATH` environment variable
set and **[glide](https://glide.sh/)** installed.

```bash
curl https://glide.sh/get | sh
```

And then:

1. Clone this repo and cd to the project

    ```bash
    git clone https://github.com/redBorder/m2http.git && cd m2http
    ```
2. Install dependencies and compile

    ```bash
    make
    ```
3. Install on desired directory

    ```bash
    prefix=/opt/rb make install
    ```

## Usage

```
Usage of m2http:
  --config string
        Config file
  --debug
        Show debug info
  --version
        Print version info
```

To run `m2http` just execute the following command:

```bash
m2http --config path/to/config/file
```

## Example config file

```yaml
pipeline:
  queue: 50                       # Max internal queue size
  backoff: 15                     # Time to wait between retries in seconds
  retries: 1                      # Number of retries on fail (-1 not limited)

mqtt:
  broker: "localhost"             # MQTT brokers
  port: 1883                      # MQTT broker port
  clientid: "m2kafka"             # Client ID
  topics:                         # MQTT topics to listen
    - rb_nmsp
    - rb_radius
    - rb_flow
    - rb_loc
    - rb_monitor
    - rb_state
    - rb_social

http:
  workers: 1                      # Number of workers, one connection per worker
  url: "http://localhost:8888"    # URL of the HTTP endpoint
  insecure: false                 # Skip SSSL verification
```
