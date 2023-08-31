

default:
    @just --list

killbound port:
    lsof -i :{{port}} | awk 'NR != 1 { print $2 }' | xargs kill

# Careful, kill the ports that may have been left over from a bad run (8080, 8081, 8082)
killall:
    @just killbound 8080
    @just killbound 8081
    @just killbound 8082

# built the tunny application
build:
    go build -o bin/tunny ./cmd/main