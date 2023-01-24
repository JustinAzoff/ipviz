# ipviz
Visualize zeek conn logs using a hilbert space filling curve

## Quickstart

    $ go build && ./ipviz &
    $ tail -F conn.log | nc localhost 9999

![ipviz screenshot](screenshot.gif)
