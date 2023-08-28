

default:
    @just --list

killbound port:
    lsof -i :{{port}} | awk 'NR != 1 { print $2 }' | xargs kill

killall:
    @just killbound 8080
    @just killbound 8081
    @just killbound 8082